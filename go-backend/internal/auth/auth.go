package auth

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/bridge"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/msteinert/pam"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const sessionDuration = 6 * time.Hour

func RegisterAuthRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", loginHandler)
		auth.GET("/me", AuthMiddleware(), meHandler)
		auth.GET("/logout", AuthMiddleware(), logoutHandler)
	}
}

func pamAuth(username, password string) error {
	t, err := pam.StartFunc("login", username, func(s pam.Style, msg string) (string, error) {
		return password, nil
	})
	if err != nil {
		return err
	}
	return t.Authenticate(0)
}

func trySudo(password string) bool {
	// Try to run 'sudo -S -l' to check if we can get privileged access
	cmd := exec.Command("sudo", "-S", "-l")
	cmd.Env = append(cmd.Env, "LANG=C") // force English output for easier parsing

	// Sudo will read password from stdin
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, password+"\n")
	}()
	err = cmd.Run()
	// If sudo exits 0 and "may run" is in the output, we have sudo
	return err == nil && (bytes.Contains(out.Bytes(), []byte("may run")) || bytes.Contains(stderr.Bytes(), []byte("may run")))
}

func loginHandler(c *gin.Context) {

	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 1. Authenticate with PAM
	if err := pamAuth(req.Username, req.Password); err != nil {
		logger.Warnf("❌ Authentication failed for user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	// 2. Check if user has sudo rights
	privileged := trySudo(req.Password)

	// 3. Create session (with privilege info)
	sessionID := uuid.New().String()
	user := utils.User{ID: req.Username, Name: req.Username}
	session.CreateSession(sessionID, user, sessionDuration, privileged)
	sess := session.Get(sessionID)

	// 4. Start main socket for this session
	if sess == nil {
		logger.Errorf("Failed to get session after creation (id=%s)", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
		return
	}

	err := bridge.StartBridgeSocket(sess)
	if err != nil {
		logger.Errorf("Failed to start main socket: %v", err)
		session.DeleteSession(sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session socket"})
		return
	}

	// 5. Start the bridge process for this session
	if err := bridge.StartBridge(sess, req.Password); err != nil {
		if privileged {
			logger.Warnf("Privileged bridge failed, falling back to unprivileged: %v", err)
			privileged = false
			if err2 := bridge.StartBridge(sess, req.Password); err2 != nil {
				logger.Errorf("Unprivileged bridge also failed: %v", err2)
				session.DeleteSession(sessionID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start bridge"})
				return
			}
		} else {
			logger.Errorf("Bridge failed to start: %v", err)
			session.DeleteSession(sessionID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start bridge"})
			return
		}
	}

	// 6. Set session cookie
	env := os.Getenv("GO_ENV")
	isHTTPS := c.Request.TLS != nil
	secureCookie := env == "production" && isHTTPS

	c.SetCookie("session_id", sessionID, int(sessionDuration.Seconds()), "/", "", secureCookie, true)

	// 7. Send response
	c.JSON(http.StatusOK, gin.H{"success": true, "privileged": privileged})
}

func logoutHandler(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.Status(http.StatusOK)
		return
	}

	s := session.Get(sessionID)
	if s == nil {
		logger.Debugf("[auth] No session found for ID: %s (already expired?)", sessionID)
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.Status(http.StatusOK)
		return
	}

	session.DeleteSession(sessionID)
	if s.User.ID != "" {
		bridge.CallWithSession(s, "control", "shutdown", nil)
		bridge.CleanupBridgeSocket(s)
	}
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	logger.Infof("👋 Logged out session: %s", sessionID)
	c.Status(http.StatusOK)
}

func meHandler(c *gin.Context) {
	user := c.MustGet("user").(utils.User)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

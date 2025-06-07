package auth

import (
	"bytes"
	"go-backend/internal/bridge"
	"go-backend/internal/logger"
	"go-backend/internal/session"
	"go-backend/internal/utils"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

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

	// 4. Start main socket for this session
	err := bridge.StartBridgeSocket(sessionID, req.Username)
	if err != nil {
		logger.Errorf("Failed to start main socket: %v", err)
		session.DeleteSession(sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session socket"})
		return
	}

	// 5. Start the bridge process for this session
	var bridgeErr error
	if err := bridge.StartBridge(sessionID, req.Username, privileged, req.Password); err != nil {
		// If privileged failed, try normal (only fallback if you want)
		if privileged {
			logger.Warnf("Privileged bridge failed, falling back to unprivileged: %v", err)
			privileged = false
			bridgeErr = bridge.StartBridge(sessionID, req.Username, privileged, req.Password)
		} else {
			bridgeErr = err
		}
	}
	if bridgeErr != nil {
		logger.Errorf("Failed to start bridge for session %s: %v", sessionID, bridgeErr)
		session.DeleteSession(sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start backend bridge"})
		return
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
	if err == nil {
		s := session.Get(sessionID) // Safe concurrent read
		var username string
		if s == nil {
			logger.Debugf("[auth] No session found for ID: %s (already expired?)", sessionID)
		}
		if s != nil {
			username = s.User.ID
		}
		session.DeleteSession(sessionID)
		if username != "" {
			bridge.CallWithSession(sessionID, username, "control", "shutdown", nil)
			bridge.CleanupBridgeSocket(sessionID, username)
		}
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		logger.Infof("👋 Logged out session: %s", sessionID)
	}
	c.Status(http.StatusOK)
}

func meHandler(c *gin.Context) {
	user := c.MustGet("user").(utils.User)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

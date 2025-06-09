package cleanup

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mordilloSan/LinuxIO/go-backend/internal/bridge"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/logger"
	"github.com/mordilloSan/LinuxIO/go-backend/internal/session"
)

type BridgeHealthRequest struct {
	Type    string `json:"type"`    // e.g., "healthcheck" or "validate"
	Session string `json:"session"` // sessionID
}
type BridgeHealthResponse struct {
	Status  string `json:"status"` // "ok" or "invalid"
	Message string `json:"message,omitempty"`
}

func KillLingeringBridgeStartupProcesses() {
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		logger.Errorf("❌ Failed to read /proc: %v", err)
		return
	}

	for _, entry := range procEntries {
		if !entry.IsDir() || !IsNumeric(entry.Name()) {
			continue
		}

		pid := entry.Name()
		cmdlineBytes, err := os.ReadFile(fmt.Sprintf("/proc/%s/cmdline", pid))
		if err != nil || len(cmdlineBytes) == 0 {
			continue
		}

		cmdline := strings.ReplaceAll(string(cmdlineBytes), "\x00", " ")

		if strings.Contains(cmdline, "linuxio-bridge") &&
			strings.Contains(cmdline, "sudo -S env") &&
			strings.Contains(cmdline, "LINUXIO_SESSION_USER="+os.Getenv("LINUXIO_SESSION_USER")) {
			pidInt, _ := strconv.Atoi(pid)
			logger.Debugf("⚠️ Found lingering bridge process (pid=%d): %s", pidInt, cmdline)
			killParentTree(pidInt)
		}
	}
}

func killParentTree(pid int) {
	for {
		stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			logger.Debugf("killParentTree: could not read stat for pid %d: %v", pid, err)
			break
		}
		fields := strings.Fields(string(stat))
		if len(fields) < 4 {
			logger.Debugf("killParentTree: stat fields < 4 for pid %d", pid)
			break
		}

		ppid, _ := strconv.Atoi(fields[3])
		if ppid <= 1 || ppid == pid {
			logger.Debugf("killParentTree: hit root or self for pid %d (ppid %d)", pid, ppid)
			break
		}

		commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid))
		if err != nil {
			logger.Debugf("killParentTree: could not read comm for ppid %d: %v", ppid, err)
			break
		}

		comm := strings.TrimSpace(string(commBytes))
		logger.Debugf("killParentTree: pid=%d, ppid=%d, comm='%s'", pid, ppid, comm)
		if comm == "sudo" || comm == "env" {
			logger.Debugf("🛑 Killing sudo/env process (pid=%d, ppid=%d, comm=%s)", pid, ppid, comm)
			_ = syscall.Kill(ppid, syscall.SIGTERM)
			_ = syscall.Kill(pid, syscall.SIGTERM)
			time.Sleep(250 * time.Millisecond) // Give time for defer/logs to flush
			_ = syscall.Kill(ppid, syscall.SIGKILL)
			_ = syscall.Kill(pid, syscall.SIGKILL)
			break
		}
		pid = ppid
	}
}

func CheckMainProcessHealth(sess *session.Session) bool {
	sock := bridge.MainSocketPath(sess)
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		logger.Warnf("⚠️ Could not connect to main socket: %v", err)
		return false
	}
	defer conn.Close()

	req := BridgeHealthRequest{
		Type:    "validate",
		Session: sess.SessionID,
	}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		logger.Warnf("⚠️ Failed to send health request: %v", err)
		return false
	}

	var resp BridgeHealthResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		logger.Warnf("⚠️ Failed to decode health response: %v", err)
		return false
	}
	logger.Debugf("Healthcheck: got %+v", resp)
	return resp.Status == "ok"
}

func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

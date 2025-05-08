// go-backend/update/changelog.go
package update

import (
	"go-backend/internal/logger"
	"go-backend/internal/utils"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func getChangelogHandler(c *gin.Context) {
	packageName := c.Query("package")
	if strings.TrimSpace(packageName) == "" {
		logger.Warning.Println("[update] Missing package query parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing package query parameter"})
		return
	}

	distro, _ := utils.GetDistroID()
	logger.Info.Printf("[update] Fetching changelog for '%s' on distro '%s'\n", packageName, distro)
	changelog := getChangelog(distro, packageName)

	c.JSON(http.StatusOK, gin.H{
		"package":   packageName,
		"changelog": changelog,
	})
}

func getChangelog(distro string, packageName string) string {
	if strings.TrimSpace(packageName) == "" {
		return "Changelog not available"
	}

	ids := strings.Fields(strings.ToLower(distro))
	var cmd *exec.Cmd
	var useAptParser bool

	switch {
	case containsAny(ids, "debian", "ubuntu"):
		cmd = exec.Command("apt", "changelog", packageName)
		useAptParser = true
	case containsAny(ids, "rhel", "fedora", "centos", "rocky", "almalinux"):
		cmd = exec.Command("dnf", "changelog", "info", packageName)
	default:
		logger.Warning.Printf("[update] Unsupported distro for changelog: %s\n", distro)
		return "Changelog not available"
	}

	output, err := cmd.CombinedOutput()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		logger.Warning.Printf("[update] No changelog found or command failed: %v\n", err)
		return "Changelog not available"
	}

	lines := strings.Split(string(output), "\n")

	if useAptParser {
		return parseAptChangelog(lines)
	}

	return string(output)
}

func containsAny(slice []string, values ...string) bool {
	for _, s := range slice {
		for _, v := range values {
			if s == v {
				return true
			}
		}
	}
	return false
}

func parseAptChangelog(lines []string) string {
	var buf strings.Builder
	skipHeaders := []string{
		"WARNING: apt does not have a stable CLI interface",
		"Get:",
	}

	lineCount := 0
	started := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip known headers/warnings
		skip := false
		for _, prefix := range skipHeaders {
			if strings.HasPrefix(trimmed, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Skip the package header line like: "package (1.0.0-1) distro; urgency=medium"
		if !started && strings.Contains(trimmed, "urgency=") && strings.HasSuffix(trimmed, "medium") {
			continue
		}

		// Stop after the maintainer signature block
		if strings.HasPrefix(trimmed, "-- ") {
			break
		}

		started = true
		buf.WriteString(trimmed + "\n")
		lineCount++

		if lineCount >= 30 {
			break
		}
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		return "Changelog not available"
	}
	return output
}

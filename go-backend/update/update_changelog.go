package update

import (
	"os/exec"
	"strings"
)

func getChangelog(distro string, packageName string) string {
	if strings.TrimSpace(packageName) == "" {
		return "Changelog not available"
	}

	ids := strings.Fields(strings.ToLower(distro))
	var cmd *exec.Cmd

	switch {
	case containsAny(ids, "debian", "ubuntu"):
		cmd = exec.Command("apt", "changelog", packageName)
	case containsAny(ids, "rhel", "fedora", "centos", "rocky", "almalinux"):
		cmd = exec.Command("dnf", "changelog", "info", packageName)
	default:
		return "Changelog not available"
	}

	output, err := cmd.CombinedOutput()
	if err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return "Changelog not available"
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

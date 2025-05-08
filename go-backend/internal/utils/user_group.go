// utils/user_group.go
package utils

import (
	"os/exec"
	"slices"
	"strings"
)

// IsUserInGroup checks if the given user is in a specified group.
func IsUserInGroup(username, group string) (bool, error) {
	out, err := exec.Command("id", "-nG", username).Output()
	if err != nil {
		return false, err
	}

	groups := strings.Fields(string(out))
	if slices.Contains(groups, group) {
			return true, nil
		}
	return false, nil
}

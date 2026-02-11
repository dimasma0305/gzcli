package gzcli

import (
	"fmt"
	"strings"
)

// generateUsername keeps the requested username and only modifies it if it collides.
func generateUsername(realName string, maxLength int, existingUsernames map[string]struct{}) (string, error) {
	baseUsername := strings.TrimSpace(realName)
	if baseUsername == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	if len(baseUsername) > maxLength {
		baseUsername = baseUsername[:maxLength]
	}

	// Ensure uniqueness
	username := baseUsername
	for i := 1; ; i++ {
		if _, exists := existingUsernames[username]; !exists {
			existingUsernames[username] = struct{}{}
			return username, nil
		}

		suffix := fmt.Sprint(i)
		if newLen := len(baseUsername) + len(suffix); newLen <= maxLength {
			username = baseUsername + suffix
		} else {
			username = baseUsername[:maxLength-len(suffix)] + suffix
		}
	}
}

package gzcli

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

// LeetSpeakMap defines common rune replacements for leetspeak transformations.
var LeetSpeakMap = map[rune]rune{
	'a': '4',
	'e': '3',
	'i': '1',
	'o': '0',
	's': '5',
	't': '7',
	'g': '9',
}

// transformRandomly applies a series of random transformations to a string to generate a username.
// These transformations include converting spaces to underscores, applying leetspeak replacements,
// and randomly changing the case of letters.
func transformRandomly(s string) string {
	//nolint:gosec // G404: Weak RNG acceptable for non-security username generation
	localRand := rand.New(rand.NewSource(rand.Int63())) // Local generator seeded from global source
	var transformed strings.Builder
	transformed.Grow(len(s)) // Pre-allocate capacity

	for _, r := range s {
		switch r {
		case ' ':
			transformed.WriteByte('_')
		default:
			// Leetspeak replacement with 50% probability
			if replacement, exists := LeetSpeakMap[r]; exists && localRand.Intn(2) == 0 {
				transformed.WriteRune(replacement)
			} else {
				// Random case transformation
				if localRand.Intn(2) == 0 {
					r = unicode.ToUpper(r)
				} else {
					r = unicode.ToLower(r)
				}
				transformed.WriteRune(r)
			}
		}
	}
	return transformed.String()
}

// generateUsername creates a unique username based on a real name, with a specified maximum length.
// It uses a set of existing usernames to ensure the generated username is unique.
// The function cleans the real name, applies random transformations, and appends a numerical suffix
// if necessary to ensure uniqueness.
func generateUsername(realName string, maxLength int, existingUsernames map[string]struct{}) (string, error) {
	// Clean and normalize base username
	var baseBuilder strings.Builder
	for _, r := range strings.ToLower(realName) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			baseBuilder.WriteRune(r)
		}
	}
	baseUsername := baseBuilder.String()

	// Apply transformations and truncate
	transformed := transformRandomly(baseUsername)
	if len(transformed) > maxLength {
		transformed = transformed[:maxLength]
	}

	// Ensure uniqueness
	username := transformed
	for i := 1; ; i++ {
		if _, exists := existingUsernames[username]; !exists {
			existingUsernames[username] = struct{}{}
			return username, nil
		}

		suffix := fmt.Sprint(i)
		if newLen := len(transformed) + len(suffix); newLen <= maxLength {
			username = transformed + suffix
		} else {
			username = transformed[:maxLength-len(suffix)] + suffix
		}
	}
}

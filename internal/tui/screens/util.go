package screens

import "strings"

// safeRepeat wraps strings.Repeat to gracefully handle edge cases
// from dynamic terminal dimensions approaching zero.
func safeRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

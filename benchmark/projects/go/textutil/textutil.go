package textutil

import "strings"

// Truncate cuts s to n runes and adds "..." when it cuts.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

// Words counts the words in s.
func Words(s string) int {
	return len(strings.Fields(s))
}

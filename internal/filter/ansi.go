package filter

import "regexp"

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StripANSI(b []byte) []byte {
	return ansiRegex.ReplaceAll(b, nil)
}

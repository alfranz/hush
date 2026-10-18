package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse reads KEY=VALUE lines. Empty lines and lines starting with # are skipped.
func Parse(text string) (map[string]string, error) {
	out := map[string]string{}
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: missing =", i+1)
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return out, nil
}

// Int returns the value of key as an int, or def if it is missing.
func Int(m map[string]string, key string, def int) (int, error) {
	v, ok := m[key]
	if !ok {
		return def, nil
	}
	return strconv.Atoi(v)
}

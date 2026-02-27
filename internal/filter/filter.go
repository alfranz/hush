package filter

import (
	"bytes"
	"regexp"
)

type Options struct {
	Head      int
	Tail      int
	Grep      string
	StripANSI bool
	AgentMode bool
}

func Apply(raw []byte, opts Options) []byte {
	result := raw

	// 1. ANSI stripping (agent mode implies this)
	if opts.StripANSI || opts.AgentMode {
		result = StripANSI(result)
	}

	// 2. Agent mode transformations
	if opts.AgentMode {
		result = ApplyAgentMode(result)
	}

	// 3. Grep filter
	if opts.Grep != "" {
		result = applyGrep(result, opts.Grep)
	}

	// 4. Head/Tail
	if opts.Head > 0 || opts.Tail > 0 {
		result = applyHeadTail(result, opts.Head, opts.Tail)
	}

	return result
}

func applyGrep(b []byte, pattern string) []byte {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return b // invalid pattern, return unfiltered
	}
	lines := bytes.Split(b, []byte("\n"))
	var matched [][]byte
	for _, line := range lines {
		if re.Match(line) {
			matched = append(matched, line)
		}
	}
	return bytes.Join(matched, []byte("\n"))
}

func applyHeadTail(b []byte, head, tail int) []byte {
	lines := bytes.Split(b, []byte("\n"))
	// Remove trailing empty line from split
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	if tail > 0 && tail < len(lines) {
		lines = lines[len(lines)-tail:]
	}
	if head > 0 && head < len(lines) {
		lines = lines[:head]
	}
	return bytes.Join(lines, []byte("\n"))
}

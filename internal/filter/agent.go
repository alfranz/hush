package filter

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches ISO 8601 timestamps at start of line: 2024-01-15T10:30:00 or 2024-01-15 10:30:00.123
	timestampRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[.\d]*\s*`)
	// Matches tqdm-style progress: 50%|████ | 5/10
	progressRegex = regexp.MustCompile(`\d+%\|[█▏▎▍▌▋▊▉ ]*\|`)
)

func ApplyAgentMode(b []byte) []byte {
	b = StripANSI(b)
	b = collapseTracebacks(b)
	b = removeProgressLines(b)
	b = removeTimestamps(b)
	return b
}

func collapseTracebacks(b []byte) []byte {
	lines := bytes.Split(b, []byte("\n"))
	var result [][]byte
	i := 0
	for i < len(lines) {
		line := string(lines[i])
		if strings.TrimSpace(line) == "Traceback (most recent call last):" {
			// Count frames and find the exception line
			frameCount := 0
			j := i + 1
			for j < len(lines) {
				trimmed := strings.TrimSpace(string(lines[j]))
				if strings.HasPrefix(trimmed, "File ") {
					frameCount++
					j++ // skip the code line
				} else if trimmed == "" || strings.HasPrefix(trimmed, "File ") {
					// continue
				} else {
					break
				}
				j++
			}
			if j < len(lines) {
				exceptionLine := strings.TrimSpace(string(lines[j]))
				summary := fmt.Sprintf("Traceback ... (%d frames) → %s", frameCount, exceptionLine)
				result = append(result, []byte(summary))
				i = j + 1
			} else {
				result = append(result, lines[i])
				i++
			}
		} else {
			result = append(result, lines[i])
			i++
		}
	}
	return bytes.Join(result, []byte("\n"))
}

func removeProgressLines(b []byte) []byte {
	lines := bytes.Split(b, []byte("\n"))
	var result [][]byte
	for _, line := range lines {
		s := string(line)
		// Skip lines with carriage returns (progress bar overwrites)
		if strings.Contains(s, "\r") {
			continue
		}
		// Skip tqdm-style progress
		if progressRegex.Match(line) {
			continue
		}
		result = append(result, line)
	}
	return bytes.Join(result, []byte("\n"))
}

func removeTimestamps(b []byte) []byte {
	lines := bytes.Split(b, []byte("\n"))
	var result [][]byte
	for _, line := range lines {
		cleaned := timestampRegex.ReplaceAll(line, nil)
		result = append(result, cleaned)
	}
	return bytes.Join(result, []byte("\n"))
}

package cli

import "github.com/alfranz/hush/internal/filter"

type warningReport struct {
	count int
	lines []byte
}

func buildWarningReport(raw []byte, f sharedFlags) warningReport {
	if f.warnPattern == "" {
		return warningReport{}
	}

	cleaned := filter.Apply(raw, filter.Options{StripANSI: true})
	matches := filter.MatchLines(cleaned, f.warnPattern)
	if matches.Count == 0 {
		return warningReport{}
	}

	lines := filter.Apply(matches.Lines, filter.Options{
		Head: f.head,
		Tail: f.tail,
		Grep: f.grep,
	})

	warnTail := f.warnTail
	if warnTail <= 0 {
		warnTail = 10
	}
	lines = filter.Apply(lines, filter.Options{Tail: warnTail})

	return warningReport{
		count: matches.Count,
		lines: lines,
	}
}

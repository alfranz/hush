package cli

import (
	"context"
	"os"

	"github.com/alfranz/hush/internal/config"
	"github.com/alfranz/hush/internal/filter"
	"github.com/alfranz/hush/internal/output"
	"github.com/alfranz/hush/internal/runner"
	"github.com/spf13/cobra"
)

type sharedFlags struct {
	label       string
	tail        int
	head        int
	grep        string
	warnPattern string
	warnTail    int
}

func addSharedFlags(cmd *cobra.Command, f *sharedFlags) {
	cmd.PersistentFlags().StringVar(&f.label, "label", "", "Custom label for the summary line")
	cmd.PersistentFlags().IntVar(&f.tail, "tail", 0, "Show only last N lines of output on failure")
	cmd.PersistentFlags().IntVar(&f.head, "head", 0, "Show only first N lines of output on failure")
	cmd.PersistentFlags().StringVar(&f.grep, "grep", "", "Filter output to lines matching this regex")
	cmd.PersistentFlags().StringVar(&f.warnPattern, "warn-pattern", "", "On success, treat matching output lines as warnings")
	cmd.PersistentFlags().IntVar(&f.warnTail, "warn-tail", 0, "On warning-qualified success, show last N warning lines (default 10)")
}

// checkFlags returns the settings for one named check.
// Precedence: CLI flags > per-check config > defaults.
func checkFlags(check config.Check, d config.Defaults, cli sharedFlags) sharedFlags {
	f := sharedFlags{
		label:       check.Label,
		head:        d.Head,
		tail:        d.Tail,
		grep:        d.Grep,
		warnPattern: d.WarnPattern,
		warnTail:    d.WarnTail,
	}
	f.head = firstSet(cli.head, check.Head, f.head)
	f.tail = firstSet(cli.tail, check.Tail, f.tail)
	f.grep = firstSet(cli.grep, check.Grep, f.grep)
	f.warnPattern = firstSet(cli.warnPattern, check.WarnPattern, f.warnPattern)
	f.warnTail = firstSet(cli.warnTail, check.WarnTail, f.warnTail)
	return f
}

// firstSet returns the first value that is not the zero value.
func firstSet[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}

// job is one command and the settings to run it with.
type job struct {
	run   runner.Options
	flags sharedFlags
	// indent goes before each output line: none for a single command,
	// two spaces in batch.
	indent string
}

// runJob runs one command, prints its summary line and returns its exit code.
func runJob(j job) (int, error) {
	result, err := runner.Run(context.Background(), j.run)
	if err != nil {
		return 0, err
	}

	filtered := filter.Apply(result.Output, filter.Options{
		Head:      j.flags.head,
		Tail:      j.flags.tail,
		Grep:      j.flags.grep,
		StripANSI: true,
	})
	warnings := buildWarningReport(result.Output, j.flags)

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, filtered, warnings.count, warnings.lines, j.indent)
	return result.ExitCode, nil
}

package cli

import "github.com/spf13/cobra"

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

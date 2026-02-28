package cli

import "github.com/spf13/cobra"

type sharedFlags struct {
	label string
	tail  int
	head  int
	grep  string
}

func addSharedFlags(cmd *cobra.Command, f *sharedFlags) {
	cmd.PersistentFlags().StringVar(&f.label, "label", "", "Custom label for the summary line")
	cmd.PersistentFlags().IntVar(&f.tail, "tail", 0, "Show only last N lines of output on failure")
	cmd.PersistentFlags().IntVar(&f.head, "head", 0, "Show only first N lines of output on failure")
	cmd.PersistentFlags().StringVar(&f.grep, "grep", "", "Filter output to lines matching this regex")
}

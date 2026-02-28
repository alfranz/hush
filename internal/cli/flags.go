package cli

import "github.com/spf13/cobra"

type sharedFlags struct {
	label   string
	tail    int
	head    int
	grep    string
	noTime  bool
	color   bool
	noColor bool
}

func addSharedFlags(cmd *cobra.Command, f *sharedFlags) {
	cmd.PersistentFlags().StringVar(&f.label, "label", "", "Custom label for the summary line")
	cmd.PersistentFlags().IntVar(&f.tail, "tail", 0, "Show only last N lines of output on failure")
	cmd.PersistentFlags().IntVar(&f.head, "head", 0, "Show only first N lines of output on failure")
	cmd.PersistentFlags().StringVar(&f.grep, "grep", "", "Filter output to lines matching this regex")
	cmd.PersistentFlags().BoolVar(&f.noTime, "no-time", false, "Suppress duration in output")
	cmd.PersistentFlags().BoolVar(&f.color, "color", false, "Force colored output")
	cmd.PersistentFlags().BoolVar(&f.noColor, "no-color", false, "Disable colored output")
}

func (f *sharedFlags) colorOpt() *bool {
	if f.color {
		return ptrTo(true)
	}
	if f.noColor {
		return ptrTo(false)
	}
	return nil
}

func ptrTo[T any](v T) *T {
	return &v
}

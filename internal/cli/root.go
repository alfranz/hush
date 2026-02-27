package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/alexfranz/hush-cli/internal/filter"
	"github.com/alexfranz/hush-cli/internal/output"
	"github.com/alexfranz/hush-cli/internal/runner"
	"github.com/spf13/cobra"
)

var flags sharedFlags

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hush [flags] <command>",
		Short: "Context-efficient command runner",
		Long:  "Wraps any shell command and prints a single ✓/✗ summary line.\nOn success, shows only the summary. On failure, shows filtered output.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runRoot,
		// Don't show errors twice
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	addSharedFlags(cmd, &flags)
	cmd.AddCommand(newBatchCmd())

	return cmd
}

func Execute() {
	cmd := NewRootCmd()

	// Load config and register named check subcommands
	registerNamedChecks(cmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	command := args[0]

	result, err := runner.Run(context.Background(), runner.Options{
		Command: command,
		Label:   flags.label,
	})
	if err != nil {
		return err
	}

	filtered := filter.Apply(result.Output, filter.Options{
		Head:      flags.head,
		Tail:      flags.tail,
		Grep:      flags.grep,
		StripANSI: flags.noColor,
		AgentMode: flags.agent,
	})

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, result.Duration, filtered, output.Options{
		NoTime: flags.noTime,
		Color:  flags.colorOpt(),
	})

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

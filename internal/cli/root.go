package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/alfranz/hush/internal/filter"
	"github.com/alfranz/hush/internal/output"
	"github.com/alfranz/hush/internal/runner"
	"github.com/spf13/cobra"
)

var flags sharedFlags

const logo = `
  _               _
 | |__  _   _ ___| |__
 | '_ \| | | / __| '_ \
 | | | | |_| \__ \ | | |
 |_| |_|\__,_|___/_| |_|
`

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hush [flags] <command>",
		Short: "Context-efficient command runner",
		Long:  logo + "Run commands. Save tokens.\n✓ on success. Filtered output on failure. Built for agents.",
		Args:  cobra.ArbitraryArgs,
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
	if len(args) == 0 {
		return cmd.Help()
	}
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

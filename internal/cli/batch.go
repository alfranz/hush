package cli

import (
	"context"
	"os"
	"time"

	"github.com/alfranz/hush/internal/filter"
	"github.com/alfranz/hush/internal/output"
	"github.com/alfranz/hush/internal/runner"
	"github.com/spf13/cobra"
)

var batchFlags struct {
	sharedFlags
	continueOnError bool
}

func newBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch [flags] <command1> <command2> ...",
		Short: "Run multiple commands sequentially",
		Long:  "Runs commands sequentially and shows a summary.\nStops on first failure by default; use --continue to run all.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runBatch,
		// Don't show errors twice
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	addSharedFlags(cmd, &batchFlags.sharedFlags)
	cmd.Flags().BoolVar(&batchFlags.continueOnError, "continue", false, "Continue running after a failure")
	return cmd
}

func runBatch(cmd *cobra.Command, args []string) error {
	return executeBatch(args, batchFlags.sharedFlags, batchFlags.continueOnError)
}

func executeBatch(commands []string, f sharedFlags, continueOnError bool) error {
	start := time.Now()
	passed := 0
	total := len(commands)
	firstFailCode := 0

	for _, command := range commands {
		result, err := runner.Run(context.Background(), runner.Options{
			Command: command,
		})
		if err != nil {
			return err
		}

		filtered := filter.Apply(result.Output, filter.Options{
			Head:      f.head,
			Tail:      f.tail,
			Grep:      f.grep,
			StripANSI: f.noColor,
		})

		output.PrintResult(os.Stdout, result.Label, result.ExitCode, result.Duration, filtered, output.Options{
			NoTime: f.noTime,
			Color:  f.colorOpt(),
		})

		if result.ExitCode == 0 {
			passed++
		} else {
			if firstFailCode == 0 {
				firstFailCode = result.ExitCode
			}
			if !continueOnError {
				break
			}
		}
	}

	duration := time.Since(start)

	// Print summary if all passed or we ran all commands
	if passed == total || continueOnError {
		output.PrintBatchSummary(os.Stdout, passed, total, duration, f.noTime, f.colorOpt())
	}

	if firstFailCode != 0 {
		os.Exit(firstFailCode)
	}
	return nil
}

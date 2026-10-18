package cli

import (
	"fmt"
	"os"

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
	_ = cmd.PersistentFlags().MarkHidden("label")
	cmd.Flags().BoolVar(&batchFlags.continueOnError, "continue", false, "Continue running after a failure")
	return cmd
}

func runBatch(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("label") {
		return fmt.Errorf("--label does not work with batch: each command is labeled with its own text.\n" +
			"To name checks, define them in .hush.yaml and run hush all")
	}
	cfg := loadConfigQuiet()
	f := applyDefaults(cmd, batchFlags.sharedFlags, cfg)

	continueOnError := batchFlags.continueOnError
	if !cmd.Flags().Changed("continue") && cfg != nil && cfg.Defaults.Continue {
		continueOnError = true
	}

	jobs := make([]job, len(args))
	for i, command := range args {
		jobs[i] = job{run: runner.Options{Command: command}, flags: f}
	}
	return executeBatch(jobs, continueOnError)
}

func executeBatch(jobs []job, continueOnError bool) error {
	passed := 0
	total := len(jobs)
	firstFailCode := 0

	for _, j := range jobs {
		exitCode, err := runJob(j)
		if err != nil {
			return err
		}

		if exitCode == 0 {
			passed++
		} else {
			if firstFailCode == 0 {
				firstFailCode = exitCode
			}
			if !continueOnError {
				break
			}
		}
	}

	// Print summary if all passed or we ran all commands
	if passed == total || continueOnError {
		output.PrintBatchSummary(os.Stdout, passed, total)
	}

	if firstFailCode != 0 {
		os.Exit(firstFailCode)
	}
	return nil
}

package cli

import (
	"os"

	"github.com/alfranz/hush/internal/config"
	"github.com/alfranz/hush/internal/runner"
	"github.com/spf13/cobra"
)

func registerNamedChecks(root *cobra.Command) {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return
	}

	// Register individual check commands
	for name, check := range cfg.Checks {
		check := check
		cmd := &cobra.Command{
			Use:           name,
			Short:         "Run " + name + " check from .hush.yaml",
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runNamedCheck(check, cfg.Defaults)
			},
		}
		root.AddCommand(cmd)
	}

	// Register "all" command
	allCmd := &cobra.Command{
		Use:           "all",
		Short:         "Run all checks from .hush.yaml",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var jobs []job
			for _, name := range cfg.CheckOrder {
				jobs = append(jobs, checkJob(cfg.Checks[name], cfg.Defaults))
			}
			// Use defaults.continue for "all" command
			continueOnError := cfg.Defaults.Continue
			if cmd.Flags().Changed("continue") {
				continueOnError, _ = cmd.Flags().GetBool("continue")
			}
			return executeBatch(jobs, continueOnError)
		},
	}
	allCmd.Flags().BoolP("continue", "", false, "Continue running after a failure")
	root.AddCommand(allCmd)
}

// checkJob returns the job for one named check, with the CLI flags applied.
func checkJob(check config.Check, defaults config.Defaults) job {
	f := checkFlags(check, defaults, flags)
	return job{
		run:   runner.Options{Command: check.Cmd, Label: f.label},
		flags: f,
	}
}

func runNamedCheck(check config.Check, defaults config.Defaults) error {
	exitCode, err := runJob(checkJob(check, defaults))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

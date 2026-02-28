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

func registerNamedChecks(root *cobra.Command) {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return
	}

	// Collect check names in order for "all" command
	var checkNames []string
	for name := range cfg.Checks {
		checkNames = append(checkNames, name)
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
				return runNamedCheck(check, cfg)
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
			var commands []string
			for _, name := range checkNames {
				commands = append(commands, cfg.Checks[name].Cmd)
			}
			// Use defaults.continue for "all" command
			continueOnError := cfg.Defaults.Continue
			if cmd.Flags().Changed("continue") {
				continueOnError, _ = cmd.Flags().GetBool("continue")
			}
			return executeBatch(commands, flags, continueOnError)
		},
	}
	allCmd.Flags().BoolP("continue", "", false, "Continue running after a failure")
	root.AddCommand(allCmd)
}

func runNamedCheck(check config.Check, cfg *config.Config) error {
	result, err := runner.Run(context.Background(), runner.Options{
		Command: check.Cmd,
		Label:   check.Label,
	})
	if err != nil {
		return err
	}

	// Precedence: CLI flags > per-check config > defaults > zero
	filterOpts := filter.Options{
		StripANSI: true,
	}

	// Start with defaults
	if cfg != nil {
		filterOpts.Head = cfg.Defaults.Head
		filterOpts.Tail = cfg.Defaults.Tail
		filterOpts.Grep = cfg.Defaults.Grep
	}

	// Per-check config overrides defaults
	if check.Head > 0 {
		filterOpts.Head = check.Head
	}
	if check.Tail > 0 {
		filterOpts.Tail = check.Tail
	}
	if check.Grep != "" {
		filterOpts.Grep = check.Grep
	}

	// CLI flags override everything
	if flags.head > 0 {
		filterOpts.Head = flags.head
	}
	if flags.tail > 0 {
		filterOpts.Tail = flags.tail
	}
	if flags.grep != "" {
		filterOpts.Grep = flags.grep
	}

	filtered := filter.Apply(result.Output, filterOpts)

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, filtered)

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

package cli

import (
	"context"
	"os"

	"github.com/alexfranz/hush-cli/internal/config"
	"github.com/alexfranz/hush-cli/internal/filter"
	"github.com/alexfranz/hush-cli/internal/output"
	"github.com/alexfranz/hush-cli/internal/runner"
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
		check := check // capture
		name := name
		cmd := &cobra.Command{
			Use:           name,
			Short:         "Run " + name + " check from .hush.yaml",
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runNamedCheck(check)
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
			return executeBatch(commands, flags, false)
		},
	}
	allCmd.Flags().BoolP("continue", "", false, "Continue running after a failure")
	root.AddCommand(allCmd)
}

func runNamedCheck(check config.Check) error {
	label := check.Label
	if label == "" {
		// Will be derived from command by runner
		label = ""
	}

	result, err := runner.Run(context.Background(), runner.Options{
		Command: check.Cmd,
		Label:   label,
	})
	if err != nil {
		return err
	}

	// Merge flags: check config takes precedence for filter options, CLI flags for output options
	filterOpts := filter.Options{
		Head:      check.Head,
		Tail:      check.Tail,
		Grep:      check.Grep,
		AgentMode: check.Agent,
	}

	// CLI flags can override/augment
	if flags.head > 0 {
		filterOpts.Head = flags.head
	}
	if flags.tail > 0 {
		filterOpts.Tail = flags.tail
	}
	if flags.grep != "" {
		filterOpts.Grep = flags.grep
	}
	if flags.agent {
		filterOpts.AgentMode = true
	}
	if flags.noColor {
		filterOpts.StripANSI = true
	}

	filtered := filter.Apply(result.Output, filterOpts)

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, result.Duration, filtered, output.Options{
		NoTime: flags.noTime,
		Color:  flags.colorOpt(),
	})

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

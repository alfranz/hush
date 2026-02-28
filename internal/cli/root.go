package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/alfranz/hush/internal/config"
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

	// Apply defaults from config if CLI flags not set
	cfg := loadConfigQuiet()
	f := applyDefaults(cmd, flags, cfg)

	result, err := runner.Run(context.Background(), runner.Options{
		Command: command,
		Label:   f.label,
	})
	if err != nil {
		return err
	}

	filtered := filter.Apply(result.Output, filter.Options{
		Head:      f.head,
		Tail:      f.tail,
		Grep:      f.grep,
		StripANSI: true,
	})

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, filtered)

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

// loadConfigQuiet loads config without failing on error.
func loadConfigQuiet() *config.Config {
	cfg, _ := config.Load()
	return cfg
}

// applyDefaults merges config defaults into flags where CLI flags were not explicitly set.
func applyDefaults(cmd *cobra.Command, f sharedFlags, cfg *config.Config) sharedFlags {
	if cfg == nil {
		return f
	}
	if !cmd.Flags().Changed("tail") && cfg.Defaults.Tail > 0 {
		f.tail = cfg.Defaults.Tail
	}
	if !cmd.Flags().Changed("head") && cfg.Defaults.Head > 0 {
		f.head = cfg.Defaults.Head
	}
	if !cmd.Flags().Changed("grep") && cfg.Defaults.Grep != "" {
		f.grep = cfg.Defaults.Grep
	}
	return f
}

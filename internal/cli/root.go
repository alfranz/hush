package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		Use:   "hush [flags] <command> [args...]",
		Short: "Context-efficient command runner",
		Long:  logo + "Run commands. Save tokens.\n✓ on success. Filtered output on failure. Built for agents.",
		Example: `  hush go test ./...              run the program directly with its arguments
  hush "go vet ./... && go test"  one quoted argument runs with sh -c (pipes, &&, ;)
  hush --tail 20 make check       hush flags go before the command
  hush --label t -- go test -race ./...

Hush reads its own flags only before the command. The first argument that is
not a flag starts the command. Everything after it belongs to the command.`,
		Args: cobra.ArbitraryArgs,
		RunE: runRoot,
		// Don't show errors twice
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	addSharedFlags(cmd, &flags)
	// Stop flag parsing at the first argument that is not a flag, so that
	// "hush go test -race" passes -race to go and not to hush.
	cmd.Flags().SetInterspersed(false)
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

	// One argument is a shell command string: hush "make check && make lint".
	// More arguments are a program and its arguments: hush make check.
	opts := runner.Options{Command: args[0]}
	if len(args) > 1 {
		if strings.ContainsAny(args[0], " \t") {
			return fmt.Errorf("hush: %q looks like a quoted command, but more arguments follow it: %q\n"+
				"Put hush flags before the command: hush [flags] \"<command>\"", args[0], args[1:])
		}
		opts = runner.Options{Args: args}
	}

	// Apply defaults from config if CLI flags not set
	cfg := loadConfigQuiet()
	f := applyDefaults(cmd, flags, cfg)
	opts.Label = f.label

	result, err := runner.Run(context.Background(), opts)
	if err != nil {
		return err
	}

	filtered := filter.Apply(result.Output, filter.Options{
		Head:      f.head,
		Tail:      f.tail,
		Grep:      f.grep,
		StripANSI: true,
	})
	warnings := buildWarningReport(result.Output, f)

	output.PrintResult(os.Stdout, result.Label, result.ExitCode, filtered, warnings.count, warnings.lines)

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
	if !cmd.Flags().Changed("warn-pattern") && cfg.Defaults.WarnPattern != "" {
		f.warnPattern = cfg.Defaults.WarnPattern
	}
	if !cmd.Flags().Changed("warn-tail") && cfg.Defaults.WarnTail > 0 {
		f.warnTail = cfg.Defaults.WarnTail
	}
	return f
}

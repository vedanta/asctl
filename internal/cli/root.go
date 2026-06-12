// Package cli contains the cobra command tree. Command handlers stay thin:
// they read flags and delegate to domain services (design §7); they never
// build HTTP requests or format API responses themselves.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vedanta/asctl/internal/cli/helptext"
)

// Command group IDs, in the display order of the design's root help (§6).
const (
	groupAuth    = "auth"
	groupConfig  = "config"
	groupApps    = "apps"
	groupBeta    = "beta"
	groupStore   = "store"
	groupSigning = "signing"
	groupTeam    = "team"
	groupReports = "reports"
	groupHealth  = "health"
)

// Exit codes per design §4. Centralized error-to-exit-code mapping arrives
// with the output layer (issue #7); until then commands return plain errors
// and Execute maps usage errors only.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// Options holds the parsed values of the global flags. Subcommand
// constructors receive a *Options and read it at run time, after parsing.
type Options struct {
	ConfigPath string
	Profile    string
	App        string
	Output     string
	Quiet      bool
	Verbose    bool
	Debug      bool
	Timeout    int // seconds
	NoColor    bool
	Yes        bool
	DryRun     bool
	Limit      int
	All        bool
}

// usageError marks errors that should exit with the usage exit code.
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	return run(os.Args[1:], os.Stdout, os.Stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	cmd, _ := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	if err == nil {
		return exitOK
	}
	fmt.Fprintf(stderr, "Error: %s\n", err)

	var uerr *usageError
	if errors.As(err, &uerr) || isUnknownCommand(err) {
		return exitUsage
	}
	return exitError
}

// isUnknownCommand detects cobra's untyped unknown-command error so it exits
// with the usage code like flag errors do.
func isUnknownCommand(err error) bool {
	return strings.HasPrefix(err.Error(), "unknown command")
}

// NewRootCmd builds the command tree. It returns the parsed-flag Options
// alongside the root so subcommand constructors and tests can read the
// global flags after execution.
func NewRootCmd() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:           "asctl",
		Short:         helptext.RootShort,
		Long:          helptext.RootLong,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	registerGlobalFlags(cmd, o)

	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{err}
	})

	// The completion command ships properly grouped in issue #29.
	cmd.CompletionOptions.HiddenDefaultCmd = true

	cmd.AddCommand(newVersionCmd())

	registerUsedGroups(cmd)
	cmd.SetHelpCommandGroupID(groupHealth)

	return cmd, o
}

// allGroups lists every command group in the display order of the design's
// root help (§6).
var allGroups = []*cobra.Group{
	{ID: groupAuth, Title: "Authentication:"},
	{ID: groupConfig, Title: "Configuration:"},
	{ID: groupApps, Title: "Apps & builds:"},
	{ID: groupBeta, Title: "TestFlight:"},
	{ID: groupStore, Title: "App Store:"},
	{ID: groupSigning, Title: "Code signing:"},
	{ID: groupTeam, Title: "Team:"},
	{ID: groupReports, Title: "Reports:"},
	{ID: groupHealth, Title: "Health & inventory:"},
}

// registerUsedGroups registers only the groups referenced by at least one
// command, so empty section headers never render in --help. Call it after
// all AddCommand calls. The health group is always registered because the
// built-in help command is assigned to it.
func registerUsedGroups(cmd *cobra.Command) {
	used := map[string]bool{groupHealth: true}
	for _, c := range cmd.Commands() {
		if c.GroupID != "" {
			used[c.GroupID] = true
		}
	}
	for _, g := range allGroups {
		if used[g.ID] {
			cmd.AddGroup(g)
		}
	}
}

func registerGlobalFlags(cmd *cobra.Command, o *Options) {
	f := cmd.PersistentFlags()
	f.StringVar(&o.ConfigPath, "config", "", "path to config file (default ~/.config/asctl/config.yaml)")
	f.StringVar(&o.Profile, "profile", "", "profile to use")
	f.StringVar(&o.App, "app", "", "app context (bundle ID or ID); overrides the profile default")
	f.StringVarP(&o.Output, "output", "o", "table", "output format: table, json, csv, yaml")
	f.BoolVarP(&o.Quiet, "quiet", "q", false, "suppress non-essential output")
	f.BoolVarP(&o.Verbose, "verbose", "v", false, "show more detail")
	f.BoolVar(&o.Debug, "debug", false, "log requests and responses to stderr")
	f.IntVar(&o.Timeout, "timeout", 30, "request timeout in seconds")
	f.BoolVar(&o.NoColor, "no-color", false, "disable terminal colors")
	f.BoolVarP(&o.Yes, "yes", "y", false, "apply without asking for confirmation")
	f.BoolVar(&o.DryRun, "dry-run", false, "preview what would change; touch nothing")
	f.IntVar(&o.Limit, "limit", 0, "maximum number of results to return")
	f.BoolVar(&o.All, "all", false, "fetch all pages")
}

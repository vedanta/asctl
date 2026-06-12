package cli

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var update = flag.Bool("update", false, "update golden files")

func goldenPath(name string) string {
	return filepath.Join("..", "..", "testdata", "golden", name)
}

func assertGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := goldenPath(name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file (run with -update to create): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output does not match %s\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}
}

func TestRootHelpGolden(t *testing.T) {
	cmd, _ := NewRootCmd()
	var out bytes.Buffer
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--help: %v", err)
	}
	assertGolden(t, "help-root.txt", out.Bytes())
}

func TestRootHelpGroupedLayout(t *testing.T) {
	cmd, _ := NewRootCmd()
	var out bytes.Buffer
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("--help: %v", err)
	}
	help := out.String()
	// Only groups with commands render; with the scaffold that is Health &
	// inventory. The full set appears as resource commands land (M1+).
	for _, want := range []string{
		"asctl is a CLI for managing App Store Connect",
		"Basic workflow:",
		"Health & inventory:",
		"version", "Show CLI version",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("root help missing %q", want)
		}
	}
}

func TestGlobalFlagsParse(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, o *Options)
	}{
		{
			name: "defaults",
			args: nil,
			check: func(t *testing.T, o *Options) {
				if o.Output != "table" {
					t.Errorf("Output = %q, want table", o.Output)
				}
				if o.Timeout != 30 {
					t.Errorf("Timeout = %d, want 30", o.Timeout)
				}
				if o.Quiet || o.Verbose || o.Debug || o.NoColor || o.Yes || o.DryRun || o.All {
					t.Errorf("bool flags not false by default: %+v", o)
				}
			},
		},
		{
			name: "long forms",
			args: []string{
				"--config", "/tmp/c.yaml", "--profile", "prod",
				"--app", "com.example.app", "--output", "json",
				"--quiet", "--verbose", "--debug", "--timeout", "60",
				"--no-color", "--yes", "--dry-run", "--limit", "50", "--all",
			},
			check: func(t *testing.T, o *Options) {
				want := Options{
					ConfigPath: "/tmp/c.yaml", Profile: "prod",
					App: "com.example.app", Output: "json",
					Quiet: true, Verbose: true, Debug: true, Timeout: 60,
					NoColor: true, Yes: true, DryRun: true, Limit: 50, All: true,
				}
				if *o != want {
					t.Errorf("Options = %+v, want %+v", *o, want)
				}
			},
		},
		{
			name: "short forms",
			args: []string{"-o", "csv", "-q", "-v", "-y"},
			check: func(t *testing.T, o *Options) {
				if o.Output != "csv" || !o.Quiet || !o.Verbose || !o.Yes {
					t.Errorf("short flags not parsed: %+v", o)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, o := NewRootCmd()
			// Probe subcommand: proves global flags are visible to
			// subcommands at run time, the way real commands consume them.
			var seen *Options
			cmd.AddCommand(&cobra.Command{
				Use:     "probe",
				GroupID: groupHealth,
				Run:     func(_ *cobra.Command, _ []string) { seen = o },
			})
			cmd.SetArgs(append([]string{"probe"}, tt.args...))
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute: %v", err)
			}
			if seen == nil {
				t.Fatal("probe subcommand did not run")
			}
			tt.check(t, seen)
		})
	}
}

func TestVersionCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"version"}, &out, &errOut)
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, errOut.String())
	}
	got := out.String()
	if !strings.HasPrefix(got, "asctl dev (commit none, built unknown)") {
		t.Errorf("version output = %q", got)
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"success", []string{"version"}, exitOK},
		{"unknown flag", []string{"--bogus"}, exitUsage},
		{"unknown command", []string{"frobnicate"}, exitUsage},
		{"unexpected args", []string{"version", "extra"}, exitUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if got := run(tt.args, &out, &errOut); got != tt.want {
				t.Errorf("run(%v) = %d, want %d (stderr: %s)",
					tt.args, got, tt.want, errOut.String())
			}
			if tt.want != exitOK && !strings.HasPrefix(errOut.String(), "Error: ") {
				t.Errorf("stderr should start with Error:, got %q", errOut.String())
			}
		})
	}
}

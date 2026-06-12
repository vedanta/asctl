package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vedanta/asctl/internal/cli/helptext"
)

// Build metadata, injected at release time via
//
//	-ldflags "-X github.com/vedanta/asctl/internal/cli.version=v1.2.3 ..."
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   helptext.VersionShort,
		Long:    helptext.VersionLong,
		GroupID: groupHealth,
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "asctl %s (commit %s, built %s)\n", version, commit, date)
		},
	}
}

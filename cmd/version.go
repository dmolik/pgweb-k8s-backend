package cmd

import (
	"fmt"
	"runtime"

	pg "github.com/dmolik/pgweb-k8s-backend"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(Version())
}

func Version() *cobra.Command {
	var ver = &cobra.Command{
		Use:   "version",
		Short: "pgweb-k8s-backend version",
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf(`  Version Info

    pgweb-backend-version: %s
    commit: %s
    Go: %s

`, pg.Version, pg.Build, runtime.Version())
		},
	}

	return ver
}

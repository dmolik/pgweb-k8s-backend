package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pgweb-k8s-backend",
	Short: "pgweb-k8s-backend is a backend for pgweb that provides a REST API to access Zalando postgresql clusters",
	Long: `pgweb-k8s-backend is a backend for pgweb that provides a REST API to access Zalando postgresql clusters.`,
}

func Execute() error {
	return rootCmd.Execute()
}

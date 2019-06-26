package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	namespace = ""
)

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&namespace,
		"namespace",
		"n",
		namespace,
		"Namespace to query. If not set, all namespaces are included",
	)
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-resources",
	Short: "Plugin to access Kubernetes resource requests, limits, and usage.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Hello world!")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

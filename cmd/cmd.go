package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/howardjohn/kubectl-resources/client"
)

var (
	args = &client.Args{
		Namespace:          "",
		KubeConfig:         path.Join(os.Getenv("HOME"), ".kube", "config"),
		NamespaceBlacklist: []string{"kube-system"},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&args.Namespace,
		"namespace",
		"n",
		args.Namespace,
		"Namespace to query. If not set, all namespaces are included",
	)
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-resources",
	Short: "Plugin to access Kubernetes resource requests, limits, and usage.",
	RunE: func(cmd *cobra.Command, a []string) error {
		return client.Run(args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

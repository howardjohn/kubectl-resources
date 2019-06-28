package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/howardjohn/kubectl-resources/client"
)

var (
	namespace          = ""
	kubeConfig         = path.Join(os.Getenv("HOME"), ".kube", "config")
	namespaceBlacklist = []string{"kube-system"}
	showContainers     = false
	showNodes          = false
	verbose            = false
)

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&namespace,
		"namespace",
		"n",
		namespace,
		"namespace to query. If not set, all namespaces are included",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&showContainers,
		"show-containers",
		"c",
		showContainers,
		"include container level details",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&showNodes,
		"show-nodes",
		"d",
		showNodes,
		"include node names",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		verbose,
		"show full resource names",
	)
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-resources",
	Short: "Plugin to access Kubernetes resource requests, limits, and usage.",
	RunE: func(cmd *cobra.Command, a []string) error {
		aggregation := client.Pod
		if showContainers {
			aggregation = client.None
		}
		args := &client.Args{
			Namespace:          namespace,
			KubeConfig:         kubeConfig,
			NamespaceBlacklist: namespaceBlacklist,
			Aggregation:        aggregation,
			Verbose:            verbose,
			ShowNodes:          showNodes,
		}
		return client.Run(args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

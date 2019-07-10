package cmd

import (
	"fmt"
	"os"
	"path"

	isatty "github.com/mattn/go-isatty"

	"github.com/howardjohn/kubectl-resources/pkg/model"

	"github.com/spf13/cobra"

	"github.com/howardjohn/kubectl-resources/pkg/client"
)

var (
	namespace          = ""
	kubeConfig         = path.Join(os.Getenv("HOME"), ".kube", "config")
	namespaceBlacklist = []string{"kube-system"}
	color              = isatty.IsTerminal(os.Stdout.Fd())
	showNodes          = false
	verbose            = false
	aggregation        = "POD"
	onlyWarnings       = false
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
	rootCmd.PersistentFlags().BoolVarP(
		&onlyWarnings,
		"warnings",
		"w",
		onlyWarnings,
		"only show resources using excessive resources",
	)
	rootCmd.PersistentFlags().StringVarP(
		&aggregation,
		"by",
		"b",
		aggregation,
		"column to aggregate on. Default is pod",
	)
}

var rootCmd = &cobra.Command{
	Use:   "kubectl-resources",
	Short: "Plugin to access Kubernetes resource requests, limits, and usage.",
	RunE: func(cmd *cobra.Command, a []string) error {
		agg, err := model.AggregationFromString(aggregation)
		if err != nil {
			return err
		}

		if kc, f := os.LookupEnv("KUBECONFIG"); f {
			kubeConfig = kc
		}
		args := &model.Args{
			Namespace:          namespace,
			KubeConfig:         kubeConfig,
			NamespaceBlacklist: namespaceBlacklist,
			Aggregation:        agg,
			Verbose:            verbose,
			ShowNodes:          showNodes,
			ColoredOutput:      color,
			OnlyWarnings:       onlyWarnings,
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

package cmd

import (
	"fmt"
	"os"

	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/howardjohn/kubectl-resources/pkg/model"

	"github.com/spf13/cobra"

	"github.com/howardjohn/kubectl-resources/pkg/client"
)

var (
	color         = isatty.IsTerminal(os.Stdout.Fd())
	showNodes     = false
	verbose       = true
	aggregation   = "POD"
	onlyWarnings  = false
	allNamespaces = false
)

func init() {
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
	rootCmd.PersistentFlags().BoolVarP(
		&color,
		"color",
		"c",
		color,
		"show colors for pods using excessive resources",
	)
	rootCmd.PersistentFlags().StringVarP(
		&aggregation,
		"by",
		"b",
		aggregation,
		"column to aggregate on. Default is pod",
	)
}

var (
	kubeConfigFlags = genericclioptions.NewConfigFlags(false)

	kubeResouceBuilderFlags = genericclioptions.NewResourceBuilderFlags().
				WithAllNamespaces(false).
				WithLabelSelector("")
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-resources",
	Short: "Plugin to access Kubernetes resource requests, limits, and usage.",
	RunE: func(cmd *cobra.Command, a []string) error {
		agg, err := model.AggregationFromString(aggregation)
		if err != nil {
			return err
		}

		switch agg {
		case model.Node, model.Namespace:
			kubeResouceBuilderFlags = kubeResouceBuilderFlags.WithAllNamespaces(true)
		}

		resourceFinder := kubeResouceBuilderFlags.WithAll(true).ToBuilder(kubeConfigFlags, []string{
			"pods.metrics.k8s.io,pods",
		})
		args := &model.Args{
			ResourceFinder: resourceFinder,
			AllNamespaces:  allNamespaces,
			Aggregation:    agg,
			Verbose:        verbose,
			ShowNodes:      showNodes,
			ColoredOutput:  color,
			OnlyWarnings:   onlyWarnings,
		}
		return client.Run(args)
	},
}

func Execute() {
	flags := pflag.NewFlagSet("kubectl-resources", pflag.ExitOnError)
	pflag.CommandLine = flags

	kubeConfigFlags.AddFlags(flags)
	kubeResouceBuilderFlags.AddFlags(flags)
	flags.AddFlagSet(rootCmd.PersistentFlags())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

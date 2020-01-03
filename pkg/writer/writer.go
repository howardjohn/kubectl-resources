package writer

import (
	"os"
	"strings"

	"github.com/howardjohn/kubectl-resources/pkg/model"

	"github.com/howardjohn/kubectl-resources/pkg/util"
)

func Write(response map[string]*model.PodResource, args *model.Args) error {
	resources := make([]*model.PodResource, 0, len(response))
	for _, res := range response {
		resources = append(resources, res)
	}

	if !args.Verbose {
		simplifyPodNames(resources)
		simplifyNodeNames(resources)
	}

	var allRows []ResourceRow
	for _, pod := range resources {
		if args.OnlyWarnings && !(pod.Cpu().ExceedsRequest() || pod.Memory().ExceedsRequest()) {
			continue
		}
		allRows = append(allRows, PodToRows(pod)...)
	}

	ColoredTableWriter{
		Writer: os.Stdout,
		Header: true,
		Footer: true,
		Args:   args,
	}.WriteRows(allRows)

	return nil
}

func simplifyPodNames(resources []*model.PodResource) {
	names := map[string]int{}
	for _, pod := range resources {
		parts := strings.Split(pod.Name, "-")
		shortName := strings.Join(parts[:len(parts)-2], "-")
		names[shortName]++
	}
	for _, pod := range resources {
		parts := strings.Split(pod.Name, "-")
		// Skip pods that don't follow assumptions
		if len(parts) < 3 {
			continue
		}
		shortName := strings.Join(parts[:len(parts)-2], "-")
		if names[shortName] > 1 {
			pod.Name = shortName + "-" + parts[len(parts)-1]
		} else {
			pod.Name = shortName
		}
	}
}

func simplifyNodeNames(resources []*model.PodResource) {
	var nameParts []util.Part
	for _, pod := range resources {
		if len(pod.Node) > 0 {
			nameParts = append(nameParts, strings.Split(pod.Node, "-"))
		}
	}
	lcp := strings.Join(util.LongestCommonPrefix(nameParts), "-") + "-"
	for _, pod := range resources {
		pod.Node = strings.TrimPrefix(pod.Node, lcp)
	}
}

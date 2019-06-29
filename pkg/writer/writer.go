package writer

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/howardjohn/kubectl-resources/pkg/model"

	"github.com/howardjohn/kubectl-resources/pkg/util"
)

const (
	tabwriterMinWidth = 8
	tabwriterWidth    = 8
	tabwriterPadding  = 1
	tabwriterPadChar  = '\t'
)

func sortPodResources(res []*model.PodResource) {
	sort.Slice(res, func(i, j int) bool {
		if res[i].Namespace != res[j].Namespace {
			return res[i].Namespace < res[j].Namespace
		}
		return res[i].Name < res[j].Name
	})
}

func Write(response map[string]*model.PodResource, args *model.Args) error {
	resources := make([]*model.PodResource, 0, len(response))
	for _, res := range response {
		resources = append(resources, res)
	}
	sortPodResources(resources)
	if !args.Verbose {
		simplifyPodNames(resources)
		simplifyNodeNames(resources)
	}

	w := getNewTabWriter(os.Stdout)
	if _, err := w.Write([]byte(formatHeader(args))); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	for _, res := range resources {
		rows := formatRow(res, args)
		for _, row := range rows {
			if _, err := w.Write([]byte(row)); err != nil {
				return fmt.Errorf("write failed: %v", err)
			}
		}
	}

	return w.Flush()
}

func formatHeader(args *model.Args) string {
	var headers []string
	switch args.Aggregation {
	case model.None:
		headers = append(headers, "NAMESPACE", "POD", "CONTAINER")
	case model.Pod:
		headers = append(headers, "NAMESPACE", "POD")
	case model.Namespace:
		headers = append(headers, "NAMESPACE")
	}
	if args.ShowNodes {
		headers = append(headers, "NODE")
	}
	headers = append(headers,
		"CPU USE",
		"CPU REQ",
		"CPU LIM",
		"MEM USE",
		"MEM REQ",
		"MEM LIM",
		"\n",
	)
	return strings.Join(headers, "\t")
}

func formatRow(pod *model.PodResource, args *model.Args) []string {
	rows := []string{}
	switch args.Aggregation {
	case model.None:
		for _, c := range pod.Containers {
			row := []string{
				pod.Namespace,
				pod.Name,
				c.Name,
			}
			if args.ShowNodes {
				row = append(row, pod.Node)
			}
			row = append(row,
				formatCpu(c.Cpu.Usage),
				formatCpu(c.Cpu.Request),
				formatCpu(c.Cpu.Limit),
				formatMemory(c.Memory.Usage),
				formatMemory(c.Memory.Request),
				formatMemory(c.Memory.Limit),
				"\n",
			)
			rows = append(rows, strings.Join(row, "\t"))
		}
	case model.Pod:
		row := []string{
			pod.Namespace,
			pod.Name,
		}
		if args.ShowNodes {
			row = append(row, pod.Node)
		}
		row = append(row,
			formatCpu(pod.Cpu().Usage),
			formatCpu(pod.Cpu().Request),
			formatCpu(pod.Cpu().Limit),
			formatMemory(pod.Memory().Usage),
			formatMemory(pod.Memory().Request),
			formatMemory(pod.Memory().Limit),
			"\n",
		)
		rows = append(rows, strings.Join(row, "\t"))
	}
	return rows
}

func formatCpu(i int64) string {
	if i == 0 {
		return "-"
	}
	return strconv.FormatInt(i, 10) + "m"
}

func formatMemory(i int64) string {
	if i == 0 {
		return "-"
	}
	mb := int64(float64(i) / (1024 * 1024 * 1024))
	return strconv.FormatInt(mb, 10) + "Mi"
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
		nameParts = append(nameParts, strings.Split(pod.Node, "-"))
	}
	lcp := strings.Join(util.LongestCommonPrefix(nameParts), "-") + "-"
	for _, pod := range resources {
		pod.Node = strings.TrimPrefix(pod.Node, lcp)
	}
}

// GetNewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func getNewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(output, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, 0)
}

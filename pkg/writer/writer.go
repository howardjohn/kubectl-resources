package writer

import (
	"fmt"
	"io"
	"os"
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

func Write(response map[string]*model.PodResource, args *model.Args) error {
	resources := make([]*model.PodResource, 0, len(response))
	for _, res := range response {
		resources = append(resources, res)
	}

	if !args.Verbose {
		simplifyPodNames(resources)
		simplifyNodeNames(resources)
	}

	w := getNewTabWriter(os.Stdout)
	if _, err := w.Write([]byte(formatHeader(args))); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	var allRows []*ResourceRow
	for _, pod := range resources {
		allRows = append(allRows, PodToRows(pod)...)
	}

	rows := AggregateRows(allRows, args.Aggregation)
	SortRows(rows)

	for _, row := range rows {
		if _, err := w.Write([]byte(formatRowNew(row, args))); err != nil {
			return fmt.Errorf("write failed: %v", err)
		}
	}

	footer := AggregateRows(allRows, model.Total)[0]
	footer.Name = ""
	footer.Node = ""
	footer.Namespace = ""
	footer.Container = ""
	if _, err := w.Write([]byte(formatRowNew(footer, args))); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	//for _, res := range resources {
	//	rows := formatRow(res, args)
	//	for _, row := range rows {
	//		if _, err := w.Write([]byte(row)); err != nil {
	//			return fmt.Errorf("write failed: %v", err)
	//		}
	//	}
	//}

	return w.Flush()
}

func showNode(args *model.Args) bool {
	if args.Aggregation == model.Node {
		return true
	}
	if !args.ShowNodes {
		return false
	}
	return args.Aggregation == model.Pod || args.Aggregation == model.Container
}

func formatHeader(args *model.Args) string {
	var headers []string
	switch args.Aggregation {
	case model.Container:
		headers = append(headers, "NAMESPACE", "POD", "CONTAINER")
	case model.Pod:
		headers = append(headers, "NAMESPACE", "POD")
	case model.Namespace:
		headers = append(headers, "NAMESPACE")
	}
	if showNode(args) {
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

func formatRowNew(row *ResourceRow, args *model.Args) string {
	var out []string
	switch args.Aggregation {
	case model.Container:
		out = append(out, row.Namespace, row.Name, row.Container)
	case model.Pod:
		out = append(out, row.Namespace, row.Name)
	case model.Namespace:
		out = append(out, row.Namespace)
	}
	if showNode(args) {
		out = append(out, row.Node)
	}
	out = append(out,
		formatCpu(row.Cpu.Usage),
		formatCpu(row.Cpu.Request),
		formatCpu(row.Cpu.Limit),
		formatMemory(row.Memory.Usage),
		formatMemory(row.Memory.Request),
		formatMemory(row.Memory.Limit),
		"\n",
	)
	return strings.Join(out, "\t")
}

func formatRow(pod *model.PodResource, args *model.Args) []string {
	rows := []string{}
	switch args.Aggregation {
	case model.Container:
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

package writer

import (
	"fmt"
	"github.com/howardjohn/kubectl-resources/pkg/model"
	"github.com/howardjohn/kubectl-resources/pkg/util"
	"github.com/juju/ansiterm"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	tabwriterMinWidth = 0
	tabwriterWidth    = 8
	tabwriterPadding  = 2
	tabwriterPadChar  = ' '
)

type ColorTabWriter struct {
	w *ansiterm.TabWriter
}

func (c ColorTabWriter) Write(s string, color ...ansiterm.Color) {
	if len(color) > 0 {
		c.w.SetForeground(color[0])
	}
	if _, err := c.w.Write([]byte(s + "\t")); err != nil {
		panic(err)
	}
	c.w.Reset()
}

func (c ColorTabWriter) WriteEnd(s string, color ...ansiterm.Color) {
	if len(color) > 0 {
		c.w.SetForeground(color[0])
	}
	if _, err := c.w.Write([]byte(s + "\n")); err != nil {
		panic(err)
	}
	c.w.Reset()
}

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
		writeRow(row, args, ColorTabWriter{w})
		//for i, part := range formatRow(row, args) {
		//	w.Reset()
		//	if i == 2 {
		//		w.SetForeground(ansiterm.Green)
		//	}
		//	col := ""
		//	if i > 0 {
		//		col += "\t"
		//	}
		//	col += part
		//	if _, err := w.Write([]byte(col)); err != nil {
		//		return fmt.Errorf("write failed: %v", err)
		//	}
		//}

	}
	w.SetForeground(ansiterm.Blue)
	if args.Aggregation != model.Total {
		footer := AggregateRows(allRows, model.Total)[0]
		footer.Name = ""
		footer.Node = ""
		footer.Namespace = ""
		footer.Container = ""
		if _, err := w.Write([]byte(formatRow(footer, args)[0])); err != nil {
			return fmt.Errorf("write failed: %v", err)
		}
	}

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

func writeRow(row *ResourceRow, args *model.Args, w ColorTabWriter) {
	switch args.Aggregation {
	case model.Container:
		w.Write(row.Namespace)
		w.Write(row.Name)
		w.Write(row.Container)
	case model.Pod:
		w.Write(row.Namespace)
		w.Write(row.Name)
	case model.Namespace:
		w.Write(row.Namespace)
	}
	if showNode(args) {
		w.Write(row.Node)
	}

	w.Write(formatCpu(row.Cpu.Usage))
	w.Write(formatCpu(row.Cpu.Request))
	w.Write(formatCpu(row.Cpu.Limit), ansiterm.Cyan)

	w.Write(formatMemory(row.Memory.Usage))
	w.Write(formatMemory(row.Memory.Request))
	w.WriteEnd(formatMemory(row.Memory.Limit))
}

func formatRow(row *ResourceRow, args *model.Args) []string {
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
	return out
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
		if len(pod.Node) > 0 {
			nameParts = append(nameParts, strings.Split(pod.Node, "-"))
		}
	}
	lcp := strings.Join(util.LongestCommonPrefix(nameParts), "-") + "-"
	for _, pod := range resources {
		pod.Node = strings.TrimPrefix(pod.Node, lcp)
	}
}

// GetNewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func getNewTabWriter(output io.Writer) *ansiterm.TabWriter {
	return ansiterm.NewTabWriter(output, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, 0)
}

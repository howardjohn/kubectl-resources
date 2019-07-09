package writer

import (
	"fmt"
	"io"
	"strings"

	"github.com/howardjohn/kubectl-resources/pkg/model"
)

type ColoredTableWriter struct {
	Writer io.Writer
	Header bool
	Footer bool
	Args   *model.Args
}

func (c ColoredTableWriter) WriteRows(allRows []*ResourceRow) {
	rows := AggregateRows(allRows, c.Args.Aggregation)
	SortRows(rows)

	output := [][]string{}
	if c.Header {
		output = append(output, formatHeader(c.Args))
	}
	for _, row := range rows {
		output = append(output, formatRow(row, c.Args))
	}
	if c.Footer && c.Args.Aggregation != model.Total {
		output = append(output, formatFooter(allRows, c.Args))
	}

	for _, row := range output {
		_, _ = fmt.Fprintln(c.Writer, strings.Join(row, "\t"))
	}
}

func formatHeader(args *model.Args) []string {
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
	)
	return headers
}

func formatFooter(allRows []*ResourceRow, args *model.Args) []string {
	footer := AggregateRows(allRows, model.Total)[0]
	footer.Name = ""
	footer.Node = ""
	footer.Namespace = ""
	footer.Container = ""
	return formatRow(footer, args)
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
	)
	return out
}

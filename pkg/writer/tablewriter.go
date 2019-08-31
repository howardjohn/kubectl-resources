package writer

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"

	"github.com/howardjohn/kubectl-resources/pkg/model"
)

type ColoredTableWriter struct {
	Writer io.Writer
	Header bool
	Footer bool
	Args   *model.Args
}

type StyleFunc func(arg interface{}) aurora.Value

type cell struct {
	Value  string
	Styles []StyleFunc
}

func (cell cell) String() string {
	out := aurora.Reset(cell.Value)
	for _, style := range cell.Styles {
		out = style(out)
	}
	return out.String()
}

func c(value string, styles ...StyleFunc) cell {
	filtered := []StyleFunc{}
	for _, s := range styles {
		if s != nil {
			filtered = append(filtered, s)
		}
	}
	return cell{value, filtered}
}

func (c ColoredTableWriter) getTableOutput(allRows []*ResourceRow) [][]cell {
	rows := AggregateRows(allRows, c.Args.Aggregation)
	SortRows(rows)

	output := [][]cell{}
	if c.Header {
		output = append(output, formatHeader(c.Args))
	}
	for _, row := range rows {
		output = append(output, formatRow(row, c.Args))
	}
	if c.Footer && c.Args.Aggregation != model.Total && len(allRows) > 0 {
		output = append(output, formatFooter(allRows, c.Args))
	}
	return output
}

func (c ColoredTableWriter) WriteRows(allRows []*ResourceRow) {
	output := c.getTableOutput(allRows)
	if len(output) == 0 {
		return
	}
	sep := getMaxWidths(output)
	for _, row := range output {
		for i, col := range row {
			if c.Args.ColoredOutput {
				_, _ = fmt.Fprint(c.Writer, col.String())
			} else {
				_, _ = fmt.Fprint(c.Writer, col.Value)
			}
			if i == len(row)-1 {
				_, _ = fmt.Fprint(c.Writer, "\n")
			} else {
				padAmount := sep[i] - len(col.Value) + 2
				_, _ = fmt.Fprint(c.Writer, strings.Repeat(" ", padAmount))
			}
		}
	}
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func getMaxWidths(output [][]cell) []int {
	widths := make([]int, len(output[0]))
	for _, row := range output {
		for i, col := range row {
			widths[i] = max(widths[i], len(col.Value))
		}
	}
	return widths
}

func formatHeader(args *model.Args) []cell {
	var headers []cell
	switch args.Aggregation {
	case model.Container:
		headers = append(headers, c("NAMESPACE"), c("POD"), c("CONTAINER"))
	case model.Pod:
		headers = append(headers, c("NAMESPACE"), c("POD"))
	case model.Namespace:
		headers = append(headers, c("NAMESPACE"))
	}
	if showNode(args) {
		headers = append(headers, c("NODE"))
	}
	headers = append(headers,
		c("CPU USE"),
		c("CPU REQ"),
		c("CPU LIM"),
		c("MEM USE"),
		c("MEM REQ"),
		c("MEM LIM"),
	)
	return headers
}

func formatFooter(allRows []*ResourceRow, args *model.Args) []cell {
	footer := AggregateRows(allRows, model.Total)[0]
	footer.Name = ""
	footer.Node = ""
	footer.Namespace = ""
	footer.Container = ""
	return formatRow(footer, args)
}

func formatRow(row *ResourceRow, args *model.Args) []cell {
	var out []cell
	switch args.Aggregation {
	case model.Container:
		out = append(out, c(row.Namespace), c(row.Name), c(row.Container))
	case model.Pod:
		out = append(out, c(row.Namespace), c(row.Name))
	case model.Namespace:
		out = append(out, c(row.Namespace))
	}
	if showNode(args) {
		out = append(out, c(row.Node))
	}
	out = append(out,
		c(formatCpu(row.Cpu.Usage), styleResource(row.Cpu)),
		c(formatCpu(row.Cpu.Request)),
		c(formatCpu(row.Cpu.Limit)),
		c(formatMemory(row.Memory.Usage), styleResource(row.Memory)),
		c(formatMemory(row.Memory.Request)),
		c(formatMemory(row.Memory.Limit)),
	)
	return out
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

func formatCpu(i int64) string {
	if i == 0 {
		return "-"
	}
	return strconv.FormatInt(i, 10) + "m"
}

func styleResource(r *model.Resource) StyleFunc {
	if r.Usage > r.Limit && r.Limit != 0 {
		return aurora.Red
	}
	if r.Usage > r.Request && r.Request != 0 {
		return aurora.Yellow
	}
	return nil
}

func formatMemory(i int64) string {
	if i == 0 {
		return "-"
	}
	mb := int(float64(i) / (1024 * 1024 * 1000))
	return strconv.Itoa(mb) + "Mi"
}

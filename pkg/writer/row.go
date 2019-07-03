package writer

import (
	"sort"

	"github.com/howardjohn/kubectl-resources/pkg/model"
)

type ResourceRow struct {
	Name      string
	Namespace string
	Node      string
	Container string
	Cpu       *model.Resource
	Memory    *model.Resource
}

func PodToRows(pod *model.PodResource) []*ResourceRow {
	var rows []*ResourceRow
	for _, c := range pod.Containers {
		rows = append(rows, &ResourceRow{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Node:      pod.Node,
			Container: c.Name,
			Cpu:       c.Cpu,
			Memory:    c.Memory,
		})
	}
	return rows
}

func SortRows(res []*ResourceRow) {
	sort.Slice(res, func(i, j int) bool {
		if res[i].Namespace != res[j].Namespace {
			return res[i].Namespace < res[j].Namespace
		}
		if res[i].Name != res[j].Name {
			return res[i].Name < res[j].Name
		}
		return res[i].Container < res[j].Container
	})
}

func AggregateRows(rows []*ResourceRow, aggregation model.Aggregation) []*ResourceRow {
	type Key [4]string
	var getKey func(*ResourceRow) Key
	switch aggregation {
	case model.Container:
		return rows
	case model.Namespace:
		getKey = func(row *ResourceRow) Key {
			return Key{row.Namespace}
		}
	case model.Pod:
		getKey = func(row *ResourceRow) Key {
			return Key{row.Namespace, row.Name}
		}
	case model.Node:
		getKey = func(row *ResourceRow) Key {
			return Key{row.Node}
		}
	case model.Total:
		getKey = func(row *ResourceRow) Key {
			return Key{}
		}
	}

	rowMap := make(map[Key]*ResourceRow)
	for _, row := range rows {
		key := getKey(row)
		cur, f := rowMap[key]
		if f {
			cur.Cpu.Merge(row.Cpu)
			cur.Memory.Merge(row.Memory)
		} else {
			rowMap[key] = row
		}
	}
	result := []*ResourceRow{}
	for _, row := range rowMap {
		result = append(result, row)
	}
	return result
}

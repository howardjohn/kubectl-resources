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
		if res[i].Container != res[j].Container {
			return res[i].Container < res[j].Container
		}
		return res[i].Node < res[j].Node
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
			clearKeys(cur, aggregation)
			cur.Cpu.Merge(row.Cpu)
			cur.Memory.Merge(row.Memory)
		} else {
			rowMap[key] = row
		}
	}
	var result []*ResourceRow
	for _, row := range rowMap {
		result = append(result, row)
	}
	return result
}

func clearKeys(row *ResourceRow, aggregation model.Aggregation) {
	switch aggregation {
	case model.Total:
		fallthrough
	case model.Node:
		row.Namespace = ""
		fallthrough
	case model.Namespace:
		row.Name = ""
		fallthrough
	case model.Pod:
		row.Container = ""
		fallthrough
	case model.Container:
		return
	}
}

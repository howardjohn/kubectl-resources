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
	Cpu       model.Resource
	Memory    model.Resource
}

func (r ResourceRow) toAggregate() AggregateResourceRow {
	return AggregateResourceRow{
		Name:      r.Name,
		Namespace: r.Namespace,
		Node:      r.Node,
		Container: r.Container,
		Cpu:       []model.Resource{r.Cpu},
		Memory:    []model.Resource{r.Memory},
	}
}

type AggregateResourceRow struct {
	Name      string
	Namespace string
	Node      string
	Container string
	Cpu       []model.Resource
	Memory    []model.Resource
}

func (a *AggregateResourceRow) TotalCpu() model.Resource {
	res := model.Resource{}
	for _, c := range a.Cpu {
		res = res.Merge(c)
	}
	return res
}

func (a *AggregateResourceRow) TotalMemory() model.Resource {
	res := model.Resource{}
	for _, c := range a.Memory {
		res = res.Merge(c)
	}
	return res
}

func (a *AggregateResourceRow) Add(cpu model.Resource, mem model.Resource) {
	a.Cpu = append(a.Cpu, cpu)
	a.Memory = append(a.Memory, mem)
}

func PodToRows(pod *model.PodResource) []ResourceRow {
	var rows []ResourceRow
	for _, c := range pod.Containers {
		rows = append(rows, ResourceRow{
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

func SortRows(res []AggregateResourceRow) {
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

func AggregateRows(rows []ResourceRow, aggregation model.Aggregation) []AggregateResourceRow {
	type Key [4]string
	getKey := func(row ResourceRow) Key {
		return Key{}
	}
	switch aggregation {
	case model.Container:
		getKey = func(row ResourceRow) Key {
			return Key{row.Namespace, row.Name, row.Container}
		}
	case model.Namespace:
		getKey = func(row ResourceRow) Key {
			return Key{row.Namespace}
		}
	case model.Pod:
		getKey = func(row ResourceRow) Key {
			return Key{row.Namespace, row.Name}
		}
	case model.Node:
		getKey = func(row ResourceRow) Key {
			return Key{row.Node}
		}
	case model.Total:
		getKey = func(row ResourceRow) Key {
			return Key{}
		}
	}

	rowMap := make(map[Key]AggregateResourceRow)
	for _, row := range rows {
		key := getKey(row)
		cur, f := rowMap[key]
		if f {
			cur = clearKeys(cur, aggregation)
			cur.Add(row.Cpu, row.Memory)
			rowMap[key] = cur
		} else {
			rowMap[key] = row.toAggregate()
		}
	}
	var result []AggregateResourceRow
	for _, row := range rowMap {
		result = append(result, row)
	}
	return result
}

func clearKeys(row AggregateResourceRow, aggregation model.Aggregation) AggregateResourceRow {
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
		fallthrough
	default:
		return row
	}
}

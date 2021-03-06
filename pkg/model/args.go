package model

import (
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Args struct {
	ResourceFinder genericclioptions.ResourceFinder
	Aggregation    Aggregation
	AllNamespaces  bool
	Verbose        bool
	ShowNodes      bool
	ColoredOutput  bool
	OnlyWarnings   bool
}

type Aggregation int

const (
	Container Aggregation = iota
	Pod
	Namespace
	Node
	Total
)

func AggregationFromString(s string) (Aggregation, error) {
	if strings.EqualFold(s, "container") {
		return Container, nil
	}
	if strings.EqualFold(s, "pod") {
		return Pod, nil
	}
	if strings.EqualFold(s, "namespace") {
		return Namespace, nil
	}
	if strings.EqualFold(s, "node") {
		return Node, nil
	}
	if strings.EqualFold(s, "total") {
		return Total, nil
	}
	return 0, fmt.Errorf("%v is not a valid aggregation type. Expected one of container, pod, namespace, node, total", s)
}

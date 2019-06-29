package model

type Args struct {
	Namespace          string
	KubeConfig         string
	NamespaceBlacklist []string
	Aggregation        Aggregation
	Verbose            bool
	ShowNodes          bool
}

type Aggregation int

const (
	None Aggregation = iota
	Pod
	Namespace
)

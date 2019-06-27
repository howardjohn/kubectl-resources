package client

import (
	"io"
	"os"
	"sort"
	"text/tabwriter"
)

const (
	tabwriterMinWidth = 8
	tabwriterWidth    = 8
	tabwriterPadding  = 1
	tabwriterPadChar  = '\t'
)

func sortPodResources(res []*PodResource) {
	sort.Slice(res, func(i, j int) bool {
		if res[i].Namespace != res[j].Namespace {
			return res[i].Namespace < res[j].Namespace
		}
		return res[i].Name < res[j].Name
	})
}

func Write(response map[string]*PodResource) error {
	resources := make([]*PodResource, 0, len(response))
	for _, res := range response {
		resources = append(resources, res)
	}
	sortPodResources(resources)

	w := getNewTabWriter(os.Stdout)
	if _, err := w.Write([]byte("NAME\tNAMESPACE\tNODE\n")); err != nil {
		return err
	}
	for _, pod := range resources {
		if _, err := w.Write([]byte(pod.Name + "\t" + pod.Namespace + "\t" + pod.Node + "\n")); err != nil {
			return err
		}
	}
	return w.Flush()
}

// GetNewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func getNewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(output, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, 0)
}

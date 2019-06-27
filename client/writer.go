package client

import (
	"fmt"
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

func Write(response map[string]*PodResource, args *Args) error {
	resources := make([]*PodResource, 0, len(response))
	for _, res := range response {
		resources = append(resources, res)
	}
	sortPodResources(resources)

	w := getNewTabWriter(os.Stdout)
	if _, err := w.Write([]byte(formatHeader(args))); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	for _, res := range resources {
		row := formatRow(res, args)
		if _, err := w.Write([]byte(row)); err != nil {
			return fmt.Errorf("write failed: %v", err)
		}
	}

	return w.Flush()
}

func formatHeader(args *Args) string {
	// TODO
	return "NAME"
}

func formatRow(resource *PodResource, args *Args) string {
	// TODO
	return resource.Name + "\n"
}

// GetNewTabWriter returns a tabwriter that translates tabbed columns in input into properly aligned text.
func getNewTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(output, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, 0)
}

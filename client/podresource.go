package client

import "fmt"

type Resource struct {
	Request int
	Limit   int
	Usage   int
}

type ContainerResource struct {
	Cpu    *Resource
	Memory *Resource
}

type PodResource struct {
	Name       string
	Namespace  string
	Node       string
	Containers map[string]*ContainerResource
}

func MergePodResources(resources ...map[string]*PodResource) (map[string]*PodResource, error) {
	merged := map[string]*PodResource{}
	for _, resource := range resources {
		for key, pod := range resource {
			if merged[key] != nil {
				if merged[key].Name != pod.Name {
					return nil, fmt.Errorf("attempted to merge pods with mismatched names %v %v", merged[key].Name, pod.Name)
				}
				if merged[key].Namespace != pod.Namespace {
					return nil, fmt.Errorf("attempted to merge pods with mismatched namespace %v %v", merged[key].Namespace, pod.Namespace)
				}
			} else {
				merged[key] = &PodResource{
					Name:       pod.Name,
					Namespace:  pod.Namespace,
					Containers: make(map[string]*ContainerResource),
				}
			}

			if pod.Node != "" {
				merged[key].Node = pod.Node
			}

			for containerName, container := range pod.Containers {
				if merged[key].Containers[containerName] == nil {
					merged[key].Containers[containerName] = &ContainerResource{
						Memory: &Resource{},
						Cpu:    &Resource{},
					}
				}
				c := merged[key].Containers[containerName]
				if container.Memory.Request != 0 {
					c.Memory.Request = container.Memory.Request
				}
				if container.Memory.Limit != 0 {
					c.Memory.Limit = container.Memory.Limit
				}
				if container.Memory.Usage != 0 {
					c.Memory.Usage = container.Memory.Usage
				}
				if container.Cpu.Request != 0 {
					c.Cpu.Request = container.Cpu.Request
				}
				if container.Cpu.Limit != 0 {
					c.Cpu.Limit = container.Cpu.Limit
				}
				if container.Cpu.Usage != 0 {
					c.Cpu.Usage = container.Cpu.Usage
				}
			}
		}
	}
	return merged, nil
}

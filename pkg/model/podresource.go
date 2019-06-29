package model

import "fmt"

type Resource struct {
	Request int64
	Limit   int64
	Usage   int64
}

type ContainerResource struct {
	Name   string
	Cpu    *Resource
	Memory *Resource
}

type PodResource struct {
	Name       string
	Namespace  string
	Node       string
	Containers map[string]*ContainerResource
}

func (p *PodResource) Cpu() *Resource {
	res := &Resource{}
	for _, container := range p.Containers {
		res.Limit += container.Cpu.Limit
		res.Request += container.Cpu.Request
		res.Usage += container.Cpu.Usage
	}
	return res
}

func (p *PodResource) Memory() *Resource {
	res := &Resource{}
	for _, container := range p.Containers {
		res.Limit += container.Memory.Limit
		res.Request += container.Memory.Request
		res.Usage += container.Memory.Usage
	}
	return res
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

			for _, container := range pod.Containers {
				if merged[key].Containers[container.Name] == nil {
					merged[key].Containers[container.Name] = &ContainerResource{
						Memory: &Resource{},
						Cpu:    &Resource{},
					}
				}
				c := merged[key].Containers[container.Name]
				c.Name = container.Name
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

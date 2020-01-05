package client

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/howardjohn/kubectl-resources/pkg/model"
	"github.com/howardjohn/kubectl-resources/pkg/writer"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	metrics "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func Run(args *model.Args) error {
	var responses []map[string]*model.PodResource
	err := args.ResourceFinder.Do().Visit(func(info *resource.Info, e error) error {
		switch info.Object.GetObjectKind().GroupVersionKind().Kind {
		case "PodMetrics":
			pm := &metrics.PodMetrics{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(info.Object.(*unstructured.Unstructured).Object, pm); err != nil {
				return nil
			}
			responses = append(responses, fetchMetrics(pm))
		case "Pod":
			pm := &v1.Pod{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(info.Object.(*unstructured.Unstructured).Object, pm); err != nil {
				return nil
			}
			responses = append(responses, fetchPod(pm))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to fetch resources: %v", err)
	}
	resources, err := model.MergePodResources(responses...)
	if err != nil {
		return fmt.Errorf("failed to merge responses: %v", err)
	}

	if err := writer.Write(resources, args); err != nil {
		return fmt.Errorf("faild to write: %v", err)
	}
	return nil
}

func fetchMetrics(pod *metrics.PodMetrics) map[string]*model.PodResource {
	res := map[string]*model.PodResource{}
	key := uid(pod.Name, pod.Namespace)
	res[key] = &model.PodResource{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Containers: make(map[string]*model.ContainerResource),
	}
	for _, container := range pod.Containers {
		res[key].Containers[container.Name] = &model.ContainerResource{
			Name: container.Name,
			Cpu: &model.Resource{
				Usage: container.Usage.Cpu().MilliValue(),
			},
			Memory: &model.Resource{
				Usage: container.Usage.Memory().MilliValue(),
			},
		}
	}

	return res
}

func fetchPod(pod *v1.Pod) map[string]*model.PodResource {
	res := map[string]*model.PodResource{}
	key := uid(pod.Name, pod.Namespace)

	res[key] = &model.PodResource{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Node:       pod.Spec.NodeName,
		Containers: make(map[string]*model.ContainerResource),
	}
	for _, container := range pod.Spec.Containers {
		res[key].Containers[container.Name] = &model.ContainerResource{
			Name: container.Name,
			Cpu: &model.Resource{
				Request: container.Resources.Requests.Cpu().MilliValue(),
				Limit:   container.Resources.Limits.Cpu().MilliValue(),
			},
			Memory: &model.Resource{
				Request: container.Resources.Requests.Memory().MilliValue(),
				Limit:   container.Resources.Limits.Memory().MilliValue(),
			},
		}
	}

	return res
}

func uid(name, ns string) string {
	return name + "~" + ns
}

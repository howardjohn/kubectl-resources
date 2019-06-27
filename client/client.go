package client

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Args struct {
	Namespace  string
	KubeConfig string
}

func createClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

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

func FetchMetrics(cfg *rest.Config, ns string) (map[string]*PodResource, error) {

	metricsclient, err := metrics.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	podList, err := metricsclient.MetricsV1beta1().PodMetricses(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %v", err)
	}
	if podList.Continue != "" {
		fmt.Println("Continue:", podList.Continue)
	}

	res := map[string]*PodResource{}
	for _, pod := range podList.Items {
		key := uid(pod.Name, pod.Namespace)
		res[key] = &PodResource{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Containers: make(map[string]*ContainerResource),
		}
		for _, container := range pod.Containers {
			res[key].Containers[container.Name] = &ContainerResource{
				Cpu: &Resource{
					Usage: int(container.Usage.Cpu().MilliValue()),
				},
				Memory: &Resource{
					Usage: int(container.Usage.Memory().MilliValue()),
				},
			}
		}
	}
	return res, nil
}

func uid(name, ns string) string {
	return name + "~" + ns
}

func FetchPods(cfg *rest.Config, ns string) (map[string]*PodResource, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	podList, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{FieldSelector: "status.phase=Running"})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}
	if podList.Continue != "" {
		fmt.Println("Continue:", podList.Continue)
	}

	res := map[string]*PodResource{}
	for _, pod := range podList.Items {
		key := uid(pod.Name, pod.Namespace)

		res[key] = &PodResource{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Node:       pod.Spec.NodeName,
			Containers: make(map[string]*ContainerResource),
		}
		for _, container := range pod.Spec.Containers {
			res[key].Containers[container.Name] = &ContainerResource{
				Cpu: &Resource{
					Request: int(container.Resources.Requests.Cpu().MilliValue()),
					Limit:   int(container.Resources.Limits.Cpu().MilliValue()),
				},
				Memory: &Resource{
					Request: int(container.Resources.Requests.Memory().MilliValue()),
					Limit:   int(container.Resources.Limits.Memory().MilliValue()),
				},
			}
		}
	}
	return res, nil
}

func Run(args *Args) error {
	config, err := clientcmd.BuildConfigFromFlags("", args.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %v", err)
	}
	metricsResponse, err := FetchMetrics(config, args.Namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch metrics: %v", err)
	}

	podResponse, err := FetchPods(config, args.Namespace)
	if err != nil {
		return fmt.Errorf("failed to fetch pod: %v", err)
	}

	resources, err := MergePodResources(metricsResponse, podResponse)
	if err != nil {
		return fmt.Errorf("failed to merge responses: %v", err)
	}

	if err := Write(resources); err != nil {
		return fmt.Errorf("faild to write: %v", err)
	}
	return nil
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

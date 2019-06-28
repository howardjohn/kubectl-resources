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

type Aggregation int

const (
	None Aggregation = iota
	Pod
	Namespace
)

type Args struct {
	Namespace          string
	KubeConfig         string
	NamespaceBlacklist []string
	Aggregation        Aggregation
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

	filterBlacklist(resources, args.NamespaceBlacklist)

	if err := Write(resources, args); err != nil {
		return fmt.Errorf("faild to write: %v", err)
	}
	return nil
}

func filterBlacklist(resources map[string]*PodResource, blacklist []string) {
	blMap := make(map[string]struct{})
	for _, ns := range blacklist {
		blMap[ns] = struct{}{}
	}
	for k, v := range resources {
		if _, f := blMap[v.Namespace]; f {
			delete(resources, k)
		}
	}
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
				Name: container.Name,
				Cpu: &Resource{
					Usage: container.Usage.Cpu().MilliValue(),
				},
				Memory: &Resource{
					Usage: container.Usage.Memory().MilliValue(),
				},
			}
		}
	}

	return res, nil
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
				Name: container.Name,
				Cpu: &Resource{
					Request: container.Resources.Requests.Cpu().MilliValue(),
					Limit:   container.Resources.Limits.Cpu().MilliValue(),
				},
				Memory: &Resource{
					Request: container.Resources.Requests.Memory().MilliValue(),
					Limit:   container.Resources.Limits.Memory().MilliValue(),
				},
			}
		}
	}

	return res, nil
}

func uid(name, ns string) string {
	return name + "~" + ns
}

package client

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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

func Run(args *Args) error {
	//client, err := createClient(args.KubeConfig)
	//if err != nil {
	//	return fmt.Errorf("failed to create client: %v", err)
	//}

	config, err := clientcmd.BuildConfigFromFlags("", args.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %v", err)
	}

	metricsclient, err := metrics.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create metrics client: %v", err)
	}

	podlist, err := metricsclient.MetricsV1beta1().PodMetricses(args.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod metrics: %v", err)
	}
	if podlist.Continue != "" {
		fmt.Println("Continue:", podlist.Continue)
	}

	res := map[string]*PodResource{}
	for _, pod := range podlist.Items {
		res[pod.Name] = &PodResource{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Containers: make(map[string]*ContainerResource),
		}
		for _, container := range pod.Containers {
			res[pod.Name].Containers[container.Name] = &ContainerResource{
				Cpu: &Resource{
					Usage: int(container.Usage.Cpu().MilliValue()),
				},
				Memory: &Resource{
					Usage: int(container.Usage.Memory().MilliValue()),
				},
			}
		}
	}
	return nil
}

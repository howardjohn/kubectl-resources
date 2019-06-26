package client

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Args struct {
	Namespace  string
	KubeConfig string
}

func createClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to get kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func Run(args *Args) error {
	client, err := createClient(args.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	ns, err := client.CoreV1().Namespaces().Get("default", metav1.GetOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Got namespace: %v\n", ns)
	return nil
}

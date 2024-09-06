package internal

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// MaxRetries sets the maximum number of retries for kubernetes API calls
const MaxRetries = 15

// Creates a default kubernetes client
func kubernetesDefaultClient() (clientset *kubernetes.Clientset) {
	kubeconf := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconf)
	if err != nil {
		panic(err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return
}

// Creates a dynamic kubernetes client
func kubernetesDynamicClient() (clientset *dynamic.DynamicClient) {
	kubeconf := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconf)
	if err != nil {
		panic(err)
	}
	clientset, err = dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return
}

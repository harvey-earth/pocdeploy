package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// MaxRetries sets the maximum number of retries for kubernetes API calls
const MaxRetries = 15

// Creates a default kubernetes client
func kubernetesDefaultClient() (clientset *kubernetes.Clientset, err error) {
	kubeconf := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconf)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return
}

// Creates a dynamic kubernetes client
func kubernetesDynamicClient() (clientset *dynamic.DynamicClient, err error) {
	kubeconf := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconf)
	if err != nil {
		return nil, err
	}
	clientset, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return
}

func writeTempFile(content []byte) (tempfile *os.File, err error) {
	tempfile, err = os.CreateTemp("", "pocdeploy-*.yaml")
	if err != nil {
		err = fmt.Errorf("error creating tempfile: %w", err)
		return nil, err
	}

	if _, err := tempfile.Write(content); err != nil {
		err = fmt.Errorf("error writing to tempfile %s: %w", tempfile.Name(), err)
		return nil, err
	}

	if err = tempfile.Close(); err != nil {
		err = fmt.Errorf("error closing tempfile %s: %w", tempfile.Name(), err)
		return nil, err
	}
	return
}

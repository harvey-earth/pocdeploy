package internal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespaces creates the app and monitoring namespaces
func CreateNamespaces() error {
	Info("Creating namespace")

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		err = fmt.Errorf("error creating client for namespaces: %w", err)
		return err
	}

	appNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app",
		},
	}
	if _, err = clientset.CoreV1().Namespaces().Create(context.Background(), appNS, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("error creating namespace %s: %w", string(appNS.ObjectMeta.Name), err)
		return err
	}

	Info("Namespace created")
	return nil
}

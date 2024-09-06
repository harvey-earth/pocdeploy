package internal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespaces creates the app and monitoring namespaces
func CreateNamespaces() {
	fmt.Println("Creating namespaces")

	clientset := kubernetesDefaultClient()

	appNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app",
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(context.Background(), appNS, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	monNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "monitoring",
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), monNS, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Namespaces created")
}

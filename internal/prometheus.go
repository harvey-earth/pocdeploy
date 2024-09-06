package internal

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// ConfigureMonitoring creates all necessary components for prometheus monitoring
func ConfigureMonitoring() {
	clientset := kubernetesDefaultClient()

	prometheusConfigMap(clientset)
	prometheusDeployment(clientset)
	prometheusService(clientset)

}

func prometheusConfigMap(clientset *kubernetes.Clientset) {
	cfgMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-conf",
			Namespace: "monitoring",
		},
		Data: map[string]string{
			"prometheus.yml": `
global:
  scrape_interval: 15s
  evaluation_interval: 15s
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']`,
		},
	}

	_, err := clientset.CoreV1().ConfigMaps("monitoring").Create(context.Background(), cfgMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func prometheusDeployment(clientset *kubernetes.Clientset) {
	fmt.Println("Configuring prometheus deployment")

	reps := int32(1)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus",
			Namespace: "monitoring",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &reps,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "prometheus",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "prometheus",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "prometheus",
							Image: "prom/prometheus",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9090,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/prometheus",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "prometheus-conf",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i := 1; ; i++ {
		_, err := clientset.AppsV1().Deployments("monitoring").Create(context.Background(), deployment, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Retrying prometheus deployment %d of 5\n", i+1)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				panic(err)
			}
		} else {
			break
		}
	}
	fmt.Println("Prometheus Deployment configured")
}

func prometheusService(clientset *kubernetes.Clientset) {
	fmt.Println("Configuring prometheus service")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-service",
			Namespace: "monitoring",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "prometheus",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       9090,
					TargetPort: intstr.FromInt32(9090),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := clientset.CoreV1().Services("monitoring").Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Prometheus Service configured")
}

package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	d "github.com/harvey-earth/pocdeploy/deploy"
)

// ConfigureMonitoring creates all necessary components for prometheus monitoring
func ConfigureMonitoring() error {
	// clientset, err := kubernetesDefaultClient()
	// if err != nil {
	// 	return err
	// }
	clientdyn, err := kubernetesDynamicClient()
	if err != nil {
		err = fmt.Errorf("error creating client: %w", err)
		return err
	}

	err = installPrometheus()
	if err != nil {
		err = fmt.Errorf("error installing prometheus: %w", err)
		return err
	}

	err = configurePrometheus(clientdyn)
	if err != nil {
		err = fmt.Errorf("error configuring prometheus: %w", err)
		return err
	}

	// err = prometheusConfigMap(clientset)
	// if err != nil {
	// 	return err
	// }
	// err = prometheusDeployment(clientset)
	// if err != nil {
	// 	return err
	// }
	// err = prometheusService(clientset)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func installPrometheus() error {
	fmt.Println("Installing Prometheus operator")

	promContent, err := d.DeployFiles.ReadFile("common/server/prometheus-operator.yaml")
	if err != nil {
		return err
	}
	tempfile, err := os.CreateTemp("", "prometheus-operator-*.yaml")
	if err != nil {
		err = fmt.Errorf("error creating tempfile %s: %w", tempfile.Name(), err)
		return err
	}
	defer os.Remove(tempfile.Name())

	if _, err := tempfile.Write(promContent); err != nil {
		err = fmt.Errorf("error writing to tempfile %s: %w", tempfile.Name(), err)
		return err
	}
	if err = tempfile.Close(); err != nil {
		err = fmt.Errorf("error closing tempfile %s: %w", tempfile.Name(), err)
		return err
	}

	cmd := exec.Command("kubectl", "apply", "-f", tempfile.Name())
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running kubectl: %w", err)
		return err
	}
	return nil
}

func configurePrometheus(clientset *dynamic.DynamicClient) error {
	fmt.Println("Configuring Prometheus operator")

	prometheusGVR := schema.GroupVersionResource{
		Group: "monitoring.coreos.com",
		Version: "v1",
	}

	podMonitor := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "monitoring.coreos.com/v1",
			"kind": "PodMonitor",
			"metadata": map[string]any{
				"name": "monitoring",
				"namespace": "monitoring",
				"labels": map[string]string{
					"app.kubernetes.io/component": "podmonitor",
					"app.kubernetes.io/name": "prometheus",
				},
			},
			"spec": map[string]any{
				"monitoring": map[string]string{
					"enablePodMonitor": "true",
				},
			},
		},
	}

	for i := 1; ; i++ {
		_, err := clientset.Resource(prometheusGVR).Namespace("monitoring").Create(context.Background(), podMonitor, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Retrying prometheus configuration %d of %d\n", i, MaxRetries)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for prometheus configuration: %w", err)
				return err
			}
		} else {
			break
		}
	}
	fmt.Println("Prometheus PodMonitor configured")
	return nil
}

// func prometheusConfigMap(clientset *kubernetes.Clientset) error {
// 	cfgMap := &corev1.ConfigMap{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "prometheus-conf",
// 			Namespace: "monitoring",
// 		},
// 		Data: map[string]string{
// 			"prometheus.yml": `
// global:
//   scrape_interval: 15s
//   evaluation_interval: 15s
// scrape_configs:
//   - job_name: 'prometheus'
//     static_configs:
//       - targets: ['localhost:9090']`,
// 		},
// 	}

// 	_, err := clientset.CoreV1().ConfigMaps("monitoring").Create(context.Background(), cfgMap, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func prometheusDeployment(clientset *kubernetes.Clientset) error {
// 	fmt.Println("Configuring prometheus deployment")

// 	reps := int32(1)

// 	deployment := &appsv1.Deployment{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "prometheus",
// 			Namespace: "monitoring",
// 		},
// 		Spec: appsv1.DeploymentSpec{
// 			Replicas: &reps,
// 			Selector: &metav1.LabelSelector{
// 				MatchLabels: map[string]string{
// 					"app": "prometheus",
// 				},
// 			},
// 			Template: corev1.PodTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Labels: map[string]string{
// 						"app": "prometheus",
// 					},
// 				},
// 				Spec: corev1.PodSpec{
// 					Containers: []corev1.Container{
// 						{
// 							Name:  "prometheus",
// 							Image: "prom/prometheus",
// 							Ports: []corev1.ContainerPort{
// 								{
// 									ContainerPort: 9090,
// 								},
// 							},
// 							VolumeMounts: []corev1.VolumeMount{
// 								{
// 									Name:      "config-volume",
// 									MountPath: "/etc/prometheus",
// 								},
// 							},
// 						},
// 					},
// 					Volumes: []corev1.Volume{
// 						{
// 							Name: "config-volume",
// 							VolumeSource: corev1.VolumeSource{
// 								ConfigMap: &corev1.ConfigMapVolumeSource{
// 									LocalObjectReference: corev1.LocalObjectReference{
// 										Name: "prometheus-conf",
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for i := 1; ; i++ {
// 		_, err := clientset.AppsV1().Deployments("monitoring").Create(context.Background(), deployment, metav1.CreateOptions{})
// 		if err != nil {
// 			fmt.Printf("Retrying prometheus deployment %d of 5\n", i+1)
// 			time.Sleep(time.Duration(i*2) * time.Second)
// 			if i >= MaxRetries {
// 				return err
// 			}
// 		} else {
// 			break
// 		}
// 	}
// 	fmt.Println("Prometheus Deployment configured")
// 	return nil
// }

// func prometheusService(clientset *kubernetes.Clientset) error {
// 	fmt.Println("Configuring prometheus service")

// 	service := &corev1.Service{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "prometheus-service",
// 			Namespace: "monitoring",
// 		},
// 		Spec: corev1.ServiceSpec{
// 			Selector: map[string]string{
// 				"app": "prometheus",
// 			},
// 			Ports: []corev1.ServicePort{
// 				{
// 					Port:       9090,
// 					TargetPort: intstr.FromInt32(9090),
// 					Protocol:   corev1.ProtocolTCP,
// 				},
// 			},
// 		},
// 	}

// 	_, err := clientset.CoreV1().Services("monitoring").Create(context.Background(), service, metav1.CreateOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Prometheus Service configured")
// 	return nil
// }

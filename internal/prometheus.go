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
	fmt.Println("Configuring monitoring")
	clientdyn, err := kubernetesDynamicClient()
	if err != nil {
		err = fmt.Errorf("error creating client: %w", err)
		return err
	}

	if err = configurePrometheus(clientdyn); err != nil {
		err = fmt.Errorf("error configuring prometheus: %w", err)
		return err
	}

	return nil
}

// InstallMonitoring installs the prometheus operator
func InstallMonitoring() error {
	fmt.Println("Installing monitoring")
	if err := installPrometheus(); err != nil {
		err = fmt.Errorf("error installing prometheus: %w", err)
		return err
	}
	return nil
}

func configurePrometheus(clientset *dynamic.DynamicClient) error {
	fmt.Println("Configuring Prometheus operator")

	prometheusGVR := schema.GroupVersionResource{
		Group:   "monitoring.coreos.com",
		Version: "v1",
		// Resource: "podMonitor",
	}

	prometheus := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "monitoring.coreos.com/v1",
			"kind": "Prometheus",
			"metadata": map[string]string{
				"name": "monitoring",
			},
			"spec": map[string]any{
				"podMonitorSelector": map[string]any{
					"matchLabels": map[string]string{
						"app.kubernetes.io/name": "backend",
					},
				},
				"resources": map[string]any{
					"requests": map[string]string{
						"memory": "400Mi",
					},
				},
				"enableAdminAPI": "false",
			},
		},
	}

	podMonitor := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "monitoring.coreos.com/v1",
			"kind":       "PodMonitor",
			"metadata": map[string]any{
				"name":      "monitoring",
				"namespace": "app",
				"labels": map[string]string{
					"app.kubernetes.io/component": "podmonitor",
					"app.kubernetes.io/name":      "prometheus",
				},
			},
			"spec": map[string]any{
				"monitoring": map[string]string{
					"enablePodMonitor": "true",
				},
				"selector": map[string]any{
					"matchLabels": map[string]string{
						"app.kubernetes.io/name": "backend",
					},
				},
				"podMetricsEndpoints": map[string]string{
					"port": "metrics",
				},
			},
		},
	}

	// Install Prometheus Resource
	if _, err := clientset.Resource(prometheusGVR).Namespace("app").Create(context.Background(), prometheus, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("error installing prometheus resource: %w", err)
		return err
	}

	// Install PodMonitor
	for i := 1; ; i++ {
		if _, err := clientset.Resource(prometheusGVR).Namespace("app").Create(context.Background(), podMonitor, metav1.CreateOptions{}); err != nil {
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

func installPrometheus() error {
	fmt.Println("Installing Prometheus operator")

	promContent, err := d.DeployFiles.ReadFile("common/server/prometheus-operator.yaml")
	if err != nil {
		return err
	}
	tempfile, err := writeTempFile(promContent)
	if err != nil {
		err = fmt.Errorf("error writing prometheus operator tempfile: %w", err)
	}
	defer os.Remove(tempfile.Name())

	cmd := exec.Command("kubectl", "apply", "--server-side", "-f", tempfile.Name())

	// for i := 1; ; i++ {
	// 	if err := cmd.Run(); err != nil {
	// 		fmt.Printf("Retrying prometheus operator install %d of %d\n", i, MaxRetries)
	// 		time.Sleep(time.Duration(i*2) * time.Second)
	// 		if i >= MaxRetries {
	// 			err = fmt.Errorf("error running kubectl: %w", err)
	// 			return err
	// 		}
	// 	} else {
	// 		break
	// 	}
	// }
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running kubectl: %w", err)
		return err
	}
	fmt.Println("Prometheus Operator installed")
	return nil
}

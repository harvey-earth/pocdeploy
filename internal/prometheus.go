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
	Info("Configuring monitoring")
	clientdyn, err := kubernetesDynamicClient()
	if err != nil {
		err = fmt.Errorf("error creating client: %w", err)
		return err
	}

	if err = configurePrometheus(clientdyn); err != nil {
		err = fmt.Errorf("error configuring prometheus: %w", err)
		return err
	}

	Info("Monitoring configured")
	return nil
}

// InstallMonitoring installs the prometheus operator
func InstallMonitoring() error {
	Info("Installing monitoring")
	if err := installPrometheus(PrometheusVersion); err != nil {
		err = fmt.Errorf("error installing prometheus: %w", err)
		return err
	}

	return nil
}

func configurePrometheus(clientset *dynamic.DynamicClient) error {
	Debug("Configuring Prometheus operator")

	prometheusGVR := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheuses",
	}

	podGVR := schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "podmonitors",
	}

	prometheus := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "monitoring.coreos.com/v1",
			"kind":       "Prometheus",
			"metadata": map[string]string{
				"name": "monitoring",
			},
			"spec": map[string]any{
				"serviceAccountName": "prometheus",
				"podMonitorSelector": map[string]any{
					"matchLabels": map[string]string{
						"app.kubernetes.io/name": "prometheus",
					},
				},
				"resources": map[string]any{
					"requests": map[string]string{
						"memory": "400Mi",
					},
				},
				"enableAdminAPI": false,
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
				"selector": map[string]any{
					"matchLabels": map[string]string{
						"app.kubernetes.io/name": "backend",
					},
				},
				"podMetricsEndpoints": []map[string]any{
					{
						"port": "metrics",
					},
				},
			},
		},
	}

	// Install Prometheus Resource
	for i := 1; ; i++ {
		if _, err := clientset.Resource(prometheusGVR).Namespace("app").Create(context.Background(), prometheus, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying prometheus resource configuration %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("error installing prometheus resource: %w", err)
				return err
			}
		} else {
			break
		}
	}

	// Install PodMonitor
	for i := 15; ; i++ {
		if _, err := clientset.Resource(podGVR).Namespace("app").Create(context.Background(), podMonitor, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying prometheus podmonitor configuration %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for prometheus configuration: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Debug("Prometheus PodMonitor configured")
	return nil
}

func installPrometheus(vers string) error {
	Debug("Installing Prometheus operator")

	// Write to temp file
	promContent, err := d.DeployFiles.ReadFile("common/server/prometheus-operator-" + vers + ".yaml")
	if err != nil {
		return err
	}
	tempfile, err := writeTempFile(promContent)
	if err != nil {
		err = fmt.Errorf("error writing prometheus operator tempfile: %w", err)
	}
	defer os.Remove(tempfile.Name())

	// Run install command
	cmd := exec.Command("kubectl", "apply", "--server-side", "-f", tempfile.Name())
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running kubectl: %w", err)
		return err
	}

	Debug("Prometheus Operator installed")
	return nil
}

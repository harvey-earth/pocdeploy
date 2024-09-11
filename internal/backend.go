package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	d "github.com/harvey-earth/pocdeploy/deploy"
)

// ConfigureBackend sets up CloudNative PG
func ConfigureBackend() error {
	fmt.Println("Configuring CloudNative PG Cluster")
	namespace := "app"

	clientset, err := kubernetesDynamicClient()
	if err != nil {
		return err
	}

	postgresGVR := schema.GroupVersionResource{
		Group:    "postgresql.cnpg.io",
		Version:  "v1",
		Resource: "clusters",
	}

	postgresCluster := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "postgresql.cnpg.io/v1",
			"kind":       "Cluster",
			"metadata": map[string]any{
				"name":      "poc-backend-cluster",
				"namespace": namespace,
				"labels": map[string]string{
					"app.kubernetes.io/component": "cluster",
					"app.kubernetes.io/name":      "backend",
				},
			},
			"spec": map[string]any{
				"instances": 3,
				"storage": map[string]any{
					"size": "1Gi",
				},
			},
		},
	}

	for i := 1; ; i++ {
		if _, err := clientset.Resource(postgresGVR).Namespace(namespace).Create(context.Background(), postgresCluster, metav1.CreateOptions{}); err != nil {
			fmt.Printf("Retrying backend configuration %d of %d\n", i, MaxRetries)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("reached end of retries for backend configuration: %w", err)
				return err
			}
		} else {
			break
		}
	}
	fmt.Println("CloudNative PG Cluster configured")
	return nil
}

// InitBackend creates a job in the created frontend container to run migrations
func InitBackend() error {
	fmt.Println("Starting backend migrations job")

	imgStr := viper.GetString("frontend.image") + ":" + viper.GetString("frontend.version")
	var backoffLimit int32 = 10

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		err = fmt.Errorf("error creating default client for init backend: %w", err)
		return err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend-init",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "job",
				"app.kubernetes.io/name":      "backend-init",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "backend-init",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "backend-init",
							Image: imgStr,
							Command: []string{
								"/env/bin/python",
								"/app/manage.py",
								"migrate",
							},
							ImagePullPolicy: corev1.PullNever,
							Env: []corev1.EnvVar{
								{
									Name: "DATABASE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "poc-backend-cluster-app",
											},
											Key: "dbname",
										},
									},
								},
								{
									Name: "DATABASE_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "poc-backend-cluster-app",
											},
											Key: "username",
										},
									},
								},
								{
									Name: "DATABASE_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "poc-backend-cluster-app",
											},
											Key: "password",
										},
									},
								},
								{
									Name: "DATABASE_HOST",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "poc-backend-cluster-app",
											},
											Key: "host",
										},
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	if _, err = clientset.BatchV1().Jobs("app").Create(context.Background(), job, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("error creating backend-init job: %w", err)
		return err
	}
	fmt.Println("Backend migration job started")
	return nil
}

// InstallBackend installs the CNPG operator
func InstallBackend() error {
	fmt.Println("Installing CNPG operator")

	// Use embedded cnpg-1.24.0.yaml file to pass to kubectl apply
	cnpgContent, err := d.DeployFiles.ReadFile("common/server/cnpg-1.24.0.yaml")
	if err != nil {
		return err
	}

	tempfile, err := writeTempFile(cnpgContent)
	if err != nil {
		err = fmt.Errorf("error writing CNPG tempfile: %w", err)
	}
	defer os.Remove(tempfile.Name())

	cmd := exec.Command("kubectl", "apply", "--server-side", "-f", tempfile.Name())
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error installing CNPG with kubectl: %w", err)
		return err
	}
	fmt.Println("CNPG Deployed server side")
	return nil
}

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
	Info("Configuring CloudNative PG Cluster")
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
			msg := fmt.Sprintf("Retrying backend configuration %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("reached end of retries for backend configuration: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Info("CloudNative PG Cluster configured")
	return nil
}

// InitBackend starts the job to initialize the backend for each type of framework
func InitBackend(t string) error {
	Info("Starting backend initialization")
	switch t {
	case "django":
		err := initDjangoBackend()
		if err != nil {
			err = fmt.Errorf("error initializing django backend: %w", err)
			return err
		}
	case "ror":
		err := initRORBackend()
		if err != nil {
			err = fmt.Errorf("error initializing ruby on rails backend: %w", err)
			return err
		}
	}

	Info("Backend initialized")
	return nil
}

// InitBackend creates a job in the created frontend container to run migrations
func initDjangoBackend() error {
	Debug("Starting django backend migrations job")

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

	Debug("Django backend migration job started")
	return nil
}

func initRORBackend() error {
	Debug("Starting Ruby on Rails backend migration job")

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
								"bundle",
								"exec",
								"rails",
								"db:prepare",
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
								{
									Name: "SECRET_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "secret-key",
											},
											Key: "key",
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

	Debug("Ruby on Rails Backend migration job started")
	return nil
}

// InstallBackend installs the CNPG operator
func InstallBackend() error {
	Debug("Installing CNPG operator")

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

	Debug("CNPG Operator installed")
	return nil
}

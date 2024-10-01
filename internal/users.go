package internal

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDjangoAdminUser creates a job in a Django container to create an admin user using the variables from pocdeploy.yaml
func CreateDjangoAdminUser() error {
	Info("Creating admin user creation job")

	createStr := "from django.contrib.auth import get_user_model;User = get_user_model();User.objects.create_superuser('" + viper.GetString("frontend.admin.username") + "', '" + viper.GetString("frontend.admin.email") + "', '" + viper.GetString("frontend.admin.password") + "');"
	var backoffLimit int32 = 10
	imgStr := viper.GetString("frontend.image") + ":" + viper.GetString("frontend.version")

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		return err
	}

	job := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-admin",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "job",
				"app.kubernetes.io/name":      "create-admin",
			},
		},
		Spec: v1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "create-admin",
							Image: imgStr,
							Command: []string{
								"/env/bin/python",
								"manage.py",
								"shell",
								"--command",
								createStr,
							},
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
							ImagePullPolicy: corev1.PullNever,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	jobClient := clientset.BatchV1().Jobs("app")
	_, err = jobClient.Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		err = fmt.Errorf("error creating create-admin job: %w", err)
		return err
	}

	Info("Admin user creation job started")
	return nil
}

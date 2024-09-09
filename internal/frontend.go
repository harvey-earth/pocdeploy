package internal

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	d "github.com/harvey-earth/pocdeploy/deploy"
)

// ConfigureFrontend creates a deployment, service, and ingress for the frontend built container
func ConfigureFrontend() error {

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		return err
	}
	err = frontendDeployment(clientset)
	if err != nil {
		return err
	}
	err = frontendService(clientset)
	if err != nil {
		return err
	}
	if viper.GetString("type") == "kind" {
		err = applyKindNginxIngress()
		if err != nil {
			return err
		}
	}
	err = frontendIngress(clientset)
	if err != nil {
		return err
	}

	return nil
}

// frontendDeployment creates the deployment
func frontendDeployment(clientset *kubernetes.Clientset) error {
	fmt.Println("Creating frontend deployment")

	imgStr := viper.GetString("frontend.image") + ":" + viper.GetString("frontend.version")
	reps := viper.GetInt32("frontend.size.min")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-deployment",
			Namespace: "app",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &reps,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "frontend",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "frontend",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "frontend",
							Image: imgStr,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/admin",
										Port: intstr.FromInt(8000),
									},
								},
								InitialDelaySeconds: 3,
								PeriodSeconds:       3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(8000),
									},
								},
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
								{
									Name: "SECRET_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "django-secret-key",
											},
											Key: "key",
										},
									},
								},
							},
							ImagePullPolicy: corev1.PullNever,
						},
					},
				},
			},
		},
	}

	for i := 1; ; i++ {
		_, err := clientset.AppsV1().Deployments("app").Create(context.Background(), deployment, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Retrying frontend deployment %d of 15\n", i)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				return err
			}
		} else {
			break
		}
	}
	fmt.Println("Frontend deployment configured")
	return nil
}

// frontendService creates the frontend-service
func frontendService(clientset *kubernetes.Clientset) error {
	fmt.Println("Configuring frontend service")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-service",
			Namespace: "app",
			Labels: map[string]string{
				"app": "frontend",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "frontend",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       8000,
					TargetPort: intstr.FromInt32(8000),
					NodePort:   30880,
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	for i := 1; ; i++ {
		_, err := clientset.CoreV1().Services("app").Create(context.Background(), service, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Retrying frontend service %d of 15\n", i)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
			return err
			}
		} else {
			break
		}
	}
	fmt.Println("Frontend service configured")
	return nil
}

// Creates the frontend ingress
func frontendIngress(clientset *kubernetes.Clientset) error {
	fmt.Println("Configuring frontend ingress")

	pathPtr := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-ingress",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/name": "ingress-nginx",
			},
			Annotations: map[string]string{
				// "nginx.ingress.kubernetes.io/rewrite-target": "/$1",
				"nginx.ingress.kubernetes.io/configuration-snippet": `
				location /static/ {
					root /staticfiles;
					expires 1y;
					add_header Cache-Control "public, max-age=31536000, immutable";
				}`,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/static",
									PathType: &pathPtr,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "frontend-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 8000,
											},
										},
									},
								},
								{
									Path:     "/",
									PathType: &pathPtr,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "frontend-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 8000,
											},
										},
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
		_, err := clientset.NetworkingV1().Ingresses("app").Create(context.Background(), ingress, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Retrying frontend ingress %d of 15\n", i)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				return err
			}
		} else {
			break
		}
	}
	fmt.Println("Frontend Ingress configured")
	return nil
}

// This installs nginx-ingress for Kind
func applyKindNginxIngress() error {
	ingressContent, err := d.DeployFiles.ReadFile("kind/k8s/nginx-ingress.yaml")
	if err != nil {
		return err
	}
	tempfile, err := os.CreateTemp("", "nginx-ingress-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	if _, err := tempfile.Write(ingressContent); err != nil {
		return err
	}
	if err = tempfile.Close(); err != nil {
		return err
	}
	cfgName := tempfile.Name()

	for i := 1; ; i++ {
		cmd := exec.Command("kubectl", "apply", "-f", cfgName)
		if err := cmd.Run(); err != nil {
			if i >= MaxRetries {
				return err
			} else {
				time.Sleep(time.Duration(i*2) * time.Second)
				fmt.Printf("Retrying time %d of 5\n", i)
			}
		} else {
			break
		}
	}
	fmt.Println("Kind nginx ingress deployed")
	return nil
}

// CreateSecretKeySecret creates a secret key for the Django application
func CreateSecretKeySecret() error {
	randomString, err := generateSecretKey()
	if err != nil {
		return err
	}
	encodedString := base64.StdEncoding.EncodeToString([]byte(randomString))

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "django-secret-key",
			Namespace: "app",
		},
		Data: map[string][]byte{
			"key": []byte(encodedString),
		},
		Type: v1.SecretTypeOpaque,
	}

	_, err = clientset.CoreV1().Secrets("app").Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// generateSecretKey generates a 50 character random string
func generateSecretKey() (string, error) {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+=<>?/{}-|"
	var result []byte

	for i := 0; i < 50; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return "", err
		}
		result = append(result, characters[num.Int64()])
	}
	return string(result), nil
}

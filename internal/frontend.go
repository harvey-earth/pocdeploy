package internal

import (
	"context"
	"crypto/rand"
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
	Info("Configuring frontend")

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		return err
	}
	if err = frontendDeployment(clientset); err != nil {
		err = fmt.Errorf("error with frontend deployment: %w", err)
		return err
	}
	if err = frontendService(clientset); err != nil {
		err = fmt.Errorf("error with frontend service: %w", err)
		return err
	}
	if viper.GetString("type") == "kind" {
		if err = applyKindNginxIngress(); err != nil {
			err = fmt.Errorf("error with kind nginx ingress: %w", err)
			return err
		}
	}
	if err = frontendIngress(clientset); err != nil {
		err = fmt.Errorf("error with frontend ingress: %w", err)
		return err
	}

	Info("Frontend configured")
	return nil
}

// frontendDeployment creates the deployment
func frontendDeployment(clientset *kubernetes.Clientset) error {
	Info("Creating frontend deployment")
	name := viper.GetString("frontend.image")
	vers := viper.GetString("frontend.version")
	checkPath := viper.GetString("frontend.check_path")
	imgStr := name + ":" + vers
	reps := viper.GetInt32("frontend.size.min")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-deployment",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "controller",
				"app.kubernetes.io/name":      name,
				"app.kubernetes.io/version":   vers,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &reps,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": name,
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
										Path: checkPath,
										Port: intstr.FromInt(8000),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       30,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(8000),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds: 30,
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
												Name: "secret-key",
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
		if _, err := clientset.AppsV1().Deployments("app").Create(context.Background(), deployment, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying frontend deployment %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for frontend deployment: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Info("Frontend deployment created")
	return nil
}

// frontendService creates the frontend-service
func frontendService(clientset *kubernetes.Clientset) error {
	Info("Creating frontend service")
	name := viper.GetString("frontend.image")
	vers := viper.GetString("frontend.version")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-service",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "service",
				"app.kubernetes.io/name":      name,
				"app.kubernetes.io/version":   vers,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": name,
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
		if _, err := clientset.CoreV1().Services("app").Create(context.Background(), service, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying frontend service %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for frontend service: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Info("Frontend service created")
	return nil
}

func frontendIngress(clientset *kubernetes.Clientset) error {
	Info("Creating frontend ingress")
	frontendType := viper.GetString("frontend.type")

	switch frontendType {
	case "django":
		if err := frontendDjangoIngress(clientset); err != nil {
			err = fmt.Errorf("error configuring django frontend ingress: %w", err)
			return err
		}
	case "ror":
		if err := frontendRorIngress(clientset); err != nil {
			err = fmt.Errorf("error configuring RoR frontend ingress: %w", err)
			return err
		}
	}

	Info("Frontend ingress created")
	return nil
}

// Creates the frontend ingress
func frontendDjangoIngress(clientset *kubernetes.Clientset) error {
	Debug("Configuring Django frontend ingress")

	pathPtr := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-ingress",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "ingress",
				"app.kubernetes.io/name":      "ingress-nginx",
			},
			Annotations: map[string]string{
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
		if _, err := clientset.NetworkingV1().Ingresses("app").Create(context.Background(), ingress, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying django frontend ingress %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for django frontend ingress: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Debug("Django frontend ingress configured")
	return nil
}

// Creates the frontend ingress
func frontendRorIngress(clientset *kubernetes.Clientset) error {
	Debug("Configuring RoR frontend ingress")

	pathPtr := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "frontend-ingress",
			Namespace: "app",
			Labels: map[string]string{
				"app.kubernetes.io/component": "ingress",
				"app.kubernetes.io/name":      "ingress-nginx",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
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
		if _, err := clientset.NetworkingV1().Ingresses("app").Create(context.Background(), ingress, metav1.CreateOptions{}); err != nil {
			msg := fmt.Sprintf("Retrying RoR frontend ingress %d of %d", i, MaxRetries)
			Debug(msg)
			time.Sleep(time.Duration(i*2) * time.Second)
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries for RoR frontend ingress: %w", err)
				return err
			}
		} else {
			break
		}
	}

	Debug("RoR frontend ingress configured")
	return nil
}

// This installs nginx-ingress for Kind
func applyKindNginxIngress() error {
	Debug("Installing nginx-ingress for Kind")
	ingressContent, err := d.DeployFiles.ReadFile("kind/k8s/nginx-ingress.yaml")
	if err != nil {
		return err
	}
	tempfile, err := os.CreateTemp("", "nginx-ingress-*.yaml")
	if err != nil {
		err = fmt.Errorf("error creating tempfile %s: %w", tempfile.Name(), err)
		return err
	}
	defer os.Remove(tempfile.Name())

	if _, err := tempfile.Write(ingressContent); err != nil {
		err = fmt.Errorf("error writing to tempfile %s: %w", tempfile.Name(), err)
		return err
	}
	if err = tempfile.Close(); err != nil {
		err = fmt.Errorf("error closing tempfile %s: %w", tempfile.Name(), err)
		return err
	}
	cfgName := tempfile.Name()

	for i := 1; ; i++ {
		cmd := exec.Command("kubectl", "apply", "-f", cfgName)
		if err := cmd.Run(); err != nil {
			if i >= MaxRetries {
				err = fmt.Errorf("end of retries to apply kind nginx ingress: %w", err)
				return err
			} else {
				msg := fmt.Sprintf("Retrying time %d of %d", i, MaxRetries)
				Debug(msg)
				time.Sleep(time.Duration(i*2) * time.Second)
			}
		} else {
			break
		}
	}

	Debug("Kind nginx-ingress installed")
	return nil
}

// CreateSecretKeySecret creates a secret key for the Django application
func CreateSecretKeySecret() error {
	randomString, err := generateSecretKey()
	if err != nil {
		err = fmt.Errorf("error generating secret key: %w", err)
		return err
	}
	encodedString := base64.StdEncoding.EncodeToString([]byte(randomString))

	clientset, err := kubernetesDefaultClient()
	if err != nil {
		err = fmt.Errorf("error creating client for secret django-secret-key: %w", err)
		return err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-key",
			Namespace: "app",
		},
		Data: map[string][]byte{
			"key": []byte(encodedString),
		},
		Type: v1.SecretTypeOpaque,
	}

	if _, err = clientset.CoreV1().Secrets("app").Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		err = fmt.Errorf("error creating secret key secret: %w", err)
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

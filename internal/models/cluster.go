package models

import ()

// KubernetesClusterType is an enum representing the k8s cluster type
type KubernetesClusterType int

// Enum types
const (
	Kind KubernetesClusterType = iota
	EKS
	AKS
	GKE
)

// KubernetesCluster represents the k8s cluster
type KubernetesCluster struct {
	Name    string
	Version string
	Type    KubernetesClusterType
}

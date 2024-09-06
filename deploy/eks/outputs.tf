output "cluster_endpoint" {
  description = "Endpoint for EKS Control Plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "SG ID's for Control Plane"
  value       = module.eks.cluster_security_group_id
}

output "region" {
  description = "AWS Region"
  value       = var.region
}

output "cluster_name" {
  description = "EKS Cluster Name"
  value       = var.cluster_name
}

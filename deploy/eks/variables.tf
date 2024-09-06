variable "region" {
  description = "AWS Region"
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "Name of EKS Cluster"
  type        = string
  default     = "pocdeploy-eks"
}

variable "vpc_name" {
  description = "Name of VPC for EKS Cluster"
  type        = string
  default     = "pocdeploy-vpc"
}

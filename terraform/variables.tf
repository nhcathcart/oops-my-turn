variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
}

variable "project_name" {
  description = "Base project name used in resource naming"
  type        = string
  default     = "oops-my-turn"
}

variable "environment" {
  description = "Deployment environment name"
  type        = string
}

variable "frontend_bucket_force_destroy" {
  description = "Whether Terraform may destroy the frontend bucket even if it contains objects"
  type        = bool
  default     = false
}

variable "vpc_cidr" {
  description = "CIDR block for the oops-my-turn VPC"
  type        = string
  default     = "10.42.0.0/16"
}

variable "api_image_tag" {
  description = "Container image tag to deploy for the API service"
  type        = string
  default     = "latest"
}

variable "frontend_url" {
  description = "Public frontend URL used by the API for redirects"
  type        = string
  default     = "http://localhost:5173"
}

variable "backend_url" {
  description = "Public backend URL used by the API for callbacks"
  type        = string
  default     = ""
}

variable "google_client_id" {
  description = "Google OAuth client ID"
  type        = string
  default     = "replace-me"
  sensitive   = true
}

variable "google_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  default     = "replace-me"
  sensitive   = true
}

variable "api_desired_count" {
  description = "Desired number of API tasks"
  type        = number
  default     = 1
}

variable "api_cpu" {
  description = "CPU units for the API task definition"
  type        = number
  default     = 256
}

variable "api_memory" {
  description = "Memory in MiB for the API task definition"
  type        = number
  default     = 512
}

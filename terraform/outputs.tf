output "api_ecr_repository_url" {
  description = "ECR repository URL for the API image"
  value       = aws_ecr_repository.api.repository_url
}

output "frontend_bucket_name" {
  description = "S3 bucket name for frontend assets"
  value       = aws_s3_bucket.frontend.bucket
}

output "vpc_id" {
  description = "VPC ID for the starter deployment"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs for load balancers and NAT"
  value       = values(aws_subnet.public)[*].id
}

output "private_app_subnet_ids" {
  description = "Private application subnet IDs for ECS services"
  value       = values(aws_subnet.private_app)[*].id
}

output "private_data_subnet_ids" {
  description = "Private data subnet IDs for database resources"
  value       = values(aws_subnet.private_data)[*].id
}

output "alb_security_group_id" {
  description = "Security group ID for the public ALB"
  value       = aws_security_group.alb.id
}

output "api_security_group_id" {
  description = "Security group ID for the API service"
  value       = aws_security_group.api.id
}

output "database_security_group_id" {
  description = "Security group ID for Postgres and RDS Proxy"
  value       = aws_security_group.database.id
}

output "db_proxy_security_group_id" {
  description = "Security group ID for the RDS Proxy"
  value       = aws_security_group.db_proxy.id
}

output "database_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.main.address
}

output "database_name" {
  description = "Application database name"
  value       = aws_db_instance.main.db_name
}

output "database_secret_arn" {
  description = "Secrets Manager ARN containing the database master credentials"
  value       = aws_secretsmanager_secret.db_master.arn
}

output "db_proxy_endpoint" {
  description = "RDS Proxy endpoint for application connections"
  value       = aws_db_proxy.main.endpoint
}

output "alb_dns_name" {
  description = "Public DNS name of the application load balancer"
  value       = aws_lb.api.dns_name
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "api_service_name" {
  description = "ECS service name for the API"
  value       = aws_ecs_service.api.name
}

output "migrate_task_definition_arn" {
  description = "ECS task definition ARN for one-off database migrations"
  value       = aws_ecs_task_definition.migrate.arn
}

output "app_runtime_secret_arn" {
  description = "Secrets Manager ARN containing API runtime secrets"
  value       = aws_secretsmanager_secret.app_runtime.arn
}

output "frontend_cloudfront_domain_name" {
  description = "CloudFront domain name for the frontend"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "frontend_cloudfront_distribution_id" {
  description = "CloudFront distribution ID for the frontend"
  value       = aws_cloudfront_distribution.frontend.id
}

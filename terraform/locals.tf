locals {
  name_prefix          = "${var.project_name}-${var.environment}"
  az_count             = 2
  api_image            = "${aws_ecr_repository.api.repository_url}:${var.api_image_tag}"
  resolved_backend_url = var.backend_url != "" ? var.backend_url : "http://${aws_lb.api.dns_name}"

  tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

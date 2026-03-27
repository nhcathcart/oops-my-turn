# Terraform

This directory holds the AWS deployment infrastructure for the starter template.

The initial layout is intentionally small:

- `versions.tf`: Terraform and provider version constraints
- `providers.tf`: AWS provider configuration
- `variables.tf`: shared input variables
- `locals.tf`: naming and tagging conventions
- `ecr.tf`: ECR repository for the API image
- `s3.tf`: S3 bucket for frontend assets
- `outputs.tf`: useful identifiers for later stages
- `envs/dev.tfvars`: starter values for a development environment

This stack provisions the deployed runtime for the API, database, and frontend.

## Usage

From the repo root:

```bash
terraform -chdir=terraform init
terraform -chdir=terraform plan -var-file=envs/dev.tfvars
terraform -chdir=terraform apply -var-file=envs/dev.tfvars
```

Or via Make:

```bash
make terraform-version
make terraform-fmt
make terraform-init
make terraform-plan
make terraform-apply
```

The Make targets run Terraform in Docker, so you do not need a local Terraform installation.

If you use an AWS CLI profile locally, the Make targets also mount `~/.aws` into the container. For example:

```bash
AWS_PROFILE=starter AWS_REGION=us-east-1 make terraform-plan
```

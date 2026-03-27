.DEFAULT_GOAL := help

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# --- Backend ---

.PHONY: run
run: ## Run the API server
	@nc -z localhost 5432 2>/dev/null || $(MAKE) db-up
	@set -a; [ -f .env ] && . ./.env; cd backend && go run cmd/api/main.go

.PHONY: infra-up
infra-up: ## Start local infrastructure and run migrations
	@cd backend && docker compose up -d postgres_migrated

.PHONY: db-up
db-up: ## Start postgres and run migrations
	@cd backend && docker compose up -d postgres_migrated

.PHONY: db-reset
db-reset: ## Recreate the local Postgres database and rerun migrations for the current branch
	@cd backend && docker compose rm -sf postgres migrator postgres_migrated
	@cd backend && docker compose up -d postgres_migrated

.PHONY: db-shell
db-shell: ## Open a psql shell
	@cd backend && docker compose exec postgres psql -U oops_my_turn oops_my_turn

.PHONY: new-migration
new-migration: ## Create a new timestamped migration file (usage: make new-migration NAME=add_users)
	@cd backend && \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="schemata/migrations/$${timestamp}-$(NAME).sql"; \
	printf '-- +migrate Up\n\n-- +migrate Down\n' > "$$filename"; \
	echo "Created migration: $$filename"

.PHONY: migrate-up
migrate-up: ## Apply pending migrations
	@cd backend && DB_MIGRATIONS_ENABLED=true go run cmd/migrate/main.go

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	@cd backend && DB_MIGRATIONS_ENABLED=true go run cmd/migrate/main.go --down

.PHONY: print-spec
print-spec: ## Print the OpenAPI spec as YAML
	@cd backend && go run cmd/api/main.go --print-spec

.PHONY: generate-client
generate-client: ## Generate the typed Go SDK from the OpenAPI spec
	@cd backend && go run cmd/api/main.go --print-spec > sdk/openapi.yaml
	@cd backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
		-generate "types,client" -package sdk sdk/openapi.yaml > sdk/sdk.go

.PHONY: test
test: ## Run backend unit tests
	@bash -o pipefail -c 'cd backend && go test -tags unit ./... 2>&1 | grep -v "^\?" | grep -v "\[no tests to run\]" | awk '"'"'/^ok /{printf "\033[32m✓\033[0m %s\n", substr($$0, 4); next} /^FAIL/{printf "\033[31m✗\033[0m %s\n", substr($$0, 5); next} {print}'"'"''

.PHONY: integration
integration: ## Run integration tests (requires Docker)
	@bash -o pipefail -c 'cd backend && go test -tags integration ./... 2>&1 | grep -v "^\?" | grep -v "\[no tests to run\]" | awk '"'"'/^ok /{printf "\033[32m✓\033[0m %s\n", substr($$0, 4); next} /^FAIL/{printf "\033[31m✗\033[0m %s\n", substr($$0, 5); next} {print}'"'"''

.PHONY: generate-models
generate-models: ## Generate Bob ORM models from the live database schema
	@cd backend && go run github.com/stephenafamo/bob/gen/bobgen-psql@v0.42.0 -c bobgen.yaml

.PHONY: generate
generate: migrate-up generate-models generate-client ## Run all code generators (backend models, Go SDK, TypeScript client)
	@pnpm --dir frontend run openapi-ts

.PHONY: lint
lint: ## Run all linters (backend + frontend)
	@cd backend && test -z "$$(gofmt -l .)"
	@cd backend && go vet -tags unit ./...
	@pnpm --dir frontend run lint

.PHONY: lint-fix
lint-fix: ## Auto-fix all linter issues (backend + frontend)
	@cd backend && gofmt -w .
	@pnpm --dir frontend run lint:fix

.PHONY: setup-hooks
setup-hooks: ## Configure git to use committed hooks from .githooks/
	@git config core.hooksPath .githooks
	@echo "Git hooks configured."

# --- Dev ---

.PHONY: dev
dev: ## Start backend and frontend together (backend runs in background)
	@$(MAKE) run & \
	API_PID=$$!; \
	trap "kill $$API_PID 2>/dev/null" INT TERM EXIT; \
	echo "Waiting for API to be ready..."; \
	n=0; while [ $$n -lt 30 ] && ! curl -sf http://localhost:9000/healthz > /dev/null; do \
		n=$$(($$n+1)); sleep 1; \
	done; \
	curl -sf http://localhost:9000/healthz > /dev/null || { echo "API failed to start after 30s"; exit 1; }; \
	echo "API is up. Starting frontend..."; \
	pnpm --dir frontend run dev

# --- Frontend ---

.PHONY: frontend-dev
frontend-dev: ## Start the frontend dev server (port 5173)
	@pnpm --dir frontend run dev

.PHONY: frontend-build
frontend-build: ## Build the frontend for production
	@pnpm --dir frontend run build

.PHONY: frontend-test
frontend-test: ## Run frontend unit tests with Vitest
	@pnpm --dir frontend run test

.PHONY: frontend-preview
frontend-preview: ## Preview the frontend production build
	@pnpm --dir frontend run preview

.PHONY: frontend-lint
frontend-lint: ## Run frontend linter
	@pnpm --dir frontend run lint

.PHONY: frontend-lint-fix
frontend-lint-fix: ## Auto-fix frontend linter issues
	@pnpm --dir frontend run lint:fix

.PHONY: frontend-typecheck
frontend-typecheck: ## Run TypeScript type checker on the frontend
	@pnpm --dir frontend run typecheck

# --- Containers ---

TF_IMAGE_NAME := oops-my-turn-terraform:local
TF_DOCKER_WORKDIR := /workspace
TF_BASE_DIRECTORY := $(TF_DOCKER_WORKDIR)/terraform
AWS_CONFIG_MOUNT := $(HOME)/.aws:/root/.aws:ro
AWS_PROFILE ?=
AWS_REGION ?= us-east-1
TF_ENV ?= dev
IMAGE_TAG ?= latest
API_IMAGE_REPO := oops-my-turn-$(TF_ENV)-api

.PHONY: docker-build-api
docker-build-api: ## Build the API container image
	@docker build --platform linux/amd64 -f backend/Dockerfile.api -t oops-my-turn-api:local backend

.PHONY: ecr-login
ecr-login: ## Log Docker into ECR using the active AWS CLI profile/credentials
	@ACCOUNT_ID=$$(AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws sts get-caller-identity --query Account --output text); \
	REGISTRY="$$ACCOUNT_ID.dkr.ecr.$(AWS_REGION).amazonaws.com"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin "$$REGISTRY"

.PHONY: docker-push-api
docker-push-api: docker-build-api ecr-login ## Build, tag, and push the API image to ECR
	@ACCOUNT_ID=$$(AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws sts get-caller-identity --query Account --output text); \
	REGISTRY="$$ACCOUNT_ID.dkr.ecr.$(AWS_REGION).amazonaws.com"; \
	REMOTE_IMAGE="$$REGISTRY/$(API_IMAGE_REPO):$(IMAGE_TAG)"; \
	echo "Pushing $$REMOTE_IMAGE"; \
	docker tag oops-my-turn-api:local "$$REMOTE_IMAGE"; \
	docker push "$$REMOTE_IMAGE"

.PHONY: migrate-deploy
migrate-deploy: terraform-build-image ## Run the deployed database migrations as a one-off ECS task
	@CLUSTER=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw ecs_cluster_name); \
	TASK_DEFINITION=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw migrate_task_definition_arn); \
	SUBNETS=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -json private_app_subnet_ids | tr -d '[:space:]' | jq -r 'join(",")'); \
	SECURITY_GROUP=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw api_security_group_id); \
	echo "Running migration task $$TASK_DEFINITION on cluster $$CLUSTER"; \
	TASK_ARN=$$(AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecs run-task \
		--cluster "$$CLUSTER" \
		--launch-type FARGATE \
		--task-definition "$$TASK_DEFINITION" \
		--network-configuration "awsvpcConfiguration={subnets=[$$SUBNETS],securityGroups=[$$SECURITY_GROUP],assignPublicIp=DISABLED}" \
		--query 'tasks[0].taskArn' \
		--output text); \
	echo "Waiting for $$TASK_ARN"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecs wait tasks-stopped --cluster "$$CLUSTER" --tasks "$$TASK_ARN"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecs describe-tasks --cluster "$$CLUSTER" --tasks "$$TASK_ARN" \
		--query 'tasks[0].containers[0].{exitCode:exitCode,reason:reason,lastStatus:lastStatus}' --output json

.PHONY: frontend-sync
frontend-sync: terraform-build-image frontend-build ## Build the frontend and sync it to the Terraform-managed S3 bucket
	@BUCKET=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw frontend_bucket_name); \
	echo "Syncing frontend/dist to s3://$$BUCKET"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws s3 sync frontend/dist/ "s3://$$BUCKET" --delete

.PHONY: frontend-invalidate
frontend-invalidate: terraform-build-image ## Invalidate the CloudFront frontend distribution cache
	@DISTRIBUTION_ID=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		--entrypoint /bin/sh \
		$(TF_IMAGE_NAME) -c 'terraform output -raw frontend_cloudfront_distribution_id 2>/dev/null || terraform state show -no-color aws_cloudfront_distribution.frontend | grep "^    id *= " | head -n 1 | cut -d'"'"'"'"'"' -f2'); \
	if [ -z "$$DISTRIBUTION_ID" ]; then echo "Unable to determine CloudFront distribution ID"; exit 1; fi; \
	echo "Invalidating CloudFront distribution $$DISTRIBUTION_ID"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws cloudfront create-invalidation --distribution-id "$$DISTRIBUTION_ID" --paths "/*"

.PHONY: frontend-deploy
frontend-deploy: frontend-sync frontend-invalidate ## Build, upload, and invalidate the frontend deployment

.PHONY: ecs-force-deploy
ecs-force-deploy: terraform-build-image ## Force a new ECS deployment for the API service
	@CLUSTER=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw ecs_cluster_name); \
	API_SERVICE=$$(docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) output -raw api_service_name); \
	echo "Forcing new ECS deployment for $$API_SERVICE on $$CLUSTER"; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecs update-service \
		--cluster "$$CLUSTER" \
		--service "$$API_SERVICE" \
		--force-new-deployment \
		>/dev/null; \
	AWS_PROFILE=$(AWS_PROFILE) AWS_REGION=$(AWS_REGION) aws ecs wait services-stable \
		--cluster "$$CLUSTER" \
		--services "$$API_SERVICE"

.PHONY: deploy-all
deploy-all: docker-push-api terraform-apply ecs-force-deploy frontend-deploy ## Push the API image, apply Terraform, force the API rollout, and deploy the frontend

# --- Terraform ---

.PHONY: terraform-build-image
terraform-build-image: ## Build the Terraform tooling image
	@docker build -t $(TF_IMAGE_NAME) -f terraform/Dockerfile .

.PHONY: terraform-version
terraform-version: terraform-build-image ## Print the Terraform version from the Docker image
	@docker run --rm \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) version

.PHONY: terraform-fmt
terraform-fmt: terraform-build-image ## Format Terraform files in Docker
	@docker run --rm \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) fmt -recursive .

.PHONY: terraform-init
terraform-init: terraform-build-image ## Initialize the Terraform working directory in Docker
	@docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) init

.PHONY: terraform-plan
terraform-plan: terraform-build-image ## Plan the Terraform dev environment in Docker
	@docker run --rm \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) plan -var-file=envs/dev.tfvars

.PHONY: terraform-apply
terraform-apply: terraform-build-image ## Apply the Terraform dev environment in Docker
	@docker run --rm -it \
		-e AWS_PROFILE \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_REGION \
		-e AWS_DEFAULT_REGION \
		-v $(AWS_CONFIG_MOUNT) \
		-v $(PWD):$(TF_DOCKER_WORKDIR) \
		-w $(TF_BASE_DIRECTORY) \
		$(TF_IMAGE_NAME) apply -var-file=envs/dev.tfvars

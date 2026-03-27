# Starter App

Starter App is a full-stack template with core infrastructure in place:

- Go API with Huma/OpenAPI
- React frontend with Vite, TanStack Query, and shadcn/ui
- PostgreSQL with sql-migrate and Bob ORM model generation
- Google OAuth login with signed session cookies
- Terraform for AWS deployment of the API, database, and frontend

The starter keeps one minimal vertical slice:

- `GET /api/v1/hello` public endpoint
- Google OAuth login/logout
- `GET /api/v1/me` protected endpoint
- frontend login page and protected dashboard

## Prerequisites

- Go 1.25+
- Node.js 20+
- pnpm
- Docker

## Quick Start

```bash
pnpm --dir frontend install

cp .env.example .env
# fill in Google OAuth credentials and JWT secret

make db-up
env GOCACHE=$(pwd)/.gocache make generate
make dev
```

Local endpoints:

- Frontend: `http://localhost:5173`
- API: `http://localhost:9000`
- API docs: `http://localhost:9000/docs`

## Required Environment Variables

```bash
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
JWT_SECRET=your-jwt-secret
```

Optional overrides:

- `FRONTEND_URL` defaults to `http://localhost:5173`
- `BACKEND_URL` defaults to `http://localhost:9000`
- `DB_HOST` defaults to `localhost`
- `DB_PORT` defaults to `5432`
- `DB_USER` defaults to `starter`
- `DB_PASSWORD` defaults to `12345`
- `DB_NAME` defaults to `starter`

## Common Commands

```bash
make dev
make run
make frontend-dev

make db-up
make db-reset
make db-shell
make migrate-up

make generate
make print-spec

make test
make integration
make frontend-test
make frontend-typecheck
make lint
```

## Codegen Workflow

This template keeps the database and API contract as the source of truth.

1. Update SQL migrations or Huma handler contracts.
2. Run `make generate`.
3. Use regenerated Bob models, Go SDK, and TypeScript client output.

Generated artifacts:

- `backend/models/generated/`
- `backend/sdk/sdk.go`
- `frontend/src/api/generated/`

## Project Layout

```text
backend/
  cmd/api/                  API entrypoint
  cmd/migrate/              Migration runner
  internal/server/          Routes, auth, handlers, middleware
  internal/repositories/    Database access layer
  schemata/migrations/      SQL migrations
  sdk/                      Generated OpenAPI client

frontend/
  src/api/                  Handwritten hooks over generated client
  src/pages/                Login page and protected home page
  src/components/           App shell and UI components
  src/providers/            React Query and auth providers

terraform/
  network.tf                VPC and subnets
  data.tf                   RDS, proxy, and Secrets Manager
  compute.tf                ECS API service and ALB
  frontend.tf               S3 + CloudFront frontend hosting
```

## Terraform

The Terraform stack provisions:

- API ECR repository
- ECS cluster and API service
- Application Load Balancer
- PostgreSQL RDS instance and RDS Proxy
- frontend S3 bucket and CloudFront distribution
- Secrets Manager entries for runtime secrets

See [terraform/README.md](/Users/nicholascathcart/nhc-starter/terraform/README.md) for usage.

## Turning This Into Your Own Template

Recommended next steps:

1. Rename the Go module and package metadata.
2. Update `terraform/envs/dev.tfvars` and `project_name`.
3. Add your first domain migration, repository, handler, and frontend route.

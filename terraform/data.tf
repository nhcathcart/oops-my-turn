resource "random_password" "db_master" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "db_master" {
  name                    = "${local.name_prefix}/database/master"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "db_master" {
  secret_id = aws_secretsmanager_secret.db_master.id
  secret_string = jsonencode({
    username = "starter"
    password = random_password.db_master.result
  })
}

resource "random_password" "jwt_secret" {
  length  = 48
  special = false
}

resource "aws_secretsmanager_secret" "app_runtime" {
  name                    = "${local.name_prefix}/app/runtime"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "app_runtime" {
  secret_id = aws_secretsmanager_secret.app_runtime.id
  secret_string = jsonencode({
    google_client_id     = var.google_client_id
    google_client_secret = var.google_client_secret
    jwt_secret           = random_password.jwt_secret.result
  })
}

resource "aws_db_subnet_group" "main" {
  name       = "${local.name_prefix}-db"
  subnet_ids = values(aws_subnet.private_data)[*].id
}

resource "aws_db_instance" "main" {
  identifier                  = "${local.name_prefix}-postgres"
  db_name                     = "starter"
  engine                      = "postgres"
  engine_version              = "16.13"
  instance_class              = "db.t4g.micro"
  allocated_storage           = 20
  max_allocated_storage       = 100
  storage_type                = "gp3"
  storage_encrypted           = true
  username                    = "starter"
  password                    = random_password.db_master.result
  port                        = 5432
  db_subnet_group_name        = aws_db_subnet_group.main.name
  vpc_security_group_ids      = [aws_security_group.database.id]
  backup_retention_period     = 1
  skip_final_snapshot         = true
  deletion_protection         = false
  publicly_accessible         = false
  multi_az                    = false
  auto_minor_version_upgrade  = true
  allow_major_version_upgrade = false
  apply_immediately           = true
}

resource "aws_iam_role" "rds_proxy" {
  name = "${local.name_prefix}-rds-proxy"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "rds.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "rds_proxy_secrets" {
  name = "${local.name_prefix}-rds-proxy-secrets"
  role = aws_iam_role.rds_proxy.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = aws_secretsmanager_secret.db_master.arn
      }
    ]
  })
}

resource "aws_db_proxy" "main" {
  name                   = "${local.name_prefix}-postgres"
  engine_family          = "POSTGRESQL"
  idle_client_timeout    = 1800
  require_tls            = false
  role_arn               = aws_iam_role.rds_proxy.arn
  vpc_security_group_ids = [aws_security_group.db_proxy.id]
  vpc_subnet_ids         = values(aws_subnet.private_app)[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "Master database credentials"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.db_master.arn
  }
}

resource "aws_db_proxy_default_target_group" "main" {
  db_proxy_name = aws_db_proxy.main.name
}

resource "aws_db_proxy_target" "main" {
  db_instance_identifier = aws_db_instance.main.identifier
  db_proxy_name          = aws_db_proxy.main.name
  target_group_name      = aws_db_proxy_default_target_group.main.name
}

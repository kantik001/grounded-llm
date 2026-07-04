# Grounded LLM — AWS reference (ECS Fargate + RDS)
#
# Deploys a minimal production-shaped stack:
#   - VPC with public subnets (ALB) and private subnets (ECS, RDS)
#   - RDS PostgreSQL 16
#   - ECS Fargate services: server, python, webapp
#   - Application Load Balancer → webapp
#
# Customize image repos/tags to your GHCR release. See docs/en/TERRAFORM.md.

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, { Name = "${var.name_prefix}-vpc" })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-igw" })
}

resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true
  tags                    = merge(var.tags, { Name = "${var.name_prefix}-public-${count.index}" })
}

resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + 8)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  tags              = merge(var.tags, { Name = "${var.name_prefix}-private-${count.index}" })
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = var.tags
}

resource "aws_route_table_association" "public" {
  count          = length(aws_subnet.public)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "alb" {
  name        = "${var.name_prefix}-alb"
  description = "ALB ingress"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_security_group" "ecs" {
  name        = "${var.name_prefix}-ecs"
  description = "ECS tasks"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 0
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_security_group" "rds" {
  name        = "${var.name_prefix}-rds"
  description = "Postgres from ECS"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_db_subnet_group" "main" {
  name       = "${var.name_prefix}-db"
  subnet_ids = aws_subnet.private[*].id
  tags       = var.tags
}

resource "aws_db_instance" "postgres" {
  identifier              = "${var.name_prefix}-postgres"
  engine                  = "postgres"
  engine_version          = "16"
  instance_class          = var.db_instance_class
  allocated_storage       = 20
  username                = var.db_username
  password                = var.db_password
  db_name                 = var.db_name
  vpc_security_group_ids  = [aws_security_group.rds.id]
  db_subnet_group_name    = aws_db_subnet_group.main.name
  skip_final_snapshot     = true
  publicly_accessible     = false
  backup_retention_period = 7
  tags                    = var.tags
}

resource "aws_ecs_cluster" "main" {
  name = "${var.name_prefix}-cluster"
  tags = var.tags
}

resource "aws_lb" "main" {
  name               = "${var.name_prefix}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id
  tags               = var.tags
}

resource "aws_lb_target_group" "webapp" {
  name        = "${var.name_prefix}-webapp"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path = "/"
  }

  tags = var.tags
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.webapp.arn
  }
}

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${var.name_prefix}"
  retention_in_days = 14
  tags              = var.tags
}

resource "aws_iam_role" "ecs_execution" {
  name = "${var.name_prefix}-ecs-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

locals {
  database_url = "postgres://${var.db_username}:${var.db_password}@${aws_db_instance.postgres.address}:5432/${var.db_name}?sslmode=require"
}

# Task definitions — replace image repos with your GHCR tags before apply.
resource "aws_ecs_task_definition" "server" {
  family                   = "${var.name_prefix}-server"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name  = "server"
    image = var.server_image
    portMappings = [{ containerPort = 8080, protocol = "tcp" }]
    environment = [
      { name = "DATABASE_URL", value = local.database_url },
      { name = "PYTHON_RAG_URL", value = "http://127.0.0.1:5000/rag/context" },
      { name = "PYTHON_BASE_URL", value = "http://127.0.0.1:5000" },
      { name = "LLM_BASE_URL", value = var.llm_base_url },
      { name = "LLM_MODEL", value = var.llm_model },
      { name = "DEFAULT_LOCALE", value = "en" },
    ]
    secrets = [
      { name = "LLM_API_KEY", valueFrom = var.llm_api_key_secret_arn },
      { name = "ADMIN_SECRET", valueFrom = var.admin_secret_arn },
      { name = "RAG_SERVICE_TOKEN", valueFrom = var.rag_service_token_secret_arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.ecs.name
        awslogs-region        = var.aws_region
        awslogs-stream-prefix = "server"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "python" {
  family                   = "${var.name_prefix}-python"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "1024"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name  = "python"
    image = var.python_image
    portMappings = [{ containerPort = 5000, protocol = "tcp" }]
    environment = [
      { name = "PYTHON_SERVICE_PORT", value = "5000" },
      { name = "DEFAULT_LOCALE", value = "en" },
      { name = "VECTOR_STORE", value = var.vector_store },
    ]
    secrets = [
      { name = "RAG_SERVICE_TOKEN", valueFrom = var.rag_service_token_secret_arn },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.ecs.name
        awslogs-region        = var.aws_region
        awslogs-stream-prefix = "python"
      }
    }
  }])
}

resource "aws_ecs_task_definition" "webapp" {
  family                   = "${var.name_prefix}-webapp"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name  = "webapp"
    image = var.webapp_image
    portMappings = [{ containerPort = 80, protocol = "tcp" }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.ecs.name
        awslogs-region        = var.aws_region
        awslogs-stream-prefix = "webapp"
      }
    }
  }])
}

# Note: wire ECS services, EFS/EFS for chroma+data, and sidecar networking in your overlay.
# This reference stops at task definitions + ALB + RDS so teams can extend without opinionated service mesh.

# Terraform — AWS reference

Reference Terraform for deploying Grounded LLM on **AWS ECS Fargate** with **RDS PostgreSQL 16** and an **Application Load Balancer**.

This is a **starting point**, not a production-complete module. Extend with ECS services, EFS for `chroma_db`/`data/`, Secrets Manager entries, TLS, and WAF before production.

---

## Layout

```text
deploy/terraform/aws/reference/
  main.tf          # VPC, RDS, ECS cluster, ALB, task definitions
  variables.tf
  outputs.tf
  terraform.tfvars.example
```

For Kubernetes, use the Helm chart instead: [K8S_DEPLOY.md](./K8S_DEPLOY.md).

---

## Prerequisites

- Terraform ≥ 1.5
- AWS account + credentials
- Container images (build locally or pull GHCR release tags)
- Secrets Manager ARNs for `LLM_API_KEY`, `ADMIN_SECRET`, `RAG_SERVICE_TOKEN`

---

## Quick start

```bash
cd deploy/terraform/aws/reference
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars — set db_password and secret ARNs

terraform init
terraform validate
terraform plan
terraform apply
```

Outputs:

- `alb_dns_name` — point your domain here (add ACM + HTTPS listener)
- `rds_endpoint` — Postgres hostname (private subnet)
- `ecs_cluster_name` — attach ECS services to task definitions

---

## What is included

| Resource | Purpose |
|----------|---------|
| VPC + subnets | Public (ALB) + private (ECS, RDS) |
| RDS Postgres 16 | Sessions, messages, audit |
| ECS Fargate cluster | Run server / python / webapp tasks |
| ALB + target group | Route HTTP to webapp |
| CloudWatch log group | Container logs |

Task definitions reference GHCR image variables. Wire **ECS services**, **EFS volumes** for Chroma and uploads, and **service discovery** between Go server and Python RAG in your overlay.

---

## Vector store on AWS

| Mode | Recommendation |
|------|----------------|
| Chroma | Mount EFS PVC on Python task (`CHROMA_PERSIST_DIR`) |
| Qdrant | Run Qdrant Cloud or self-hosted; set `VECTOR_STORE=qdrant` |

See [VECTOR_STORE.md](./VECTOR_STORE.md).

---

## Related

- [DEPLOY.md](./DEPLOY.md) — Docker Compose
- [NETWORK_SECURITY.md](./NETWORK_SECURITY.md)
- [BACKUP_RESTORE.md](./BACKUP_RESTORE.md)

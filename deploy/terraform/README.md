# Terraform reference modules

Cloud **starting points** for Grounded LLM. Extend with secrets, TLS, backups, and WAF before production.

Canonical operator guide: [docs/en/TERRAFORM.md](../../docs/en/TERRAFORM.md).

| Cloud | Path | Stack sketch |
|-------|------|----------------|
| AWS | [`aws/reference/`](./aws/reference/) | VPC, RDS Postgres, ECS Fargate, ALB |
| GCP | [`gcp/reference/`](./gcp/reference/) | See module + TERRAFORM.md |
| Azure | [`azure/reference/`](./azure/reference/) | See module + TERRAFORM.md |

Each `reference/` directory has `main.tf`, `variables.tf`, `outputs.tf`, and `terraform.tfvars.example`.

```bash
cd deploy/terraform/aws/reference   # or gcp / azure
cp terraform.tfvars.example terraform.tfvars
# edit tfvars — never commit real secrets
terraform init
terraform fmt -check
terraform validate
terraform plan
```

For Kubernetes, prefer the Helm chart: [`../helm/grounded-llm/`](../helm/grounded-llm/) · [K8S_DEPLOY.md](../../docs/en/K8S_DEPLOY.md).

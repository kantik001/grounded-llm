# Create a new knowledge domain (domain pack scaffold).
param(
    [Parameter(Mandatory = $true)][string]$DomainId,
    [string]$TenantId = "default"
)

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if ($DomainId -notmatch '^[a-z][a-z0-9_]*$') {
    Write-Error "domain_id must be lowercase slug"
    exit 1
}

$DataDir = Join-Path $Root "data\$TenantId\$DomainId"
New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
Write-Host "Created $DataDir"
Write-Host "Next: config/domains.json, documents, reindex_rag.py"

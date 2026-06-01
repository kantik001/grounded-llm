# Creates private GitHub repo "grounded-llm" and pushes main.
# Requires: GitHub CLI — https://cli.github.com/ — then: gh auth login
#
#   powershell -ExecutionPolicy Bypass -File scripts/create_github_repo.ps1

$ErrorActionPreference = "Stop"
$RepoName = "grounded-llm"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Error "Install GitHub CLI (gh) and run: gh auth login"
}

if (-not (Test-Path ".git")) {
    git init
    git branch -M main
}

if (git status --porcelain) {
    git add -A
    git commit -m "Initial commit: Grounded LLM platform core"
}

$login = gh api user -q .login
$exists = gh repo view "$login/$RepoName" 2>$null

if (-not $exists) {
    gh repo create $RepoName `
        --private `
        --source . `
        --remote origin `
        --description "Universal grounded LLM platform: RAG retrieval, Go orchestration, domain packs" `
        --push
} else {
    if (-not (git remote get-url origin 2>$null)) {
        git remote add origin "https://github.com/$login/$RepoName.git"
    }
    git push -u origin main
}

Write-Host "Repository: https://github.com/$login/$RepoName (private)"

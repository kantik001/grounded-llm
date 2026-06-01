# Smoke-тест API (локальный Docker или Go на :8080).
# Использование: .\scripts\smoke.ps1 [-BaseUrl "http://localhost:8080"]

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$failed = 0

function Test-Endpoint {
    param([string]$Name, [string]$Method, [string]$Path, [string]$Body = $null)
    $uri = "$BaseUrl$Path"
    try {
        if ($Body) {
            $r = Invoke-WebRequest -Uri $uri -Method $Method -Body $Body -ContentType "application/json; charset=utf-8" -UseBasicParsing -TimeoutSec 30
        } else {
            $r = Invoke-WebRequest -Uri $uri -Method $Method -UseBasicParsing -TimeoutSec 30
        }
        if ($r.StatusCode -ge 200 -and $r.StatusCode -lt 300) {
            Write-Host "[OK] $Name ($($r.StatusCode))"
            return $r.Content
        }
        Write-Host "[FAIL] $Name HTTP $($r.StatusCode)"
        $script:failed++
    } catch {
        Write-Host "[FAIL] $Name — $($_.Exception.Message)"
        $script:failed++
    }
    return $null
}

Write-Host "Smoke test: $BaseUrl"
Write-Host "(expects TELEGRAM_AUTH_DISABLED=true for /api/session)"

Test-Endpoint "health" GET "/health" | Out-Null
Test-Endpoint "domains" GET "/api/domains" | Out-Null
$sessionBody = '{"domain_id":"default"}'
$sessionJson = Test-Endpoint "session" POST "/api/session" $sessionBody
$sid = $null
if ($sessionJson) {
    try {
        $parsed = $sessionJson | ConvertFrom-Json
        $sid = $parsed.session_id
    } catch { }
}
if ($sid) {
    Test-Endpoint "onboarding" GET "/api/onboarding?domain_id=default" | Out-Null
    Write-Host "[INFO] session_id=$sid"
} else {
    Write-Host "[WARN] session: check TELEGRAM_AUTH_DISABLED or initData"
}

if ($failed -gt 0) {
    Write-Host "`nSmoke FAILED: $failed check(s)"
    exit 1
}
Write-Host "`nSmoke PASSED"
exit 0

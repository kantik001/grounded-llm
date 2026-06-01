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
Write-Host "(ожидается TELEGRAM_AUTH_DISABLED=true для /api/session)"

Test-Endpoint "health" GET "/health" | Out-Null
Test-Endpoint "crops" GET "/api/crops" | Out-Null
$sessionBody = '{"crop_id":"apple"}'
$sessionJson = Test-Endpoint "session" POST "/api/session" $sessionBody
if ($sessionJson -match '"session_id"\s*:\s*"([^"]+)"') {
    $sid = $Matches[1]
    Test-Endpoint "onboarding" GET "/api/onboarding?crop_id=apple" | Out-Null
    Write-Host "[INFO] session_id=$sid"
} else {
    Write-Host "[WARN] session: проверьте TELEGRAM_AUTH_DISABLED или initData"
}

if ($failed -gt 0) {
    Write-Host "`nSmoke FAILED: $failed check(s)"
    exit 1
}
Write-Host "`nSmoke PASSED"
exit 0

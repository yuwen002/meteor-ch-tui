# Full test script: race detector + coverage.
# Works around a local toolchain mismatch (system Go older than go.mod) and
# gcc installed per-user by winget. Environment changes are process-local only.

$ErrorActionPreference = "Stop"

function Resolve-ToolGo {
    $envGo = go env GOROOT 2>$null
    if ($envGo -and (Test-Path (Join-Path $envGo "bin\go.exe"))) {
        $v = & (Join-Path $envGo "bin\go.exe") version 2>$null
        if ($v -match "go1\.26") { return $envGo }
    }
    $roots = @(
        (Join-Path $env:USERPROFILE "go\pkg\mod\golang.org\toolchain"),
        "C:\Program Files\Go"
    )
    foreach ($root in $roots) {
        if (-not (Test-Path $root)) { continue }
        $candidate = Get-ChildItem $root -Directory -Filter "toolchain*" -ErrorAction SilentlyContinue |
            Where-Object { $_.Name -match "go1\.26" } |
            Select-Object -First 1
        if ($candidate -and (Test-Path (Join-Path $candidate.FullName "bin\go.exe"))) {
            return $candidate.FullName
        }
        if ($root -eq "C:\Program Files\Go") {
            $v = & "$root\bin\go.exe" version 2>$null
            if ($v -match "go1\.26") { return $root }
        }
    }
    throw "Go 1.26 toolchain not found. Install Go 1.26 or let GOTOOLCHAIN=auto download it."
}

function Resolve-Gcc {
    $gcc = Get-Command gcc -ErrorAction SilentlyContinue
    if ($gcc) { return $gcc.Source }
    $winget = Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages"
    $roots = @($winget, "C:\msys64", "C:\MinGW", "C:\TDM-GCC-64")
    foreach ($root in $roots) {
        if (-not (Test-Path $root)) { continue }
        $found = Get-ChildItem $root -Filter "gcc.exe" -Recurse -ErrorAction SilentlyContinue |
            Select-Object -First 1
        if ($found) { return $found.FullName }
    }
    throw "gcc not found. -race needs cgo and a C compiler. Install WinLibs/MSYS2 MinGW-w64."
}

$goroot = Resolve-ToolGo
$gccPath = Resolve-Gcc
$gccDir = Split-Path $gccPath

$env:GOROOT = $goroot
$env:GOTOOLCHAIN = "local"
$env:PATH = "$gccDir;$goroot\bin;$env:PATH"
$env:CGO_ENABLED = "1"

$go = Join-Path $goroot "bin\go.exe"

Write-Host "========== test-full ==========" -ForegroundColor Cyan
Write-Host (" GOROOT : " + $goroot)
Write-Host (" GCC    : " + $gccPath)
Write-Host (" Go     : " + (& $go version))
Write-Host "===============================" -ForegroundColor Cyan

Write-Host ""
Write-Host "[1/2] go test -race ./..." -ForegroundColor Yellow
& $go test -race ./... -count=1
if ($LASTEXITCODE -ne 0) { throw "race test failed" }

Write-Host ""
Write-Host "[2/2] go test -cover ./..." -ForegroundColor Yellow
& $go test ./... -cover -count=1
if ($LASTEXITCODE -ne 0) { throw "cover test failed" }

Write-Host ""
Write-Host "Done: race and cover both passed." -ForegroundColor Green

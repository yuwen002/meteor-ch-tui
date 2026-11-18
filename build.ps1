param(
    [string]$Out = "meteor-ch-tui.exe",
    [switch]$All
)

$ErrorActionPreference = "Stop"
$Version = git describe --tags --always --dirty 2>$null; if (-not $Version) { $Version = "dev" }
$BuildTime = (Get-Date).ToString("yyyy-MM-dd_HHmmss")
$GoVersion = (go version).Split(" ")[2]

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " meteor-ch-tui 构建脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " 版本: $Version" -ForegroundColor Gray
Write-Host " 时间: $BuildTime" -ForegroundColor Gray
Write-Host " Go:   $GoVersion" -ForegroundColor Gray
Write-Host "----------------------------------------" -ForegroundColor Cyan

$escVersion = $Version -replace "'", "''"
$escTime = $BuildTime -replace "'", "''"
$escGo = $GoVersion -replace "'", "''"
$Ldflags = "-s -w -X ""main.Version=$escVersion"" -X ""main.BuildTime=$escTime"" -X ""main.GoVersion=$escGo"""

if ($All) {
    $Targets = @(
        @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
        @{ GOOS = "linux";   GOARCH = "amd64"; Ext = "" },
        @{ GOOS = "darwin";  GOARCH = "amd64"; Ext = "" },
        @{ GOOS = "darwin";  GOARCH = "arm64"; Ext = "" }
    )
    $DistDir = Join-Path $PSScriptRoot "dist"
    if (-not (Test-Path $DistDir)) { New-Item -ItemType Directory -Path $DistDir | Out-Null }

    foreach ($t in $Targets) {
        $Name = "meteor-ch-tui-$($t.GOOS)-$($t.GOARCH)$($t.Ext)"
        $OutPath = Join-Path $DistDir $Name
        Write-Host "  -> 编译 $Name ..." -NoNewline -ForegroundColor Yellow
        $env:GOOS = $t.GOOS; $env:GOARCH = $t.GOARCH
        & go build -ldflags $Ldflags -trimpath -o $OutPath . 2>$null
        if ($LASTEXITCODE -eq 0) {
            $Size = [math]::Round((Get-Item $OutPath).Length / 1MB, 2)
            Write-Host " OK ($Size MB)" -ForegroundColor Green
        } else {
            Write-Host " 失败" -ForegroundColor Red
        }
    }
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Write-Host ""
    Write-Host "产物目录: $DistDir" -ForegroundColor Cyan
} else {
    Write-Host "  -> 编译 $Out ..." -NoNewline -ForegroundColor Yellow
    & go build -ldflags $Ldflags -trimpath -o $Out .
    if ($LASTEXITCODE -eq 0) {
        $Size = [math]::Round((Get-Item $Out).Length / 1MB, 2)
        Write-Host " OK ($Size MB)" -ForegroundColor Green
        Write-Host ""
        Write-Host "运行: .\$Out" -ForegroundColor Cyan
    } else {
        Write-Host " 失败" -ForegroundColor Red
        exit 1
    }
}
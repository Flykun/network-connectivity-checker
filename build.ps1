# build.ps1 - Go 自动交叉编译脚本
$ErrorActionPreference = "Stop"
$NAME = "network_check"
$OUT_DIR = "bin"

# 如果 bin 目录不存在则创建
if (!(Test-Path $OUT_DIR)) { New-Item -ItemType Directory -Path $OUT_DIR }

Write-Host "[1/3] Building for Linux ARM64..." -ForegroundColor Cyan
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="arm64"
go build -o "$OUT_DIR/${NAME}_arm64" .

Write-Host "[2/3] Building for Linux x86_64 (amd64)..." -ForegroundColor Cyan
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"
go build -o "$OUT_DIR/${NAME}_amd64" .

Write-Host "[3/3] Building for Windows x86_64..." -ForegroundColor Cyan
$env:CGO_ENABLED="0"; $env:GOOS="windows"; $env:GOARCH="amd64"
go build -o "$OUT_DIR/port_check.exe" .

# 恢复/清理环境变量
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host "==========================================" -ForegroundColor Green
Write-Host " Build Success! Artifacts saved in ./$OUT_DIR/" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Get-ChildItem $OUT_DIR
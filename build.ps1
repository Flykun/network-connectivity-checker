# build.ps1 - 参数化 Go 交叉编译脚本
[CmdletBinding()]
param (
    # 指定要编译的工具名称，例如: port_check 或 ping_sweeper。默认为 "all"
    [Parameter(Position = 0)]
    [string]$App = "all"
)

$ErrorActionPreference = "Stop"
$OUT_DIR = "bin"

# 如果 bin 目录不存在则创建
if (!(Test-Path $OUT_DIR)) { New-Item -ItemType Directory -Path $OUT_DIR }

# 1. 确定要编译的目标工具列表
$allApps = @("port_check", "ping_sweeper")
$targetApps = @()

if ($App -eq "all") {
    $targetApps = $allApps
}
else {
    if ($allApps -contains $App) {
        $targetApps = @($App)
    }
    else {
        Write-Host "错误: 未知工具 '$App'，可选工具为: $($allApps -join ', ')" -ForegroundColor Red
        exit 1
    }
}

# 2. 定义编译的目标架构平台 (名称, GOOS, GOARCH, 后缀)
$TARGETS = @(
    @{ Name = "Linux ARM64";  OS = "linux";   Arch = "arm64"; Ext = "";     Suffix = "_arm64" },
    @{ Name = "Linux AMD64";  OS = "linux";   Arch = "amd64"; Ext = "";     Suffix = "_x86_64" },
    @{ Name = "Windows AMD64"; OS = "windows"; Arch = "amd64"; Ext = ".exe"; Suffix = "" }   # Windows 直接用 .exe
)

Write-Host "开始按需编译项目..." -ForegroundColor Yellow

# 3. 循环执行编译
foreach ($appName in $targetApps) {
    $srcPath = "./cmd/$appName"

    if (!(Test-Path $srcPath)) {
        Write-Host "警告: 找不到源码目录 $srcPath，跳过..." -ForegroundColor Red
        continue
    }

    # 为每个工具创建对应的输出文件夹：bin/port_check、bin/ping_sweeper
    $appOutDir = "$OUT_DIR/$appName"
    if (!(Test-Path $appOutDir)) {
        New-Item -ItemType Directory -Path $appOutDir | Out-Null
    }

    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host " 正在编译工具: $appName (吐出 3 个平台包) " -ForegroundColor Cyan
    Write-Host "==========================================" -ForegroundColor Cyan

    foreach ($target in $TARGETS) {
        $env:CGO_ENABLED = "0"
        $env:GOOS = $target.OS
        $env:GOARCH = $target.Arch

        # 最终文件名：port_check_arm / port_check_x86_64 / port_check.exe
        $outputName = "${appName}$($target.Suffix)$($target.Ext)"
        $outputPath = "$appOutDir/$outputName"

        Write-Host "--> Building for $($target.Name) -> $outputName" -ForegroundColor Gray
        go build -o $outputPath $srcPath
    }
}


# 4. 清理环境变量
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "编译完成！当前 bin 目录产物：" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

Get-ChildItem $OUT_DIR
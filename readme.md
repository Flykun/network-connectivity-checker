# Network Connectivity Checker (网络端口连通性检测工具)

轻量级、并发型的网络 TCP 端口连通性检测工具，基于 Go 语言开发。支持跨平台编译，可部署于 Linux、Windows 等多种服务器及操作系统环境。

---

## 目录结构

```text
.
├── bin/                        # 编译产物存放目录
│   ├── network_check_amd64     # Linux x86_64 二进制文件
│   ├── network_check_arm64     # Linux ARM64 二进制文件
│   └── port_check.exe          # Windows x86_64 可执行文件
├── build.ps1                   # Windows PowerShell 一键交叉编译脚本
├── go.mod                      # Go 模块配置文件
├── main.go                     # 主程序源码
└── README.md                   # 项目说明文档
```

---

## 功能特性

- **高并发检测**：使用 Go 协程（Goroutines）并行测试所有目标 IP/端口，检测耗时极短。
- **错误区分**：清晰区分连接超时（Timeout）、网络拒绝、地址无效等详细错误类型。
- **环境探测**：自动识别运行时操作系统（OS）、CPU 架构及 Go 运行环境版本。
- **跨平台支持**：支持本地 Windows 双击防闪退交互，同时完美适配 Linux 自动化无交互执行。

---

## 一键编译

本项目已配置自动化编译脚本 `build.ps1`，可在本地 Windows (PowerShell) 环境下一键构建所有架构的产物。

### 1. 执行全量编译

在 VSCode 终端或 PowerShell 中运行：

```powershell
./build.ps1
```

脚本会自动打包以下三个版本的二进制文件并保存至 `bin/` 目录：

1. `network_check_arm64` (Linux ARM64 / 华为云 / 飞腾等)
2. `network_check_amd64` (Linux x86_64 / Intel / AMD)
3. `port_check.exe` (Windows x86_64)

---

## 使用指南

### 1. Linux 服务器部署与运行

通过 SCP 或 SFTP 将 `bin/` 目录下对应的 Linux 版本文件上传至目标服务器：

```bash
# 1. 赋予执行权限
chmod +x ./network_check_amd64

# 2. 运行检测
./network_check_amd64
```

### 2. Windows 本地测试

直接双击 `bin/port_check.exe` 或在 CMD/PowerShell 中运行：

```powershell
. in\port_check.exe
```

*注：Windows 环境下运行完毕后会等待按下 `Enter` 键退出，防止控制台窗口闪退。*

---

## 目标配置与自定义

如需调整测试的目标 IP 或端口，直接修改 `main.go` 中的 `targets` 列表与超时时间：

```go
// main.go
targets := []string{
    "114.114.114.114:53",
    "8.8.8.8:53",
    "1.1.1.1:53",
}

timeout := 2 * time.Second
```

修改后保存，重新运行 `./build.ps1` 即可生效。

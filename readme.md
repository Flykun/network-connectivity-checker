
# Go Network Tools (Go 网络工具集)

这是一个基于 Go 语言编写的高效轻量级网络运维工具集，支持单工具按需编译与多平台交叉编译。

---

## 🛠️ 内置工具说明


| 工具名称           | 源码入口                   | 功能说明                                               |
| :------------------- | :--------------------------- | :------------------------------------------------------- |
| **`port_check`**   | `cmd/port_check/main.go`   | TCP 端口连通性检测工具，支持批量扫描与目标端口状态排查 |
| **`ping_sweeper`** | `cmd/ping_sweeper/main.go` | 局域网 / IP 网段存活主机快速 ICMP / Ping 扫描工具      |

---

## 📂 项目目录结构

```text
go_learn/
├── bin/                    # 编译产物输出目录（已在 .gitignore 中忽略）
│   ├── ping_sweeper/       # ping_sweeper 各平台二进制文件
│   │   ├── ping_sweeper_arm64
│   │   ├── ping_sweeper_x86_64
│   │   └── ping_sweeper.exe
│   └── port_check/         # port_check 各平台二进制文件
│       ├── port_check_arm64
│       ├── port_check_x86_64
│       └── port_check.exe
├── cmd/                    # 工具源码目录
│   ├── ping_sweeper/
│   │   └── main.go
│   └── port_check/
│       └── main.go
├── build.ps1               # PowerShell 参数化交叉编译脚本
├── go.mod                  # Go 模块文件
├── .gitignore              # Git 忽略文件
└── README.md               # 项目说明文档
```

---

## 🚀 编译与构建指南

项目提供 `build.ps1` 自动化脚本，支持一键将 Go 源码交叉编译为多个平台的无依赖二进制可执行文件。

### 1. 支持的架构与平台

每次编译会为目标工具自动生成以下 3 个平台的二进制包：


| 目标平台 (OS / Arch) | 产物命名规则        | 典型应用场景                                   |
| :--------------------- | :-------------------- | :----------------------------------------------- |
| **Linux ARM64**      | `<app_name>_arm64`  | 鲲鹏 / 飞腾服务器、树莓派、ARM 架构 Linux 容器 |
| **Linux AMD64**      | `<app_name>_x86_64` | 常见 Linux 服务器 (CentOS / Ubuntu / Debian)   |
| **Windows AMD64**    | `<app_name>.exe`    | Windows 64位 PC / 服务器                       |

### 2. 编译命令示例

在 PowerShell 中，进入项目根目录执行：

#### ① 编译全部工具（默认）

```powershell
./build.ps1
# 或
./build.ps1 -App all
```

#### ② 仅编译 `port_check` 工具

```powershell
./build.ps1 port_check
# 或
./build.ps1 -App port_check
```

#### ③ 仅编译 `ping_sweeper` 工具

```powershell
./build.ps1 ping_sweeper
# 或
./build.ps1 -App ping_sweeper
```

---

## ⚙️ 注意事项与环境说明

1. **PowerShell 编码处理**
   * 在 Windows 5.1 环境下执行 `build.ps1` 脚本时，如果遇到 `字符串缺少终止符` 等编码报错，请确保将 `build.ps1` 保存为 **UTF-8 with BOM** 格式。
2. **Go 环境要求**
   * 本项目依赖 Go 1.18+ 标准库，支持静默依赖打包，无需额外的第三方 C 库支持（脚本内已强制禁用 CGO，即 `CGO_ENABLED=0`）。

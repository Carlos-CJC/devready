# DevScope

[![CI](https://github.com/Carlos-CJC/devscope/actions/workflows/ci.yml/badge.svg)](https://github.com/Carlos-CJC/devscope/actions/workflows/ci.yml)

> 面向本机开发环境的终端状态面板 —— 不在项目里时看这台机器,进入项目后看环境是否就绪。

[English](README.md) | **中文**

> ⚠️ **早期阶段。** Machine Scope(本机系统面板,macOS)与 Project Scope 的 Go 检查已可用,详见[状态](#状态)。

## 这是什么

DevScope **不是** `btop`、`lazygit` 或 `lazydocker` 的替代品,它回答的是另一个问题:

> 这台开发机现在处于什么状态?进入项目后,当前环境满足这个项目的要求吗?

它根据**当前所在目录**自动在两种 Scope 间切换:

- **Machine Scope** —— 不在项目里。`devscope` 展示 CPU、内存、磁盘、网络、已安装开发工具、服务与监听端口。
- **Project Scope** —— 在项目里。自动识别项目类型,然后查看要求(`info`)与检查环境(`check`)。

## 状态

早期、持续开发中。目前可用的部分:

| 模块 | 状态 |
|---|---|
| CLI:`devscope` / `info` / `check` / `port` | ✅ |
| Machine Scope:CPU / 内存 / 磁盘 / 网络(macOS) | ✅ |
| Machine Scope:开发工具(Git、Go、Docker)+ 服务(sshd、Docker) | ✅ |
| `devscope port`:经 `lsof` 列出监听端口 | ✅ |
| Go:项目识别(`go.mod`)+ 版本需求解析 | ✅ |
| Go:环境侦测(已装/激活版本,识别 asdf) | ✅ |
| Go:版本检查 → PASS/FAIL/WARN/SKIP + 汇总 | ✅ |
| 其它工具(Python / Docker / LaTeX / PlatformIO)侦测 | ⬜ 计划中 |
| TUI(Bubble Tea / Lip Gloss) | ⬜ 计划中 |
| Linux 后端(当前指标诚实降级为 SKIP) | ⬜ 计划中 |

## 安装

需要 Go 1.22+。

```bash
go install github.com/Carlos-CJC/devscope/cmd/devscope@latest
```

或从源码构建:

```bash
git clone https://github.com/Carlos-CJC/devscope
cd devscope
go build -o devscope ./cmd/devscope
```

## 用法

```bash
devscope          # 机器 / 项目概况
devscope info     # 查看项目要求(不执行检查)
devscope check    # 检查当前环境是否满足要求
devscope port     # 监听端口(macOS,经 lsof)
```

### `devscope` —— Machine / Project Scope

```
DevScope  my-mac

SYSTEM

  CPU     Apple M4  10 核
          ███░░░░░░░░░░░░░░░░░░░░░  13%
  LOAD    2.34  1.71  1.53

  MEMORY  9.2 GB / 16.0 GB
          ██████████████░░░░░░░░░░  58%

  STORAGE
  System          17.1 GB / 228 GB   █░░░░░░░░░░░░  7%
  /Volumes/Data   2.1 GB / 931 GB    ░░░░░░░░░░░░░  <1%

  NETWORK
  en1             ↓ 18.9 KB/s   ↑ 99.9 KB/s

DEVELOPMENT
  ✓ Git    2.54.0  (system)
  ✓ Go     1.22.0  (asdf)
  ✓ Docker 29.8.1  (system)

SERVICES
  ○ sshd       未监听
  ○ Docker     未运行
```

### `devscope info`

```
Project Information

  project       devscope
  type          Go

REQUIREMENTS
  Go         >= 1.22.0   (go.mod)
```

### `devscope check`

```
Checking project environment...

✓ Go  1.22.0

Check Summary

  Requirement      Status     Current
  Go               ✓ PASS   1.22.0

  Result: 1 passed
```

### 降级而非伪造

当工具装上了却没被激活(例如 Go 已安装但没选版本),DevScope 会如实说明,而不是假装一切正常:

```
DevScope

DEVELOPMENT
  ○ Go  —  (不可用)
```

"降级而非伪造"贯穿整个工具:取不到的数据一律报告为 `○ SKIP` / 不可用,绝不伪造成 0。

## 工作原理

```
Machine Scope(无项目)              Project Scope(项目内)
  System                             info  → 它需要什么?
  Dev tools                          check → 我满足了吗?
  Services
  Ports
```

数据采集与渲染严格分离:

```
ui → check → project / services / environment → system → versionmanager → domain
```

`domain` 是零依赖的纯数据类型;UI 层永远不判断"这个平台有没有这个指标",它只渲染采集器产出的结果。

## 目录结构

```
cmd/devscope/             CLI 入口
internal/domain/          纯数据类型(零依赖)
internal/system/          系统指标采集(macOS 后端)
internal/services/        服务与监听端口
internal/versionmanager/  版本管理器侦测(asdf)
internal/environment/     开发工具状态侦测
internal/project/         项目识别与需求解析
internal/version/         版本比较
internal/check/           检查引擎
```

## 路线图

- [ ] Python / Docker / LaTeX / PlatformIO 侦测
- [ ] 上述工具的 `check`
- [ ] TUI(Bubble Tea / Lip Gloss)
- [ ] Linux 后端
- [ ] `devscope.yaml` 显式覆盖

## 许可证

MIT,详见 [LICENSE](LICENSE)。

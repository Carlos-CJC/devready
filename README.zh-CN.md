# DevReady

[![CI](https://github.com/Carlos-CJC/devready/actions/workflows/ci.yml/badge.svg)](https://github.com/Carlos-CJC/devready/actions/workflows/ci.yml)

> **这台机器,现在能不能直接跑起这个项目?** —— DevReady 只回答这一个问题。

[English](README.md) | **中文**

> ⚠️ **早期阶段。** 就绪检查(Project Scope)与机器面板(Machine Scope,macOS)均已可用,详见[状态](#状态)。

## 这是什么

DevReady **不是** `btop`、`lazygit` 或 `lazydocker` 的替代品。那些工具告诉你这台机器正在忙什么;DevReady 回答的是一个更窄、更可行动的问题:

> 这个项目就摆在我面前 —— 我能在这台机器上跑起来吗?还是缺了点什么?

`cd` 进一个项目,它会读出项目需要什么,对照这台机器实际装了、并且**当前生效**的环境逐项核对,然后告诉你到底能不能直接开跑。走出项目目录,它就改为展示这台机器的状态 —— 也就是支撑那个判断的背景。

它根据**当前所在目录**自动在两种 Scope 间切换:

- **Project Scope** —— 在项目里。自动识别项目类型,查看它需要什么(`info`),并给出判断(`check`)。
- **Machine Scope** —— 不在项目里。`devready` 展示 CPU、内存、磁盘、网络、已装开发工具、服务与监听端口 —— 这台机器上有什么,随时等着和一个项目对账。

> 说明:CLI 输出目前为**中文**,英文输出 / i18n 计划中。

## 状态

早期、持续开发中。目前可用的部分:

| 模块 | 状态 |
|---|---|
| CLI:`devready` / `info` / `check` / `port` | ✅ |
| Machine Scope:CPU / 内存 / 磁盘 / 网络(macOS) | ✅ |
| Machine Scope:开发工具(Git、Go、Python、Docker、Docker Compose、LaTeX)+ 服务 | ✅ |
| `devready port`:经 `lsof` 列出监听端口 | ✅ |
| Go:项目识别(`go.mod`)+ 版本需求解析 | ✅ |
| Go:环境侦测(已装/激活版本,识别 asdf) | ✅ |
| Python:项目识别(`pyproject.toml` / `requirements.txt`)+ 环境标记 | ✅ |
| Python:环境侦测(识别 pyenv)+ 版本检查 | ✅ |
| Docker / Docker Compose:项目识别(Dockerfile / compose)+ 可用性检查 | ✅ |
| LaTeX:侦测(xelatex) | ✅ |
| PlatformIO:项目识别(`platformio.ini`)+ 可用性检查 | ✅ |
| Go / Python 版本检查 → PASS/FAIL/WARN/SKIP + 汇总 | ✅ |
| 实时面板(TUI,Bubble Tea / Lip Gloss) | ✅ |
| Linux 后端(当前指标诚实降级为 SKIP) | ⬜ 计划中 |

## 安装

需要 Go 1.22+。

```bash
go install github.com/Carlos-CJC/devready/cmd/devready@latest
```

或从源码构建:

```bash
git clone https://github.com/Carlos-CJC/devready
cd devready
go build -o devready ./cmd/devready
```

## 用法

```bash
devready          # 机器实时面板(每 2 秒刷新,按 q 退出)
devready info     # 这个项目需要什么?(不执行检查)
devready check    # 这台机器能不能跑起这个项目?给出判断
devready port     # 监听端口(macOS,经 lsof)
```

### `devready check` —— 核心判断

在项目里,这就是这个工具的落点:读出项目要求,逐项判断这台机器是否满足。

```
Checking project environment...

✓ Go  1.22.0

Check Summary

  Requirement      Status     Current
  Go               ✓ PASS   1.22.0

  Result: 1 passed
```

Python、Docker、PlatformIO 项目同样支持。需求只声明"需要某工具"而未给版本时,
如实降级为 `⚠ WARN` 而非伪造通过:

```
Checking project environment...

✗ Python  3.9.6  (低于要求的 3.11;已装 3.12.0 但未激活)
✓ Docker  29.8.1

Check Summary

  Requirement      Status     Current
  Python           ✗ FAIL   3.9.6
  Docker           ✓ PASS   29.8.1

  Result: 1 passed · 1 failed
```

### `devready info`

```
Project Information

  project       devready
  type          Go

REQUIREMENTS
  Go         >= 1.22.0   (go.mod)
```

### `devready` —— 机器面板

终端内每 2 秒自动重采并刷新,按 `q`(或 `esc` / `Ctrl-C`)退出,按 `r` 立即刷新。
输出被管道或重定向时(非终端),自动降级为单次快照的纯文本。

```
DevReady  my-mac

SYSTEM

  CPU     Apple M4  10 核
          ███░░░░░░░░░░░░░░░░░░░░░  13%
  LOAD    2.34  1.71  1.53

  MEMORY  9.2 GB / 16.0 GB
          ██████████████░░░░░░░░░░  58%

  STORAGE
  System               17.1 GB / 228 GB   █░░░░░░░░░░░░  7%
  /Volumes/Data        2.1 GB / 931 GB    ░░░░░░░░░░░░░  <1%

  NETWORK
  en1                  ↓ 18.9 KB/s   ↑ 99.9 KB/s

DEVELOPMENT
  ✓ Git             2.54.0  (system)
  ✓ Go              1.22.0  (asdf)
  ✓ Python          3.12.0  (pyenv)
  ✓ Docker          29.8.1  (system)
  ✓ Docker Compose  2.24.0  (system)
  ✓ LaTeX           XeLaTeX  (system)

SERVICES
  ○ sshd       未监听
  ○ Docker     未运行
```

### 降级而非伪造

当工具装上了却没被激活(例如 Go 已安装但没选版本),DevReady 会如实说明,而不是假装一切正常:

```
DevReady

DEVELOPMENT
  ○ Go  —  (不可用)
```

"降级而非伪造"贯穿整个工具:取不到的数据一律报告为 `○ SKIP` / 不可用,绝不伪造成 0。
一个不可信的就绪判断,比没有判断更糟。

## 工作原理

```
Machine Scope(无项目)              Project Scope(项目内)
  System                             info  → 它需要什么?
  Dev tools                          check → 这台机器跑得起来吗?
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
cmd/devready/             CLI 入口
internal/domain/          纯数据类型(零依赖)
internal/system/          系统指标采集(macOS 后端)
internal/services/        服务与监听端口
internal/versionmanager/  版本管理器侦测(asdf、pyenv)
internal/environment/     开发工具状态侦测
internal/project/         项目识别与需求解析
internal/version/         版本比较
internal/check/           检查引擎(就绪判断)
internal/ui/              渲染与实时面板(唯一依赖 Bubble Tea / Lip Gloss 的层)
```

## 路线图

- [ ] Linux 后端
- [ ] `devready.yaml` 显式覆盖
- [ ] Node / Rust 项目识别与 `check`
- [ ] compose `services` / `ports` 解析(需引入 YAML 解析)

## 许可证

MIT,详见 [LICENSE](LICENSE)。

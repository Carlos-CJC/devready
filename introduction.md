# DevScope

> 一个本地开发机的终端状态面板与项目环境检查工具。
> 在不在项目里时看这台机器;进入项目后,看这个项目需要什么、当前是否满足。

---

# 0. 设计决策(本文档的定调)

以下四条是本设计明确锁定的边界,后续所有章节都服从它们:

1. **只做本机。** DevScope 监控**运行它的那台机器**,不涉及任何远程 / SSH / 网络穿透。想看某台服务器的状态,就把 DevScope 装到那台服务器上运行。
   - 理由:引入远程就把工具从"一个二进制"变成"一个需要部署拓扑的系统",复杂度指数上升,收益却只是省一次 SSH 登录。
2. **不做传感器采集。** 不采集温度、功耗、频率、电压,不调用 `powermetrics`,不使用 `IOReport` / `IOHID` 等私有接口读取 GPU 利用率。
   - 理由:Apple Silicon 上这些数据没有公开、稳定的用户态 API,免 root 方案依赖私有接口、随时可能被系统更新破坏。**不为一个随时会坏的功能支付长期维护成本。**
3. **不做 GPU 指标(第一版)。** GPU 利用率与显存同上条理由,全部推迟,且不预留半成品接口。
4. **数据层与 UI 层严格分离。** 见第 2、16 章。这是本项目最重要的架构原则。

> 结论:DevScope 只采集**标准、稳定、跨平台可得**的指标(CPU 利用率、内存、磁盘、网络、进程/端口、开发工具版本),采集不到的诚实标记为 SKIP,而不是伪造或强行读取。

---

# 1. 产品定位

DevScope 不是 `btop`、`lazygit` 或 `lazydocker` 的替代品。

它解决的是另一个问题:

> **快速了解当前这台开发机处于什么状态,并在进入项目后,检查当前环境是否满足项目要求。**

它根据**当前所在目录**自动进入两种状态:

```text
┌──────────────────────┐
│  未进入项目           │
│  Machine Scope       │
│  看这台机器           │
└──────────┬───────────┘
           │  cd 进入项目目录
┌──────────▼───────────┐
│  已进入项目           │
│  Project Scope       │
│  看项目环境           │
└──────────────────────┘
```

DevScope 运行在它要监控的机器上,本机即目标机。

---

# 2. 设计模式(本项目采用的模式)

这是全项目一致性的来源,新增任何模块都遵循这五条:

### 2.1 Collector 模式(数据采集)

每个指标源实现统一接口,返回**值 + 状态**,而不是裸值或 panic:

```go
type MetricStatus int

const (
    MetricOK          MetricStatus = iota // 采集成功
    MetricUnavailable                     // 平台不支持 / 无法采集 → 映射为 SKIP
    MetricError                           // 采集出错 → 映射为 WARN/FAIL
)

type Metric[T any] struct {
    Value  T
    Status MetricStatus
    Err    error
}

type Collector[T any] interface {
    Name() string
    Collect(ctx context.Context) Metric[T]
}
```

**为什么重要:** UI 永远不判断"这个平台有没有这个数据",它只渲染 `Metric`。采集不到就是 `MetricUnavailable`,自然落到 SKIP。

### 2.2 Strategy 模式(平台后端)

同一指标在不同操作系统用不同实现,通过 Go 构建标签选择,不做运行期 `if runtime.GOOS`:

```text
internal/system/cpu_darwin.go    //go:build darwin
internal/system/cpu_linux.go     //go:build linux
```

### 2.3 Registry 模式(检查器)

`check` 引擎不写死检查项。每种需求类型注册一个 `Checker`,引擎遍历项目的 `Requirement`,按 `Kind` 查注册表分派(见第 11 章)。

### 2.4 Command 模式(CLI)

`devscope` 的每个子命令是一个独立入口,共享同一套 domain/collector 层(见第 14、15 章)。

### 2.5 Degrade-not-fake(降级而非伪造)

贯穿全局的不变式:**取不到的数据标记为 SKIP/Unavailable,绝不显示为 0 或"正常"。** 这是本工具可信度的底线。

---

# 3. 两种核心状态

## State 1:Machine Scope

当前目录不属于可识别项目。执行:

```bash
devscope
```

展示 CPU、内存、磁盘、网络、已安装开发环境、服务、监听端口。

回答:**"这台机器现在怎么样?"**

## State 2:Project Scope

当前目录属于一个项目。执行:

```bash
devscope info     # 这个项目需要什么?
devscope check    # 我现在满足了吗?
```

---

# 4. Machine Scope —— 全局 Dashboard

执行 `devscope`,目标 UI:

```text
╭─ MacBook Air ──────────────────────────────────────╮
│                                                    │
│  SYSTEM                                            │
│                                                    │
│  CPU     Apple M4                                  │
│          ███████░░░░░░░░░░░░░░░░░  18%             │
│  LOAD    0.42  0.51  0.60                          │
│                                                    │
│  MEMORY  8.9 / 16 GB                               │
│          ███████████████░░░░░░░░░  56%             │
│                                                    │
│  STORAGE                                           │
│  System  67 / 228 GB                               │
│          ███████░░░░░░░░░░░░░░░░░  30%             │
│  Data    2.3 / 931 GB                              │
│          ░░░░░░░░░░░░░░░░░░░░░░░  <1%              │
│                                                    │
│  NETWORK                                           │
│  en0      ↓ 12 KB/s    ↑ 3 KB/s                    │
│                                                    │
│  DEVELOPMENT                                       │
│  ✓ Git          2.51                               │
│  ✓ Go           1.22                               │
│  ✓ Python       3.12                               │
│  ✓ Docker       29.1                               │
│  ✓ LaTeX        XeLaTeX                            │
│  ✓ PlatformIO   6.1                                │
│                                                    │
│  SERVICES                                          │
│  ● sshd         :22                                │
│  ● Docker       running                            │
│  ● Tailscale    connected                          │
│                                                    │
╰────────────────────────────────────────────────────╯
```

Dashboard 原则:

* 一眼能看懂,信息密度适中
* 不需要用户操作,默认自动刷新
* 重点突出 CPU / Memory / Storage
* **无温度、无功耗、无 GPU** —— 违反第 0 章的东西一律不上

---

# 5. 系统监控(仅标准指标)

## CPU

```text
CPU   Apple M4
      █████████░░░░░░░░░  42%
```

* total utilization(取自系统标准接口)
* CPU 型号与核心数
* load average

**不做:** 温度、功耗、频率、逐核曲线。

## Memory

```text
MEMORY
8.9 / 16 GB
██████████████░░░░░░░░  56%
```

至少提供 total / used / available / percentage。

## Storage

```text
SYSTEM   67 / 228 GB   30%
DATA      2.3 / 931 GB  <1%
```

遍历挂载点,过滤系统伪文件系统,显示容量与占比。

## Network

```text
NETWORK
en0      ↓ 12 KB/s    ↑ 3 KB/s
```

每张活动接口的实时收发速率(对两次采样求差)。**不做**连接跟踪、不做按进程流量。

---

# 6. Development Environment

Machine Scope 下显示已安装的主要开发工具及版本:

```text
DEVELOPMENT
✓ Git           2.51
✓ GitHub CLI    2.40
✓ Go            1.22
✓ Python        3.12
✓ Docker        29.1
✓ Docker Compose 2.24
✓ LaTeX         XeLaTeX
✓ PlatformIO    6.1
```

**注意:不把 Python 当成一个简单的全局环境。** 项目可能用 pyenv / `.venv` / Poetry / uv / Conda。

Machine Scope 只回答:**"机器上有没有 Python,默认版本是多少。"** 具体项目环境由 Project Scope 判断。

---

# 7. Services 与 Port

## Services

```text
SERVICES
● sshd          :22
● Docker        running
● Tailscale     connected
```

仅检测**本机**服务:进程是否存在、端口是否在监听。"SSH" 在这里只是"本机 sshd 是否在跑",与远程访问无关。

## Port

```bash
devscope port
```

```text
╭─ Listening Ports ───────────────────────────╮
│                                             │
│  :22      sshd                              │
│  :5432    postgres     Docker               │
│  :6379    redis        Docker               │
│  :1883    mosquitto    Docker               │
│  :8080    nginx        Docker               │
│                                             │
╰─────────────────────────────────────────────╯
```

不是替代 `lsof`。目标是把端口信息翻译成:**"这台机器现在有哪些服务在工作?"**

> **权限降级:** 部分平台无法读取其他用户进程的端口归属时,只显示端口与已知的 Docker 映射,归属未知的标记而不报错。

---

# 8. Project Scope —— 项目识别

进入项目目录后自动检测:

```text
PROJECT DETECTED

my-project
Go + Docker
```

识别依据:

```text
go.mod            → Go
pyproject.toml    → Python
requirements.txt  → Python
package.json      → Node
Cargo.toml        → Rust
compose.yml       → Docker Compose
Dockerfile        → Docker
platformio.ini    → PlatformIO
devscope.yaml     → 显式覆盖(最高优先级)
```

---

# 9. `devscope info`

职责:**只展示项目环境要求,不执行检查。**

```text
╭─ Project Information ─────────────────────────────╮
│                                                   │
│  PROJECT       go-stock-platform                  │
│  TYPE          Go + Docker                        │
│                                                   │
│  REQUIREMENTS                                     │
│  Go             >= 1.22                           │
│  Docker         required                          │
│  PostgreSQL     required                          │
│  Redis          required                          │
│                                                   │
│  SERVICES                                         │
│  PostgreSQL     :5432                             │
│  Redis          :6379                             │
│                                                   │
╰───────────────────────────────────────────────────╯
```

它只回答:**"项目要求什么。"**

---

# 10. `devscope check`

这是 DevScope 最重要的功能。

区别于传统的静态打勾列表,采用**逐项实时检查动画 + 最终静态汇总**:

```text
Checking project environment...

✓ Go version            1.22.8
✓ Docker                29.1
✓ Python environment    3.12.0 (.venv)
⠋ Checking PostgreSQL...
```

失败时:

```text
✓ Go version            1.22.8
✓ Docker                29.1
✓ Python environment    3.12.0 (.venv)
✓ PostgreSQL            :5432
✗ Redis                 not running
```

**过程是动态的,结果是静态的。** 完成后不直接结束,而是输出汇总表:

```text
╭─ Check Summary ───────────────────────────────────────╮
│  Requirement       Status       Current               │
│  ───────────────────────────────────────────────────  │
│  Go                ✓ PASS       1.22.8                │
│  Docker            ✓ PASS       29.1                  │
│  Python            ✓ PASS       3.12.0                │
│  PostgreSQL        ✓ PASS       :5432                 │
│  Redis             ✗ FAIL       not running           │
│  Port :8080        ⚠ WARN       already occupied      │
│  ───────────────────────────────────────────────────  │
│  Result: 4 passed · 1 failed · 1 warning              │
╰───────────────────────────────────────────────────────╯
```

这个"动画 → 汇总"的交互,是 DevScope 相对纯 Dashboard 的辨识度所在。

---

# 11. 检查流程与状态

每个检查项的状态机:

```text
Pending → Checking → Passed / Failed / Warning / Skipped
```

统一四态:

| 标记 | 含义 |
|------|------|
| `✓ PASS` | 满足要求 |
| `✗ FAIL` | 明确不满足要求 |
| `⚠ WARN` | 存在潜在问题,但不一定阻止项目运行 |
| `○ SKIP` | 当前条件下无法或无需检查 |

**SKIP 是一等公民**,例如:

```text
GPU temperature    ○ SKIP    unsupported on this Mac
```

不把无法获取的数据伪装成正常(对应 2.5 节)。

核心数据结构:

```go
type Status int

const (
    Pass Status = iota
    Fail
    Warn
    Skip
)

type CheckResult struct {
    Name     string // 检查项名称
    Status   Status
    Required string // 要求,如 ">=1.22"
    Current  string // 实测,如 "1.22.8"
    Message  string // 补充说明
}
```

---

# 12. 项目需求推断规则表

这是业务复杂度的核心,必须显式定义而不是"看着办"。

| 文件 | 推断出的要求 | 规则 | 无法判断时 |
|------|-------------|------|-----------|
| `go.mod` | Go 版本 | 读 `go` 指令作为最低版本 | → SKIP |
| `pyproject.toml` | Python 版本 | 读 `requires-python` | → WARN(python 需要但版本未知) |
| `requirements.txt` | Python | 存在即要求 python,版本未知 | → WARN |
| `.python-version` / `.venv` / `poetry.lock` / `uv.lock` / `environment.yml` | 环境管理器 | 识别 venv/poetry/uv/conda,供 `check` 校验实际生效环境 | → SKIP |
| `package.json` | Node 版本 | 读 `engines.node`;无则只要求 node 存在 | → WARN |
| `Cargo.toml` | Rust 版本 | 读 `rust-version` / edition | → SKIP |
| `Dockerfile` | Docker | 存在即要求 docker | — |
| `compose.yml` | Docker Compose + 服务 + 端口 | 解析 `services` 与 `ports`(含 `HOST:CONTAINER`) | 解析失败 → SKIP 并提示 |
| `platformio.ini` | PlatformIO | 读 boards / frameworks | → SKIP |
| `devscope.yaml` | 全部 | **覆盖以上所有推断** | — |

**项目根判定:** 自当前目录向上查找,遇到第一个含上述标识文件的目录即为项目根;遇到嵌套/多标识文件时,以**最近的**为根,并在输出中提示存在多个候选(monorepo 友好)。

**端口占用检查语义:**

```text
期望端口被期望服务占用   → PASS
期望端口被其他进程占用   → WARN
期望端口无人监听         → FAIL
```

---

# 13. 可选的 `devscope.yaml`

当自动推断无法表达完整需求时,用显式配置覆盖:

```yaml
project: go-stock-platform

requirements:
  go: ">=1.22"
  docker: true

services:
  - postgres
  - redis

ports:
  - 5432
  - 6379
```

存在时优先级最高,完全接管推断结果。

---

# 14. CLI

**第一阶段(本设计交付范围):**

```bash
devscope          # 无项目 → Machine Dashboard;有项目 → 项目概览
devscope info     # 展示项目要求
devscope check    # 逐项检查 + 汇总
devscope port     # 监听端口
```

**后续(仅在确有需求时):**

```bash
devscope doctor   # 环境自诊断
```

非目标命令(明确不做):`clean`、`docker`、`service`、`network` 等一切"替用户操作机器"的命令。

---

# 15. 技术架构

* 语言:**Go**
* TUI:**Bubble Tea**
* 样式:**Lip Gloss**
* CLI:标准库 `flag` + 命令分派表(不引入 cobra,降低依赖)
* 依赖策略:仅 bubbletea / lipgloss;`info` / `check` / `port` 的一次性输出不依赖 TUI,可在无终端环境降级为纯文本

## 目录结构

```text
devscope/
│
├── cmd/
│   └── devscope/
│       └── main.go              # 命令分派入口
│
├── internal/
│   │
│   ├── domain/                  # 纯数据结构,零依赖,被所有层引用
│   │   ├── system.go            #   Metric / CPUInfo / MemInfo / DiskInfo ...
│   │   ├── project.go           #   Project / Requirement
│   │   └── check.go             #   CheckResult / Status
│   │
│   ├── system/                  # 系统指标采集(按平台分文件)
│   │   ├── collector.go         #   接口与聚合
│   │   ├── cpu_darwin.go
│   │   ├── cpu_linux.go
│   │   ├── memory_darwin.go
│   │   ├── memory_linux.go
│   │   ├── disk_darwin.go
│   │   ├── disk_linux.go
│   │   └── network.go
│   │
│   ├── environment/             # 开发工具检测
│   │   ├── detector.go
│   │   ├── go.go
│   │   ├── python.go
│   │   ├── docker.go
│   │   ├── latex.go
│   │   └── platformio.go
│   │
│   ├── project/                 # 项目识别与需求推断(规则表落地处)
│   │   ├── detector.go
│   │   ├── parser.go
│   │   └── requirements.go
│   │
│   ├── services/                # 服务与端口
│   │   ├── services.go
│   │   └── ports.go
│   │
│   ├── check/                   # 检查引擎
│   │   ├── engine.go            #   遍历 requirements、分派 checker
│   │   ├── registry.go          #   Kind → Checker 注册表
│   │   └── checkers/
│   │       ├── version.go
│   │       ├── service.go
│   │       └── port.go
│   │
│   └── ui/                      # 仅此层依赖 Bubble Tea / Lip Gloss
│       ├── dashboard.go
│       ├── project.go
│       ├── check.go             #   逐项动画
│       ├── summary.go           #   汇总表
│       └── spinner.go
│
└── go.mod
```

---

# 16. 分层与数据流

严格单向依赖:`ui → check → project/services/environment → system → domain`。domain 不被任何层反向依赖。

## 系统指标流

```text
平台实现 (cpu_darwin / cpu_linux ...)
        │  Collector[T].Collect()
        ▼
   Metric[T]  (Value + Status)
        │  聚合
        ▼
   SystemStatus
        │
        ▼
   ui.Dashboard  (只渲染,不判断平台)
```

## 项目环境流

```text
项目文件 (go.mod / compose.yml / ...)
        │  project.Detector + Parser
        ▼
   Project { Requirements }
        │                        ▲
        │                        │  devscope.yaml 覆盖
        ▼
   check.Engine  ──分派──►  Registry[Kind] → Checker
        │
        ▼
   []CheckResult
        │
        ▼
   ui.CheckProgress (spinner) → ui.Summary (静态表)
```

**原则:更换 UI 或增加 Linux 支持,都不需要改动业务逻辑。**

## 并发与超时

所有 collector 通过 `errgroup` 并发执行,各自带 `context` 超时(建议单指标 500ms)。任何采集器挂起或失败,只影响自身呈现为 SKIP/WARN,不阻塞整个 Dashboard。

---

# 17. 目标与非目标

## 目标

* **零配置**:`cd` 进任何项目直接 `devscope` 就有结果,不需要先写配置文件
* **快**:一次性命令(`info`/`check`/`port`)冷启动 < 100ms;Dashboard 常驻刷新无卡顿
* **诚实**:取不到就 SKIP,绝不伪造(2.5 节)
* **可移植**:核心逻辑与 OS 解耦,新增平台只加后端文件
* **单文件分发**:一个静态二进制,`go install` 或直接拷贝即可用
* **可单测**:domain / project / check 纯逻辑层不碰系统调用,易测

## 非目标

* 不做远程 / SSH / 多机聚合(第 0 章)
* 不采集温度 / 功耗 / 频率 / 电压
* 不做 GPU 指标
* 不替代 `btop` / `lazygit` / `lazydocker` / `Docker Desktop` / `lsof`
* 不做完整进程管理器或网络监控
* 不自动修改用户环境、不自动删除文件

> 核心原则:**DevScope 只告诉用户"现在是什么状态"和"项目缺什么",不替用户管理机器。**

---

# 18. 风险点

| # | 风险 | 影响 | 缓解 |
|---|------|------|------|
| 1 | **平台数据源差异**:macOS 与 Linux 的 CPU/磁盘/网络接口完全不同 | 部分指标在某平台只能 SKIP | 用 Collector + 构建标签隔离;SKIP 是合法结果,不是缺陷 |
| 2 | **需求推断不准**:monorepo 嵌套、`package.json` 无 `engines`、compose 含变量/override | `info`/`check` 给出错误要求 | 第 12 章规则表 + 无法判断时降级为 WARN/SKIP + `devscope.yaml` 兜底 |
| 3 | **端口/服务归属需权限**:macOS 读取其他用户进程端口受限 | 端口归属显示不全 | 归属未知时只显示端口与 Docker 映射,不报错 |
| 4 | **冷启动性能**:每次 `exec` git/go/docker 有开销 | 命令变慢 | 并发采集 + 超时;已安装版本号做进程内缓存(TTL 数分钟) |
| 5 | **自动刷新开销**:Dashboard 周期调用外部命令 | 占用 CPU | 轻指标高频、重指标低频;版本类信息缓存 |
| 6 | **范围蔓延**:想加回传感器 / 远程 / GPU | 拖垮 MVP | 第 0 章 + 第 17 章非目标作为硬约束,新增项必须显式改文档 |
| 7 | **依赖膨胀**:bubbletea 依赖树较大 | 二进制变大、升级耦合 | 一次性命令不依赖 TUI 层;锁定版本 |

---

# 19. 计划(路线图)

按"先引擎、后 UI"排序,每阶段可独立交付、可验证。

### Phase 0 — 骨架与系统面板
* `cmd` 分派、`domain` 类型、`system` 接口与 macOS 后端(CPU/内存/磁盘/网络)
* 交付:`devscope` 静态渲染一次,输出正确
* 验收:数据与系统工具(如 `top`/`df`)对得上

### Phase 1 — 环境与服务
* `environment`(Git/Go/Python/Docker 等版本检测)、`services`、`ports`
* 交付:Dashboard 补齐 DEVELOPMENT / SERVICES 区块
* 验收:各版本号与手动执行一致

### Phase 2 — 项目识别与 `info`
* `project.Detector` + `parser` + 第 12 章规则表
* 交付:`devscope info` 正确输出要求
* 验收:对 Go / Python / Node / Docker / PlatformIO 样例项目各验证一次

### Phase 3 — `check` 引擎与交互
* `check.Engine` + registry + checkers;`ui.CheckProgress`(spinner)+ `ui.Summary`
* 交付:逐项动画 → 静态汇总表
* 验收:构造 PASS/FAIL/WARN/SKIP 四种场景各一例

### Phase 4 — 打磨
* Dashboard 自动刷新、`devscope.yaml` 覆盖、文本降级输出、安装文档
* 验收:冷启动 < 100ms;无终端环境下不崩溃

> 传感器 / GPU / 远程:不在任何阶段。如未来确需,须先更新第 0 章。

---

# 20. 最终定义

DevScope 可用三个词概括:**Scope → Check → Understand**。

```text
                 DEV SCOPE
                     │
          ┌──────────┴──────────┐
          │                     │
     MACHINE SCOPE          PROJECT SCOPE
          │                     │
     ┌────┼────┐           ┌────┴────┐
     │    │    │           │         │
   System Dev  Services    Info     Check
     │    │    │             │         │
     ▼    ▼    ▼             ▼         ▼
    CPU  Go   Ports       Requirements Results
    RAM  Py   Docker
    Disk ...
```

**一句话定义:**

> **DevScope 是一个本机开发环境 TUI:不在项目里看这台机器,进入项目后看环境;`info` 告诉你项目需要什么,`check` 用逐项动画检查并汇总告诉你当前环境是否满足要求。**

交互样例:

```text
⠋ Checking Go version...
✓ Go 1.22.8

⠋ Checking Python environment...
✓ Python 3.12.0 (.venv)

⠋ Checking PostgreSQL...
✓ PostgreSQL :5432

⠋ Checking Redis...
✗ Redis not running

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Environment Check

Go             ✓ PASS
Python         ✓ PASS
PostgreSQL     ✓ PASS
Redis          ✗ FAIL

3 passed · 1 failed
```

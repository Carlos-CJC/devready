# DevReady

[![CI](https://github.com/Carlos-CJC/devready/actions/workflows/ci.yml/badge.svg)](https://github.com/Carlos-CJC/devready/actions/workflows/ci.yml)

> **Can this machine run this project, right now?** DevReady answers that one question.

**English** | [中文](README.zh-CN.md)

> ⚠️ **Early stage.** The readiness check (Project Scope) and the machine panel (Machine Scope, macOS) work today. See [Status](#status).

## What it is

DevReady is **not** a replacement for `btop`, `lazygit`, or `lazydocker`. Those show you what a machine is doing. DevReady answers a narrower, more actionable question:

> This project is sitting in front of me — can I run it here, or is something missing?

`cd` into a project and it reads what the project needs, checks it against what's actually installed and active on this machine, and tells you whether you're good to go. Step outside a project and it shows the machine's state instead — the context behind that verdict.

It picks one of two scopes based on your current directory:

- **Project Scope** — you're inside a project. DevReady detects the project type, then lets you inspect what it needs (`info`) and get the verdict (`check`).
- **Machine Scope** — you're not in a project. `devready` shows CPU, memory, disk, network, installed dev tools, services and listening ports — what's on the machine, ready to be judged against a project.

> Note: CLI output is currently written in **Chinese**. English output / i18n is planned.

## Status

Early and in progress. What works today:

| Area | Status |
|---|---|
| CLI: `devready` / `info` / `check` / `port` | ✅ |
| Machine Scope: CPU / memory / disk / network (macOS) | ✅ |
| Machine Scope: dev tools (Git, Go, Python, Docker, Docker Compose, LaTeX) + services | ✅ |
| `devready port`: listening ports via `lsof` | ✅ |
| Go: project detection (`go.mod`) + version requirement parsing | ✅ |
| Go: environment detection (installed / active version, asdf-aware) | ✅ |
| Python: project detection (`pyproject.toml` / `requirements.txt`) + env markers | ✅ |
| Python: environment detection (pyenv-aware) + version check | ✅ |
| Docker / Docker Compose: project detection (Dockerfile / compose) + availability check | ✅ |
| LaTeX: detection (xelatex) | ✅ |
| PlatformIO: project detection (`platformio.ini`) + availability check | ✅ |
| Go / Python version check → PASS/FAIL/WARN/SKIP + summary | ✅ |
| Live panel (TUI, Bubble Tea / Lip Gloss) | ✅ |
| Linux backend (metrics degrade to SKIP today) | ⬜ planned |

## Install

Requires Go 1.22+.

```bash
go install github.com/Carlos-CJC/devready/cmd/devready@latest
```

Or from source:

```bash
git clone https://github.com/Carlos-CJC/devready
cd devready
go build -o devready ./cmd/devready
```

## Usage

```bash
devready          # live machine panel (refreshes every 2s, q to quit)
devready info     # what does this project need? (does not check)
devready check    # can this machine run this project? the verdict
devready port     # listening ports (macOS, via lsof)
```

### `devready check` — the verdict

Inside a project, this is the point of the tool: it reads the project's requirements and reports whether this machine satisfies each one.

```
Checking project environment...

✓ Go  1.22.0

Check Summary

  Requirement      Status     Current
  Go               ✓ PASS   1.22.0

  Result: 1 passed
```

Python, Docker and PlatformIO projects work the same way. When a requirement only says "this tool is needed" without a version, it degrades to `⚠ WARN` rather than faking a pass:

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

### `devready` — the machine panel

Inside a terminal it re-collects and redraws every 2 seconds; press `q` (or `esc` / `Ctrl-C`) to quit, `r` to refresh immediately. When stdout is piped or redirected (not a terminal), it degrades to a single plain-text snapshot.

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

### Degrade, don't fake

If a tool is installed but not activated (e.g. Go is present but no version is selected), DevReady says so instead of pretending everything's fine:

```
DevReady

DEVELOPMENT
  ○ Go  —  (不可用)
```

This "degrade, don't fake" rule runs through the whole tool: data that can't be collected is reported as `○ SKIP` / unavailable, never as a fabricated zero. A readiness verdict you can't trust is worse than no verdict.

## How it works

```
Machine Scope (no project)          Project Scope (inside a project)
  System                              info  → what does it need?
  Dev tools                           check → can this machine run it?
  Services
  Ports
```

Data collection and rendering are strictly separated:

```
ui → check → project / services / environment → system → versionmanager → domain
```

`domain` is plain data with no dependencies; UI code never decides whether a platform supports a metric — it only renders what the collectors produced.

## Project layout

```
cmd/devready/             CLI entry point
internal/domain/          plain data types (no dependencies)
internal/system/          system metric collectors (macOS backend)
internal/services/        services & listening ports
internal/versionmanager/  version-manager detection (asdf, pyenv)
internal/environment/     dev-tool state detection
internal/project/         project detection & requirement parsing
internal/version/         version comparison
internal/check/           check engine (the readiness verdict)
internal/ui/              rendering & live panel (only layer using Bubble Tea / Lip Gloss)
```

## Roadmap

- [ ] Linux backend
- [ ] `devready.yaml` explicit overrides
- [ ] Node / Rust project detection + `check`
- [ ] compose `services` / `ports` parsing (needs a YAML parser)

## License

MIT — see [LICENSE](LICENSE).

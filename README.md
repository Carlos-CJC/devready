# DevScope

[![CI](https://github.com/Carlos-CJC/devscope/actions/workflows/ci.yml/badge.svg)](https://github.com/Carlos-CJC/devscope/actions/workflows/ci.yml)

> A terminal panel for your local dev machine — see the machine when you're not inside a project, and see whether a project's environment is ready when you are.

**English** | [中文](README.zh-CN.md)

> ⚠️ **Early stage.** The Machine Scope panel (macOS) and the Go check in Project Scope work today. See [Status](#status).

## What it is

DevScope is **not** a replacement for `btop`, `lazygit`, or `lazydocker`. It answers a different question:

> What state is this dev machine in right now — and once I'm inside a project, does the environment satisfy what that project needs?

It picks one of two scopes based on your current directory:

- **Machine Scope** — you're not in a project. `devscope` shows CPU, memory, disk, network, installed dev tools, services and listening ports.
- **Project Scope** — you're inside a project. DevScope detects the project type, then lets you inspect requirements (`info`) and verify the environment (`check`).

> Note: CLI output is currently written in **Chinese**. English output / i18n is planned.

## Status

Early and in progress. What works today:

| Area | Status |
|---|---|
| CLI: `devscope` / `info` / `check` / `port` | ✅ |
| Machine Scope: CPU / memory / disk / network (macOS) | ✅ |
| Machine Scope: dev tools (Git, Go, Docker) + services (sshd, Docker) | ✅ |
| `devscope port`: listening ports via `lsof` | ✅ |
| Go: project detection (`go.mod`) + version requirement parsing | ✅ |
| Go: environment detection (installed / active version, asdf-aware) | ✅ |
| Go: version check → PASS/FAIL/WARN/SKIP + summary | ✅ |
| Other tools (Python / Docker / LaTeX / PlatformIO) detection | ⬜ planned |
| TUI (Bubble Tea / Lip Gloss) | ⬜ planned |
| Linux backend (metrics degrade to SKIP today) | ⬜ planned |

## Install

Requires Go 1.22+.

```bash
go install github.com/Carlos-CJC/devscope/cmd/devscope@latest
```

Or from source:

```bash
git clone https://github.com/Carlos-CJC/devscope
cd devscope
go build -o devscope ./cmd/devscope
```

## Usage

```bash
devscope          # machine / project overview
devscope info     # show project requirements (does not check)
devscope check    # check the environment against requirements
devscope port     # listening ports (macOS, via lsof)
```

### `devscope` — Machine / Project Scope

```
DevScope  my-mac

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

### Degrade, don't fake

If a tool is installed but not activated (e.g. Go is present but no version is selected), DevScope says so instead of pretending everything's fine:

```
DevScope

DEVELOPMENT
  ○ Go  —  (不可用)
```

This "degrade, don't fake" rule runs through the whole tool: data that can't be collected is reported as `○ SKIP` / unavailable, never as a fabricated zero.

## How it works

```
Machine Scope (no project)          Project Scope (inside a project)
  System                              info  → what does it need?
  Dev tools                           check → do I satisfy it?
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
cmd/devscope/             CLI entry point
internal/domain/          plain data types (no dependencies)
internal/system/          system metric collectors (macOS backend)
internal/services/        services & listening ports
internal/versionmanager/  version-manager detection (asdf)
internal/environment/     dev-tool state detection
internal/project/         project detection & requirement parsing
internal/version/         version comparison
internal/check/           check engine
```

## Roadmap

- [ ] Python / Docker / LaTeX / PlatformIO detection
- [ ] `check` for the tools above
- [ ] TUI (Bubble Tea / Lip Gloss)
- [ ] Linux backend
- [ ] `devscope.yaml` explicit overrides

## License

MIT — see [LICENSE](LICENSE).

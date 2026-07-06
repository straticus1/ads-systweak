# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

After Dark System Tools (`ads-systweak`) — a macOS optimization utility exposing 80+ system tweaks (mostly `defaults write` style settings Apple doesn't surface in System Preferences). Go, with a Fyne GUI and a Cobra CLI. Uses a Terraform-style declarative workflow: **stage → plan → apply → rollback**. Desired state is stored in `~/.ads-systweak.json`.

## Commands

```bash
go build -o ads-systweak .    # build for current platform
make build                     # universal binary via lipo into bin/ (arm64 + amd64)
make test                      # go test ./... -v
make clean

go test ./pkg/tweaks/          # run the one existing test package (registry_test.go)

# Run
./ads-systweak                 # GUI mode (default)
./ads-systweak list            # CLI: list tweaks
./ads-systweak preset stage developer-mode
./ads-systweak plan
./ads-systweak apply
./ads-systweak status <tweak-id>
./ads-systweak force-apply <tweak-id>
./ads-systweak preapprove <category>
```

## Architecture

- `main.go` → `cmd/` (Cobra): `root.go` dispatches between `cli.go` (list/plan/apply/preset/status subcommands) and `ui.go` (launches the GUI).
- `pkg/tweaks/` — the core domain. `tweak.go` defines the `Tweak` interface; `registry.go` is the catalog of all tweaks (add new tweaks here, following the existing interface pattern); `implementations.go` contains the concrete apply/revert/detect logic; `presets.go` defines the bundles (developer-mode, privacy-hardening, maximum-performance); `shell.go` executes the underlying `defaults`/system commands.
- `pkg/state/config.go` — desired-state persistence (`~/.ads-systweak.json`, including `pre_approved` categories).
- `pkg/backup/` — rollback support (all tweaks are reversible to macOS defaults).
- `pkg/api/` — an API server plus Sysinternals-style modules (`autoruns.go`, `procexp.go`, `tcpview.go`, `defaults.go`, `reliability.go`).
- `ui/` — Fyne GUI (app, layout, theme).

Key deps: `fyne.io/fyne/v2`, `github.com/spf13/cobra`, `howett.net/plist`. Privileged tweaks prompt for sudo. `TODO.md` tracks planned work.

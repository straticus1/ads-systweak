# macOS Safety Overhaul Implementation Plan

> **For Codex:** Use `${SUPERPOWERS_SKILLS_ROOT}/skills/collaboration/executing-plans/SKILL.md` to implement this plan task-by-task.

**Goal:** Make ads-systweak fail closed, restore exact supported preference values, reject unsafe web mutations, and accurately represent modern macOS capabilities.

**Architecture:** Preserve the Go CLI and Fyne application while extracting testable execution, planning, backup, and validation boundaries. Replace user-derived shell strings with argument-vector execution, add typed scalar backup receipts, propagate probe failures, and centralize mutation safety rules shared by CLI, GUI, and web handlers.

**Tech Stack:** Go 1.24+, Cobra, Fyne, `howett.net/plist`, standard-library HTTP and process execution.

---

### Task 1: Safe command execution boundary

**Files:**
- Create: `pkg/execx/runner.go`
- Create: `pkg/execx/runner_test.go`
- Modify: `pkg/tweaks/shell.go`

**Steps:**
1. Write failing tests for direct argument execution, captured stderr, and AppleScript-safe privileged execution.
2. Run `go test ./pkg/execx ./pkg/tweaks` and confirm the new API is missing.
3. Implement a small injectable runner using `exec.Command` argument vectors; retain shell execution only for static built-in command definitions.
4. Make privileged execution pass AppleScript as an argument instead of nesting it in `sh -c`.
5. Run the focused tests and commit `refactor: add safe command execution boundary`.

### Task 2: Typed, non-destructive defaults backups

**Files:**
- Modify: `pkg/backup/backup.go`
- Create: `pkg/backup/backup_test.go`
- Modify: `pkg/tweaks/implementations.go`
- Create: `pkg/tweaks/implementations_test.go`

**Steps:**
1. Write failing tests proving first-write preservation, scalar type preservation, absent-key restoration, failed-backup mutation prevention, and consumed receipts.
2. Run focused tests and verify each regression fails for the expected reason.
3. Store typed scalar backup receipts atomically with mode `0600`; reject unsupported complex values before mutation.
4. Make `DefaultsTweak.Apply` require a successful backup and make `Revert` restore the receipt instead of overwriting it and deleting blindly.
5. Propagate restart and backup errors.
6. Run focused tests and commit `fix: make defaults rollback typed and non-destructive`.

### Task 3: Fail-closed probes and plans

**Files:**
- Modify: `pkg/tweaks/tweak.go`
- Modify: `pkg/tweaks/implementations.go`
- Create: `pkg/tweaks/plan.go`
- Create: `pkg/tweaks/plan_test.go`
- Modify: `cmd/cli.go`
- Modify: `ui/layout.go`
- Modify: `pkg/api/tweaks.go`

**Steps:**
1. Write failing tests for missing preference keys versus permission/command failures and for plans that retain an `unknown` state.
2. Implement structured probe states: applied, off, unsupported, permission denied, and error.
3. Centralize plan generation and prevent apply/revert when actual state is unknown.
4. Update CLI, GUI, and API representations to expose errors rather than treating them as off.
5. Run focused and full tests; commit `fix: fail closed when tweak state is unknown`.

### Task 4: Exact restoration for destructive command tweaks

**Files:**
- Create: `pkg/tweaks/values.go`
- Create: `pkg/tweaks/values_test.go`
- Modify: `pkg/tweaks/registry.go`

**Steps:**
1. Write failing pure-function tests for adding/removing individual NVRAM boot arguments without disturbing neighbors.
2. Write failing tests for capturing/restoring DNS, pmset, sysctl, and multi-key preference values.
3. Add value-backed command tweaks that snapshot original values and restore them exactly.
4. Convert boot arguments, DNS, low-power mode, delayed ACK, Time Machine throttling, and lid/sleep settings.
5. Reclassify one-shot actions so they cannot pretend to have reversible state.
6. Run focused tests and commit `fix: preserve original state for command tweaks`.

### Task 5: Secure and atomic local state

**Files:**
- Modify: `pkg/state/config.go`
- Create: `pkg/state/config_test.go`
- Modify: `pkg/analytics/analytics.go`

**Steps:**
1. Write failing tests for `0600` permissions, atomic replacement, malformed-config errors, and preserved maps.
2. Implement temp-file, sync, chmod, and rename persistence.
3. Stop callers from discarding load/save errors.
4. Store analytics identifiers with `0600` mode and avoid hostname-derived identifiers.
5. Run focused tests and commit `fix: persist local state atomically and privately`.

### Task 6: Harden the localhost web API

**Files:**
- Create: `pkg/api/security.go`
- Create: `pkg/api/security_test.go`
- Modify: `pkg/api/server.go`
- Modify: `pkg/api/defaults.go`
- Modify: `pkg/api/web/app.js`

**Steps:**
1. Write failing handler tests for foreign origins, missing CSRF headers, malicious numeric/string inputs, oversized bodies, and valid same-origin requests.
2. Add same-origin mutation middleware, a per-process CSRF token endpoint/header, body limits, JSON content-type enforcement, and security headers.
3. Replace all user-derived `sh -c` defaults operations with direct argument-vector calls and strict scalar parsing.
4. Route browser mutations through one JavaScript request helper that supplies the token.
5. Run API tests and commit `fix: harden localhost mutation API`.

### Task 7: GUI safety and functional Defaults editor

**Files:**
- Modify: `ui/layout.go`
- Create: `ui/model.go`
- Create: `ui/model_test.go`

**Steps:**
1. Extract and test view models for categories, filtering, staged plans, risk summaries, and typed defaults rows.
2. Add the Hidden CLI category, wire domain filtering, and parse exported plist data into real sorted keys and typed values.
3. Add edit/delete dialogs for supported scalar values using the safe execution boundary.
4. Require a plan confirmation and an additional explicit confirmation for high-risk changes.
5. Move status probing and apply work off the UI event path and refresh results afterward.
6. Run UI model and full tests; commit `fix: make GUI planning and defaults editing safe`.

### Task 8: Modern macOS compatibility and truthful catalog

**Files:**
- Modify: `cmd/root.go`
- Modify: `pkg/api/autoruns.go`
- Modify: `pkg/tweaks/registry.go`
- Modify: `pkg/tweaks/registry_test.go`
- Modify: `README.md`

**Steps:**
1. Write failing tests for runtime OS checks, command/path requirements, and modern extension inventory parsing.
2. Replace environment-based GOOS detection with `runtime.GOOS` and add macOS version/build facts.
3. Replace removed `kextstat` collection with `kmutil`/`systemextensionsctl` capability-aware collection.
4. Mark deprecated `airport` support appropriately and prefer `wdutil`; correct misleading descriptions and risk levels.
5. Add registry validation for IDs, categories, risk, required commands, reversibility, and compatibility metadata.
6. Update README claims and compatibility documentation.
7. Run full tests, vet, build, and formatting checks; commit `fix: align catalog with modern macOS`.

### Task 9: End-to-end verification and release handoff

**Files:**
- Create: `docs/SAFETY.md`
- Modify: `TODO.md`

**Steps:**
1. Document safety guarantees, unsupported cases, privilege boundaries, backups, and manual macOS VM verification.
2. Run `gofmt` and `git diff --check`.
3. Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...` with a writable temporary Go cache.
4. Exercise read-only CLI commands on macOS and verify no mutation occurs.
5. Review the final diff against every task and record any hardware/VM-only checks that remain.
6. Commit `docs: document tweak safety model` and present the branch for merge.

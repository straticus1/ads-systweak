# Guarded Danger Zone Implementation Plan

> **For Codex:** Use `${SUPERPOWERS_SKILLS_ROOT}/skills/collaboration/executing-plans/SKILL.md` to implement this plan task-by-task.

**Goal:** Hide all High-risk tweaks from the normal native and web interfaces until the user completes a session-only warning and typed-confirmation flow.

**Architecture:** Put risk filtering and acknowledgement validation in `pkg/tweaks` so both interfaces share one policy. The native GUI holds only an in-memory unlocked flag and blocks plans containing hidden High-risk work. The web server omits High-risk tweaks from its ordinary endpoint and requires a random in-memory capability header for dangerous listing, apply, and revert requests.

**Tech Stack:** Go 1.24, Fyne v2, Cobra HTTP server, embedded HTML/CSS/JavaScript, Go `httptest`.

---

### Task 1: Shared Danger Zone Policy

**Files:**
- Create: `pkg/tweaks/danger.go`
- Create: `pkg/tweaks/danger_test.go`
- Modify: `ui/model.go`
- Modify: `ui/model_test.go`

**Step 1: Write failing policy tests**

Test that `VisibleTweaks(registry, false)` excludes `RiskHigh`, that `VisibleTweaks(registry, true)` includes every tweak, and that `DangerousTweaks` returns only High-risk entries without changing order. Test that acknowledgement validation requires both the checked boolean and the exact phrase `I KNOW WHAT I AM DOING`.

**Step 2: Run the tests and verify RED**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/tweaks -run 'TestVisibleTweaks|TestDangerousTweaks|TestValidateDangerUnlock' -count=1`

Expected: FAIL because the policy functions do not exist.

**Step 3: Implement the minimal shared policy**

Add `DangerUnlockPhrase`, `VisibleTweaks`, `DangerousTweaks`, and `ValidateDangerUnlock`. Return fresh slices so callers cannot mutate the registry.

**Step 4: Add a failing native-plan guard test**

Test `CanExecutePlan(plan, unlocked)` in `ui/model_test.go`: a plan containing an actionable High-risk item must return an error while locked and succeed while unlocked; Low/Medium-only and already-blocked High-risk items must not be treated as executable hidden work.

**Step 5: Implement and verify GREEN**

Add `CanExecutePlan` to `ui/model.go`, then run:

`ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/tweaks ./ui -count=1`

Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/tweaks/danger.go pkg/tweaks/danger_test.go ui/model.go ui/model_test.go
git commit -m "feat: add shared danger zone policy"
```

### Task 2: Server-Enforced Web Capability

**Files:**
- Create: `pkg/api/danger.go`
- Create: `pkg/api/danger_test.go`
- Modify: `pkg/api/tweaks.go`
- Modify: `pkg/api/server.go`
- Modify: `pkg/api/security.go`
- Modify: `pkg/api/security_test.go`

**Step 1: Write failing endpoint tests**

Using a temporary registry containing Low-, Medium-, and High-risk fake tweaks, test that:

- `GET /api/tweaks` never serializes High-risk metadata.
- `GET /api/tweaks/dangerous` returns 403 without a capability.
- `POST /api/session/dangerous-unlock` rejects a missing acknowledgement or incorrect phrase.
- A valid unlock returns a non-empty capability.
- The returned capability allows dangerous listing.
- High-risk apply and revert return 403 without the capability and reach the handler with it.
- Low/Medium mutations remain governed by the existing CSRF/origin checks only.

**Step 2: Run the tests and verify RED**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/api -run 'TestDanger|TestTweaksEndpointHidesHighRisk' -count=1`

Expected: FAIL because the endpoint and capability enforcement do not exist.

**Step 3: Implement the capability session**

Create a server-lifetime `dangerSession` containing a cryptographically random capability. Decode unlock JSON strictly with a 64 KiB request limit inherited from mutation security. Validate the shared phrase policy and return the capability only on success. Compare capability headers in constant time.

**Step 4: Filter and protect tweak routes**

Change the ordinary list handler to serialize only `VisibleTweaks(..., false)`. Add a capability-protected dangerous list using `DangerousTweaks`. Wrap apply and revert so High-risk IDs require `X-ADS-Danger-Unlock`, while Low/Medium behavior is unchanged. Keep 404 for unknown IDs and 403 for a known locked High-risk ID.

**Step 5: Run tests and verify GREEN**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/api -count=1`

Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/api/danger.go pkg/api/danger_test.go pkg/api/tweaks.go pkg/api/server.go pkg/api/security.go pkg/api/security_test.go
git commit -m "feat: enforce danger zone capability in web api"
```

### Task 3: Guarded Web Panel

**Files:**
- Create: `pkg/api/web_contract_test.go`
- Modify: `pkg/api/web/index.html`
- Modify: `pkg/api/web/app.js`
- Modify: `pkg/api/web/style.css`

**Step 1: Write a failing embedded-asset contract test**

Read the embedded web assets and assert that they define a locked Danger Zone navigation entry and panel, acknowledgement checkbox, exact-phrase input, unlock button, dangerous capability header, and separate dangerous endpoint. Assert that neither `localStorage` nor `sessionStorage` is used for the capability.

**Step 2: Run the test and verify RED**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/api -run TestDangerZoneWebContract -count=1`

Expected: FAIL because the assets do not contain the guarded panel.

**Step 3: Build the locked and unlocked panel**

Add a Danger Zone sidebar item and tab containing only the warning/unlock form initially. Require the checkbox plus exact typed phrase before enabling the button. On success, keep the capability in a closure-scoped JavaScript variable, clear the phrase, replace the locked view with High-risk tweak cards, and attach the capability header to dangerous list/mutation requests. Do not persist it in cookies or browser storage.

**Step 4: Preserve execution confirmation**

Keep a separate `confirm()` immediately before every High-risk apply or revert. On a 403 response, discard the capability and return the panel to its locked state.

**Step 5: Style and verify GREEN**

Add visually distinct warning, acknowledgement, locked, and danger-card styles without external scripts. Run:

`ASDF_GOLANG_VERSION=1.24.6 go test ./pkg/api -count=1`

Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/api/web_contract_test.go pkg/api/web/index.html pkg/api/web/app.js pkg/api/web/style.css
git commit -m "feat: add guarded danger zone to web ui"
```

### Task 4: Guarded Native Panel

**Files:**
- Modify: `ui/layout.go`
- Modify: `ui/model_test.go`

**Step 1: Write a failing native visibility/state test**

Add a testable `DangerZoneState` expectation: it starts locked, rejects incomplete acknowledgement, unlocks for the current object after exact confirmation, and has no persistence API. Also verify that native search input uses the policy-filtered registry while locked.

**Step 2: Run the test and verify RED**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./ui -run 'TestDangerZoneState|TestNativeVisibleTweaks' -count=1`

Expected: FAIL because the state and visible-registry helper do not exist.

**Step 3: Implement the session-only native gate**

Create the state in memory inside the layout lifecycle. Build ordinary tabs and search from Low/Medium tweaks only. Add a locked Danger Zone tab with a warning button. Its custom dialog must require an acknowledgement checkbox and exact phrase before enabling confirmation. After success, replace the locked content with a High-risk-only category list and clear any active search.

**Step 4: Block hidden staged plans**

Before showing the normal plan confirmation, call `CanExecutePlan`. If locked desired state contains actionable High-risk work, show an error directing the user to unlock Danger Zone. Once unlocked, retain the existing High-risk confirmation before execution.

**Step 5: Run tests and verify GREEN**

Run: `ASDF_GOLANG_VERSION=1.24.6 go test ./ui -count=1`

Expected: PASS.

**Step 6: Commit**

```bash
git add ui/layout.go ui/model_test.go
git commit -m "feat: add guarded danger zone to native gui"
```

### Task 5: Documentation and Full Verification

**Files:**
- Modify: `README.md`
- Modify: `docs/SAFETY.md`
- Modify: `TODO.md`

**Step 1: Update documentation**

Document that High-risk tweaks are hidden in visual interfaces by default, unlocking is session-only, visibility unlock does not preapprove execution, the web capability is memory-only, and the CLI remains explicit and unchanged.

**Step 2: Format and run the complete suite**

Run:

```bash
gofmt -w pkg/tweaks/danger.go pkg/tweaks/danger_test.go pkg/api/danger.go pkg/api/danger_test.go pkg/api/tweaks.go pkg/api/server.go pkg/api/security.go pkg/api/security_test.go pkg/api/web_contract_test.go ui/model.go ui/model_test.go ui/layout.go
ASDF_GOLANG_VERSION=1.24.6 go test ./... -count=1
ASDF_GOLANG_VERSION=1.24.6 go test -race ./pkg/... -count=1
ASDF_GOLANG_VERSION=1.24.6 go vet ./...
ASDF_GOLANG_VERSION=1.24.6 go build ./...
git diff --check
```

Expected: all commands PASS and the worktree contains only intended changes.

**Step 3: Commit and push the PR branch**

```bash
git add README.md docs/SAFETY.md TODO.md docs/plans/2026-07-18-danger-zone.md
git commit -m "docs: explain guarded danger zone"
git push
```

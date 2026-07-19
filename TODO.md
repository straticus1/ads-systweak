# Roadmap

The initial safety overhaul is complete: direct argument execution, structured probe
states, typed first-write rollback receipts, atomic private state, exact restoration
for supported command tweaks, guarded localhost mutations, plan confirmation, risk
metadata, capability checks, and a modern extension inventory are implemented and
covered by tests.

## Release blockers

- [ ] Exercise every mutating tweak in disposable macOS VMs, recording expected
  apply/probe/revert behavior for each supported OS release.
- [ ] Run hardware-specific checks on Intel and Apple silicon Macs, especially NVRAM,
  power, networking, firewall, and security-related operations.
- [ ] Decide and document the supported macOS version range from that test matrix.
- [ ] Add CI for unit tests, race tests, vet, builds, and registry validation.
- [ ] Produce a signed, hardened-runtime, notarized universal application bundle.
- [ ] Add release automation, checksums, a changelog, and installation instructions.
- [ ] Capture current screenshots only after the release UI is frozen.

## Safety and correctness

- [ ] Add VM integration tests that verify actual state before and after rollback.
- [ ] Add an in-app receipt inspector/export flow without exposing unrelated values.
- [ ] Add a “restore all” preview showing exactly which receipts are actionable.
- [ ] Replace remaining legacy privileged command strings with narrowly typed helpers.
- [ ] Audit every registry entry against current Apple behavior and attach provenance
  and last-tested OS/build metadata.
- [ ] Define timeouts and cancellation for long-running external tools.
- [ ] Add structured, redacted audit logs for plans and command outcomes.

## macOS-native capabilities

- [ ] Use `SMAppService` for modern login-item, agent, and daemon inventory instead of
  treating launchd files as the complete source of truth.
- [ ] Add read-only Endpoint Security diagnostics when the required entitlement is
  available; keep the core app useful without private entitlements.
- [ ] Add NetworkExtension status and configuration diagnostics through public APIs.
- [ ] Add Unified Logging queries via `OSLogStore` for tweak-related failures.
- [ ] Add Disk Arbitration inventory and safe mount-policy diagnostics.
- [ ] Add APFS snapshot and Time Machine status views with explicit read-only modes.
- [ ] Add energy, thermal, battery-health, and power assertions diagnostics using
  supported frameworks and tools.
- [ ] Add privacy-permission visibility (TCC-facing guidance only; never bypass TCC).
- [ ] Add a sandboxed custom-tweak format with schema validation, signed community
  catalogs, explicit capabilities, and no arbitrary shell execution.

## Product and UX

- [ ] Search and filter by category, risk, support state, restart requirement, and
  last-tested macOS version.
- [ ] Show probe errors and remediation inline rather than only in dialogs or CLI text.
- [ ] Add config import/export with validation, diff preview, and secret-safe output.
- [ ] Add shell completion and machine-readable JSON output for automation.
- [ ] Add progress and cancellation for large plans.
- [ ] Improve accessibility, keyboard navigation, and VoiceOver labeling.
- [ ] Write per-tweak documentation explaining effect, risk, requirements, and exact
  rollback behavior.

## Distribution and community

- [ ] Add issue templates that request `doctor` output and a tweak ID.
- [ ] Add contributor guidance and a checklist for safe tweak submissions.
- [ ] Publish a Homebrew cask only after signing/notarization and release automation.
- [ ] Consider fleet/MDM workflows as a separate authenticated product architecture;
  do not extend the loopback web server into a remote management service.

## Explicit non-goals

- Bypassing Gatekeeper, SIP, TCC, or other macOS security controls.
- Claiming unsupported settings are reversible.
- Restoring guessed “factory defaults” when the original value was not recorded.
- Accepting arbitrary shell fragments from the GUI, API, config, or community catalog.
- Enabling analytics without explicit operator opt-in.

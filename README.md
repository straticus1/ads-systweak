# After Dark System Tools

An open-source macOS tweaker with 80+ settings for Finder, Dock, input, apps,
networking, power, privacy, and developer workflows. It includes a native GUI,
a scriptable CLI, staged plans, risk labels, and rollback receipts.

> This project changes system and application preferences. Review every plan,
> especially Medium- and High-risk changes, and keep a current backup of your Mac.

## What it does

- Probes each tweak as applied, off, unsupported, permission denied, or errored.
- Fails closed when current state cannot be determined.
- Stages desired changes and shows a Terraform-style plan before applying them.
- Hides High-risk tweaks behind a session-only Danger Zone warning, acknowledgement,
  and exact typed phrase in both visual interfaces.
- Requires a separate confirmation before every High-risk apply or revert; revealing
  the controls never preapproves execution.
- Saves typed preference values and supported command state before first mutation.
- Restores from receipts only after a successful rollback.
- Offers a local web UI protected by same-origin and CSRF checks.
- Keeps analytics disabled unless `ADS_ANALYTICS_ENABLED=true` is explicitly set.

Examples include showing hidden files, Finder path/status bars, Safari's developer
menu, faster Dock and Mission Control animations, screenshot preferences, keyboard
and trackpad settings, firewall stealth mode, Time Machine controls, and diagnostic
views for modern system and kernel extensions.

## Build and run

Go 1.24 or later is recommended.

```bash
git clone https://github.com/straticus1/ads-systweak.git
cd ads-systweak
go build -o ads-systweak .

# Native GUI
./ads-systweak

# Local web UI
./ads-systweak web
```

The application is not yet distributed as a signed and notarized release. Building
from source avoids implying that an unsigned development binary is production-ready.

## CLI workflow

```bash
# Check this Mac and the system tools the application can use
./ads-systweak doctor

# Inspect available tweaks and one tweak's status
./ads-systweak list
./ads-systweak status show-hidden-files

# Stage, review, and apply a preset
./ads-systweak preset stage developer-mode
./ads-systweak plan
./ads-systweak apply
```

`force-apply` bypasses desired-state staging, but it does not bypass compatibility,
state-probe, backup, or execution errors.

The CLI intentionally remains explicit rather than hiding IDs: selecting a High-risk
tweak by ID or applying a reviewed CLI plan is treated as an administrative workflow.
The native and web interfaces keep those tweaks out of ordinary lists and search until
the operator unlocks Danger Zone for that session.

## Presets

- **Developer Mode** enables developer-oriented Finder, Safari, Terminal, and text
  editing preferences.
- **Privacy & Hardening** groups privacy-related settings and firewall hardening;
  review the plan because availability varies by macOS release and permissions.
- **Maximum Performance** changes interface animation delays only. It does not
  disable security controls, search, backups, or filesystem integrity checks.

## State and rollback

Desired state and pre-approvals live in `~/.ads-systweak.json`. Preference and
command-state receipts live in `~/.ads-systweak-backups/`. These files are written
atomically with owner-only permissions.

Rollback is exact for supported scalar `defaults` values and command tweaks that
capture their original state. A receipt is retained if restoration fails. One-shot
actions and settings for which macOS exposes no reliable prior state are not claimed
to be reversible. See [docs/SAFETY.md](docs/SAFETY.md) for the detailed model.

## Compatibility

- Runtime mutations are supported on macOS only; other operating systems receive a
  warning and report mutation requirements as unsupported.
- The current compatibility pass was exercised on macOS 15.7.4 (Intel) and checks
  for required tools at runtime instead of assuming that every macOS release has
  identical commands.
- Some preferences are application- or release-specific. Missing commands, paths,
  permissions, and unreadable state are surfaced rather than interpreted as “off.”
- Administrator authentication is required for privileged operations.

Run `./ads-systweak doctor` when reporting a compatibility issue and include its
output, the failing tweak ID, and the error message. Do not include secrets or other
unrelated system information.

## Development

```bash
go test ./...
go test -race ./pkg/...
go vet ./...
go build ./...
```

New tweaks should use direct argument execution, define explicit compatibility
requirements, return structured probe states, declare an honest risk level, and add
tests for apply, probe, and rollback behavior. Avoid shell-string interpolation and
avoid promising reversibility without an exact restoration receipt.

## Status and roadmap

This remains a development project. Signing/notarization, packaged releases,
cross-version VM testing, and broader hardware testing are still outstanding. See
[TODO.md](TODO.md) for the current roadmap.

Licensed under the [MIT License](LICENSE).

# Safety model

After Dark System Tools changes macOS preferences and invokes Apple system tools.
Its safety model is designed to make uncertainty visible and to preserve the first
known state whenever an exact restoration is supported. It cannot make every macOS
setting reversible, nor can unit tests substitute for testing on each macOS release
and hardware family.

## Execution boundary

Normal commands are executed directly as a program plus an argument vector. User
values are not concatenated into a shell command. Failures retain the program name,
arguments, exit status, and standard error so the caller can report the real cause.

Privileged operations use macOS's administrator AppleScript prompt. The process is
still launched directly; the unavoidable privileged command string is escaped before
being embedded in AppleScript. New privileged tweaks should use fixed executable
paths and validate any variable input before constructing that command.

## Probe and plan behavior

A probe returns one of five states:

- `applied`: the requested behavior is currently active.
- `off`: the behavior is known not to be active.
- `unsupported`: a required operating system, executable, or path is unavailable.
- `permission_denied`: macOS would not allow the state to be read.
- `error`: the state could not be determined for another reason.

Only `applied` and `off` are actionable observations. Unsupported, denied, or errored
probes are not silently treated as “off”; planning and mutation fail closed. Presets
also avoid bundling High-risk tweaks.

Risk labels describe likely impact, not certainty. They are not a security boundary.
Both visual interfaces exclude High-risk names and controls from ordinary lists and
search. Their Danger Zone requires an acknowledgement and the exact phrase
`I KNOW WHAT I AM DOING`. The unlock exists only in memory and ends when the native
application closes or the browser page reloads. Revealing these controls does not
preapprove them: every High-risk apply or revert still requires a separate confirmation.

If configuration or the CLI previously staged actionable High-risk work, the locked
native GUI refuses to execute that plan and directs the operator to unlock Danger Zone.
It never applies a hidden change silently. The CLI remains visible and explicit because
selecting a tweak ID or applying a reviewed CLI plan is already an administrative act.

## Preference receipts

Before the first supported `defaults` mutation, the application reads and stores the
original domain/key state. Receipts support these scalar types without coercion:

- Boolean
- Integer
- Floating point
- String

Arrays, dictionaries, data blobs, dates, and unknown types are rejected by this
backup path. If the original key did not exist, that absence is recorded. If reading
or saving the original state fails, the mutation is not attempted.

Receipts are stored as hashed JSON filenames in `~/.ads-systweak-backups/`. The
directory is owner-only (`0700`), files are owner-only (`0600`), and writes use a
temporary file plus atomic rename. The first receipt is never overwritten by later
applies.

During rollback, an existing original value is written with its recorded type; an
originally absent value is deleted. A receipt is consumed only after restoration
succeeds. Failed restores leave the receipt available for diagnosis or retry.

## Command-state receipts

Some non-`defaults` tweaks capture a canonical state string before mutation. This is
used for supported NVRAM boot arguments, DNS configuration, `pmset`, `sysctl`, and
managed symlink changes. Their receipts share the private backup directory and the
same write-before-mutate and consume-after-restore rules.

A command tweak without a reliable prior-state capture must be represented as a
one-shot action or honestly document its non-exact revert semantics. “Restore macOS
defaults” is not considered an exact rollback unless the application recorded the
original state.

## Local web interface

The web server binds to `127.0.0.1`, not all network interfaces. Each process creates
a random CSRF token. Mutating API requests require that token; requests carrying an
`Origin` header must match the server origin. Mutation bodies are limited to 64 KiB.
Responses set a restrictive Content Security Policy plus anti-sniffing, referrer, and
framing headers. Defaults editing accepts only strict scalar JSON values and uses the
same backup-first execution path as other preference tweaks.

The ordinary tweaks endpoint never returns High-risk metadata. Completing the Danger
Zone acknowledgement returns a separate random capability, held only in page memory.
Dangerous listing, apply, and revert routes require that capability in a request header
and return `403 Forbidden` while locked. It is never placed in a cookie, URL, browser
storage, configuration file, or rollback receipt.

Loopback binding is not authentication. Do not expose this port through a proxy or
port-forward. A future remotely accessible service would require a separate identity,
authorization, and transport-security design.

## Local state and privacy

Desired state is stored in `~/.ads-systweak.json`; the optional analytics identifier
is stored in `~/.ads-systweak-uid`. Both use atomic owner-only files. The analytics ID
is random and is not derived from the hostname. Analytics remains off unless the
operator explicitly sets `ADS_ANALYTICS_ENABLED=true`.

Configuration and receipts can reveal which system preferences were selected. Treat
them as private user data and do not attach them wholesale to bug reports.

## Verification limits

The automated suite verifies execution isolation, typed receipts, failure retention,
probe-state behavior, request security, state-file permissions, registry metadata,
and GUI planning models. Before a signed release, the project still needs destructive
integration testing in disposable VMs across supported macOS releases, plus Intel and
Apple silicon hardware checks for firmware-, power-, network-, and security-related
tweaks.

Use `ads-systweak doctor` to report the OS/build and availability of system tools.
Test High-risk changes in a disposable environment first, keep a separate system
backup, and never assume that a successful command implies identical behavior on a
different macOS release.

# After Dark System Tools

> Unlock your Mac's hidden potential with 80+ professional-grade system tweaks

A free, open-source macOS optimization utility that gives you granular control over system settings Apple doesn't expose in System Preferences. Built for developers, power users, and anyone who wants their Mac to work *their* way.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![macOS](https://img.shields.io/badge/macOS-10.15+-blue.svg)

**By [After Dark](https://afterdarksys.com)** - Makers of professional system tools for developers and teams.

---

## Why After Dark System Tools?

- **80+ Battle-Tested Tweaks** - Disable .DS_Store files, show hidden files, speed up animations, enable Safari dev mode, trackpad improvements, keyboard optimization, and much more
- **Safety First** - Preview changes before applying, auto-detect current state, easy rollback
- **Dual Interface** - Beautiful GUI or powerful CLI for automation
- **Smart Presets** - One-click bundles: "Developer Mode", "Privacy & Hardening", "Maximum Performance"
- **Terraform-Style Planning** - Stage changes, review the plan, then apply
- **No Telemetry** - Open source, runs locally, your system stays yours

---

## Quick Start

### GUI Mode (Recommended)
```bash
./ads-systweak
```

### CLI Mode
```bash
# List all available tweaks
./ads-systweak list

# Stage a preset bundle
./ads-systweak preset stage developer-mode

# Review what will change
./ads-systweak plan

# Apply staged changes
./ads-systweak apply
```

---

## Featured Tweaks

### Developer Essentials
- Show hidden files and file extensions
- Enable Safari developer menu
- Show Finder path bar and status bar
- TextEdit defaults to plain text
- Disable window animations

### Privacy & Security
- Disable .DS_Store on network/USB drives
- Block captive portal detection
- Enable firewall stealth mode
- Disable crash reporter dialogs
- Disable iMessage read receipts

### Performance
- Disable Spotlight indexing
- Remove Dock auto-hide delay
- Speed up Mission Control animations
- Fast Dock animations
- Disable window animations

### Low-Level Control
- Mute startup chime
- Enable verbose boot mode
- Disable Gatekeeper (use cautiously!)
- Firewall stealth mode

[See all 80+ tweaks →](https://github.com/straticus1/ads-systweak/wiki/Tweaks)

---

## Installation

### Option 1: Download Binary
```bash
# Coming soon - Download from releases
```

### Option 2: Build from Source
```bash
# Clone the repository
git clone https://github.com/straticus1/ads-systweak.git
cd ads-systweak

# Build
go build -o ads-systweak .

# Run
./ads-systweak
```

### Option 3: Homebrew (Coming Soon)
```bash
brew install afterdark/tap/systweak
```

---

## How It Works

After Dark System Tools uses a **declarative state management** approach:

1. **Stage** - Check tweaks you want in the GUI or stage presets via CLI
2. **Plan** - Review exactly what will change on your system
3. **Apply** - Execute changes with one click (prompts for sudo when needed)
4. **Rollback** - Revert any tweak back to macOS defaults

All preferences are stored in `~/.ads-systweak.json` for easy backup and sharing.

---

## Screenshots

### Clean, Native macOS Interface
![Main Interface](docs/screenshot-main.png)

### Preset Bundles for Common Workflows
![Presets](docs/screenshot-presets.png)

### Powerful CLI for Automation
![CLI](docs/screenshot-cli.png)

---

## Safety & Transparency

- **No Hidden Actions** - Every command is logged and shown
- **Reversible** - All tweaks can be reverted to macOS defaults
- **Sudo Prompts** - Explicit permission for privileged operations
- **Open Source** - Audit the code yourself

---

## Presets

### Developer Mode
Essential tweaks for macOS developers: show hidden files, path bar, Safari dev menu, disable animations, plain text defaults.

### Privacy & Hardening
Disables captive portals, crash reporters, enables stealth firewall, blocks USB tracking.

### Maximum Performance
Disables Spotlight indexing and window animations for a snappier feel on older hardware.

---

## Advanced Usage

### CLI Examples

```bash
# Check status of a specific tweak
./ads-systweak status show-hidden-files

# Force-apply a single tweak (bypasses state)
./ads-systweak force-apply show-hidden-files

# Pre-approve a category for automation
./ads-systweak preapprove System

# Stage multiple tweaks
./ads-systweak preset stage privacy-hardening
./ads-systweak plan
./ads-systweak apply
```

### Configuration File
Your desired state is stored in `~/.ads-systweak.json`:

```json
{
  "desired_state": {
    "show-hidden-files": true,
    "dock-autohide": true,
    "safari-dev": true
  },
  "pre_approved": ["System", "Apps"]
}
```

Share this file with your team for consistent Mac setups!

---

## Requirements

- macOS 10.15 (Catalina) or later
- Administrator access for privileged tweaks (firewall, Gatekeeper, etc.)

---

## Building

```bash
# Install dependencies
go mod download

# Build for current platform
go build -o ads-systweak .

# Build for distribution
GOOS=darwin GOARCH=amd64 go build -o ads-systweak-amd64 .
GOOS=darwin GOARCH=arm64 go build -o ads-systweak-arm64 .
```

---

## Contributing

We welcome contributions! This tool is free and open source to help the Mac community.

- Add new tweaks to `pkg/tweaks/registry.go`
- Follow the existing `Tweak` interface pattern
- Test on multiple macOS versions if possible
- Submit a PR!

---

## Need More Power?

After Dark System Tools is our free gift to the macOS community. For teams and enterprises, check out our professional tools:

### [After Dark Pro Tools](https://afterdarksys.com)
- **Fleet Manager** - Deploy system configurations across hundreds of Macs
- **Audit & Compliance** - Track and enforce security policies
- **Custom Tweak Builder** - Create proprietary configuration packages
- **Priority Support** - Direct access to our engineering team

[Explore After Dark Products →](https://afterdarksys.com)

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

## Support

- **Issues & Bugs** - [GitHub Issues](https://github.com/straticus1/ads-systweak/issues)
- **Feature Requests** - [GitHub Discussions](https://github.com/straticus1/ads-systweak/discussions)
- **Commercial Support** - [afterdarksys.com/contact](https://afterdarksys.com/contact)

---

## Disclaimer

These tweaks modify system settings. While all changes are reversible, use at your own risk. Always test on a non-production machine first.

Some tweaks (like disabling Gatekeeper or Spotlight) reduce security. Understand what each tweak does before enabling it.

---

**Built with ❤️ by [After Dark](https://afterdarksys.com)**

[GitHub](https://github.com/straticus1/ads-systweak) • [Website](https://afterdarksys.com) • [Twitter](https://twitter.com/afterdarksys)

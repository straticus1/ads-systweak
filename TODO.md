# After Dark System Tools - Roadmap

## Pre-Launch (v1.0) - PRIORITY

### Critical for Public Release
- [ ] Add LICENSE file (MIT recommended)
- [ ] Create screenshot assets for README
  - [ ] Main UI with tweaks tab
  - [ ] Presets tab
  - [ ] CLI usage examples
- [ ] Code signing for macOS binary
- [ ] Create GitHub repository
  - [ ] Set proper topics: `macos`, `system-tweaks`, `optimization`, `golang`, `fyne`
  - [ ] Add repository description with keywords
  - [ ] Enable Discussions for community
- [ ] Add version flag (`--version`)
- [ ] macOS version compatibility check on startup

### Safety & UX Improvements
- [ ] Add risk levels to tweaks (Low/Medium/High)
  - [ ] Visual indicators in UI (green/yellow/red badges)
  - [ ] Warning dialogs for High-risk tweaks
- [ ] Backup mechanism
  - [ ] Save original values before first apply
  - [ ] "Restore All Defaults" button
  - [ ] Export backup to file
- [ ] Better error messages
  - [ ] Specific guidance when commands fail
  - [ ] Check for sudo access before privileged operations
  - [ ] Helpful hints for common errors
- [ ] Show which tweaks require restart in UI
- [ ] Confirmation dialog for dangerous operations (Gatekeeper, Spotlight)

### Distribution
- [ ] Build script / Makefile
  - [ ] Build for arm64 and amd64
  - [ ] Create universal binary
- [ ] GitHub Actions workflow
  - [ ] Auto-build on tag
  - [ ] Create releases with binaries
- [ ] Homebrew tap setup
  - [ ] Create formula
  - [ ] Setup afterdark/homebrew-tap repo
  - [ ] Test installation flow

---

## Marketing & Growth

### Launch Strategy
- [ ] Product Hunt launch
  - [ ] Prepare tagline: "Unlock your Mac's hidden potential"
  - [ ] Create demo video/GIF
  - [ ] Schedule launch date
- [ ] Reddit posts
  - [ ] r/macOS
  - [ ] r/MacOSBeta
  - [ ] r/golang (show HN style)
- [ ] Hacker News "Show HN"
- [ ] Twitter thread showcasing best tweaks
- [ ] Blog post: "25 Hidden macOS Tweaks Every Developer Should Know"
  - [ ] SEO optimize for "macOS tweaks", "optimize mac"
  - [ ] Link back to GitHub and afterdarksys.com

### Content Marketing
- [ ] YouTube walkthrough video
- [ ] Create comparison: Before/After benchmarks
  - [ ] Boot time improvements
  - [ ] Animation speed demos
- [ ] Case studies: "How Developer X Optimized Their Mac"
- [ ] Tweet each tweak individually with screenshots

### Community Building
- [ ] GitHub Discussions setup
  - [ ] "Share Your Setup" category
  - [ ] "Tweak Requests" category
- [ ] Discord/Slack community (optional)
- [ ] Contributor guide
- [ ] "Tweet your setup" feature in app

---

## Features (v1.1+)

### More Tweaks - Expand Registry
- [ ] **Screenshots**
  - [ ] Change default format (PNG/JPG)
  - [ ] Disable window shadow in screenshots
  - [ ] Set default save location
  - [ ] Include date in filename
- [ ] **Time Machine**
  - [ ] Disable local snapshots
  - [ ] Prevent throttling during backup
  - [ ] Exclude system files
- [ ] **Trackpad**
  - [ ] Enable three-finger drag
  - [ ] Tap to click
  - [ ] Tracking speed adjustments
- [ ] **Keyboard**
  - [ ] Key repeat rate
  - [ ] Disable auto-correct
  - [ ] Disable auto-capitalization
  - [ ] Function keys default behavior
- [ ] **Battery & Power**
  - [ ] Show battery percentage
  - [ ] Prevent sleep on lid close (external monitor)
  - [ ] Disable power chime
- [ ] **Activity Monitor**
  - [ ] Show all processes by default
  - [ ] Default sort by CPU
  - [ ] Open at login
- [ ] **Finder**
  - [ ] Default to list view
  - [ ] Show Library folder
  - [ ] Disable "Are you sure?" on empty trash
- [ ] **Dock**
  - [ ] Icon size adjustments
  - [ ] Minimize effects (Genie/Scale)
  - [ ] Position (left/bottom/right)

### UI Enhancements
- [ ] Icons for each category tab
- [ ] Risk level color coding throughout
- [ ] Export config button (share with friends)
- [ ] Import config from file
- [ ] Dark mode support (already works with Fyne, but test)
- [ ] Search improvements: filter by category, risk level
- [ ] Tweak details modal with more info
- [ ] "Popular" badge on most-used tweaks (from analytics)

### CLI Enhancements
- [ ] Color-coded output (errors in red, success in green)
- [ ] Progress indicators for batch operations
- [ ] `--dry-run` flag for apply command
- [ ] `--verbose` flag for detailed output
- [ ] Interactive mode: `systweak interactive`
- [ ] Shell completion (bash, zsh, fish)

### Analytics & Insights (Opt-in)
- [ ] Anonymous usage statistics
  - [ ] Which tweaks are most popular
  - [ ] macOS version distribution
  - [ ] Error frequency tracking
- [ ] Crash reporting (opt-in)
- [ ] "You might also like..." suggestions based on popular combos

### Social Features
- [ ] Generate shareable "My Mac Setup" card
  - [ ] Visual card with enabled tweaks
  - [ ] "Optimized with After Dark System Tools" watermark
- [ ] Export to Twitter/social media
- [ ] Preset sharing: publish to community registry
- [ ] Import community presets by URL

---

## Pro Features (Upsell to afterdarksys.com)

### Fleet Management (Enterprise)
- [ ] Deploy configs to multiple Macs
- [ ] Enforce compliance policies
- [ ] Audit trail of changes
- [ ] Centralized dashboard

### Advanced Tweaks (Pro)
- [ ] Scheduled tweak application
- [ ] Environment-based profiles (Work/Home)
- [ ] Network-based triggers
- [ ] Integration with MDM solutions

### Custom Tweak Builder
- [ ] Visual editor for creating custom tweaks
- [ ] Package as distributable
- [ ] Sign and notarize packages

---

## Technical Debt & Refactoring

### Code Quality
- [ ] Add unit tests
  - [ ] Test tweak interface implementations
  - [ ] Test state management
  - [ ] Mock shell commands for testing
- [ ] Integration tests
  - [ ] Test actual system changes in VM
- [ ] Documentation comments for exported functions
- [ ] Linting: golangci-lint setup
- [ ] CI pipeline: tests on push

### Performance
- [ ] Cache `IsApplied()` results to reduce shell calls
- [ ] Parallel status checks on startup
- [ ] Debounce search input in UI

### Architecture
- [ ] Consider plugin system for community tweaks
- [ ] Separate tweak definitions into YAML/JSON for easier contribution
- [ ] Abstract shell execution for testability

---

## Distribution & Platform

### Packaging
- [ ] Signed and notarized .app bundle
- [ ] DMG installer with custom background
- [ ] Sparkle framework for auto-updates
- [ ] Homebrew Cask support

### Multi-Platform (Future)
- [ ] Linux support (if feasible)
- [ ] Windows support (different tweaks, same UI)

---

## Community & Support

### Documentation
- [ ] Wiki with detailed tweak explanations
  - [ ] What each tweak does technically
  - [ ] Why you might want it
  - [ ] Potential risks
- [ ] FAQ section
- [ ] Video tutorials
- [ ] Contributing guide

### Support Channels
- [ ] GitHub Issues templates
  - [ ] Bug report
  - [ ] Feature request
  - [ ] Tweak request
- [ ] Commercial support offering
  - [ ] Link to afterdarksys.com/support
  - [ ] Priority bug fixes
  - [ ] Custom tweak development

---

## Metrics & Success

### Track These KPIs
- [ ] GitHub stars growth
- [ ] Download/install count
- [ ] Active users (if analytics enabled)
- [ ] Conversion to afterdarksys.com
  - [ ] Click-through rate on Pro features
  - [ ] Contact form submissions from app
- [ ] Community engagement
  - [ ] GitHub Discussions activity
  - [ ] Contributed tweaks
  - [ ] Social media mentions

### Growth Milestones
- [ ] 100 GitHub stars (HN worthy)
- [ ] 1,000 GitHub stars (Product Hunt trending)
- [ ] Featured in newsletters (Go Weekly, macOS blogs)
- [ ] 5,000+ downloads
- [ ] First community-contributed tweak merged

---

## Ideas for Future Versions

### Advanced Features
- Automated optimization recommendations based on system
- Integration with Homebrew to auto-apply after `brew install`
- Tweak profiles tied to physical location (work vs home)
- Time-based tweaks (disable animations during meetings)
- Compare with other users' setups anonymously

### Gamification
- Achievement system for trying tweaks
- "Power User" leaderboard
- Badges for different optimization levels

### Integration
- Alfred workflow
- Raycast extension
- Shell integration: `systweak toggle show-hidden-files`

---

## Notes

- Keep the free version genuinely useful (not crippled)
- Focus on developer/power user audience initially
- Use this as credibility builder for After Dark brand
- Every feature should serve either:
  1. User value (retention)
  2. Viral growth (sharing)
  3. Lead generation (links to products)

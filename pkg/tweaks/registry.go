package tweaks

func init() {
	// Disk & Storage
	Register(NewDefaultsTweak("ds-store-network", "Disable .DS_Store on Network Volumes", "Prevents the creation of .DS_Store files on SMB/NFS mounts.", CategoryNetworkStorage, RiskLow, "com.apple.desktopservices", "DSDontWriteNetworkStores", "bool", true, ""))
	Register(NewDefaultsTweak("ds-store-usb", "Disable .DS_Store on USB Volumes", "Prevents the creation of .DS_Store files on removable USB media.", CategoryDisk, RiskLow, "com.apple.desktopservices", "DSDontWriteUSBStores", "bool", true, ""))
	Register(NewDefaultsTweak("disk-verify", "Disable Disk Image Verification", "Skips verifying disk images (DMG) before mounting, speeding up the process.", CategoryDisk, RiskLow, "com.apple.frameworks.diskimages", "skip-verify", "bool", true, ""))

	// System & Finder
	Register(NewDefaultsTweak("show-hidden-files", "Show Hidden Files", "Reveals hidden items (dotfiles) in the Finder.", CategorySystem, RiskLow, "com.apple.finder", "AppleShowAllFiles", "bool", true, "Finder"))
	Register(NewDefaultsTweak("show-extensions", "Show All File Extensions", "Always display file extensions in the filesystem mapping.", CategorySystem, RiskLow, "NSGlobalDomain", "AppleShowAllExtensions", "bool", true, "Finder"))
	Register(NewDefaultsTweak("path-bar", "Show Finder Path Bar", "Displays the path bar at the bottom of Finder windows.", CategorySystem, RiskLow, "com.apple.finder", "ShowPathbar", "bool", true, "Finder"))
	Register(NewDefaultsTweak("status-bar", "Show Finder Status Bar", "Displays the status bar containing disk space in the Finder.", CategorySystem, RiskLow, "com.apple.finder", "ShowStatusBar", "bool", true, "Finder"))
	Register(NewDefaultsTweak("save-panel-expanded", "Expand Save Panel by Default", "Always show the expanded save dialog instead of the compact one.", CategorySystem, RiskLow, "NSGlobalDomain", "NSNavPanelExpandedStateForSaveMode", "bool", true, ""))
	Register(NewDefaultsTweak("print-panel-expanded", "Expand Print Panel by Default", "Always show the expanded print dialog.", CategorySystem, RiskLow, "NSGlobalDomain", "PMPrintingExpandedStateForPrint", "bool", true, ""))
	Register(NewDefaultsTweak("save-to-disk", "Save to Disk (Not iCloud) Default", "Defaults saving new documents to local disk instead of iCloud.", CategorySystem, RiskLow, "NSGlobalDomain", "NSDocumentSaveNewDocumentsToCloud", "bool", false, ""))
	Register(NewDefaultsTweak("no-ext-warning", "Disable Extension Change Warning", "Disables the warning dialog when changing a file extension.", CategorySystem, RiskLow, "com.apple.finder", "FXEnableExtensionChangeWarning", "bool", false, "Finder"))
	Register(NewDefaultsTweak("finder-list-view", "Finder: Default to List View", "Sets the default view style for all Finder folders to List View.", CategorySystem, RiskLow, "com.apple.finder", "FXPreferredViewStyle", "string", "Nlsv", "Finder"))
	Register(NewDefaultsTweak("show-library", "Show ~/Library Folder", "Makes the user Library folder visible in Finder.", CategorySystem, RiskLow, "com.apple.finder", "AppleShowAllFiles", "bool", true, "Finder"))
	Register(NewDefaultsTweak("trash-no-warning", "Disable Empty Trash Warning", "Skips the 'Are you sure?' dialog when emptying the trash.", CategorySystem, RiskLow, "com.apple.finder", "WarnOnEmptyTrash", "bool", false, "Finder"))
	Register(NewDefaultsTweak("finder-quit", "Enable Finder Quit Menu", "Adds 'Quit' to Finder's menu so you can close all windows.", CategorySystem, RiskLow, "com.apple.finder", "QuitMenuItem", "bool", true, "Finder"))
	Register(NewDefaultsTweak("finder-default-path", "Finder: Open Home on New Window", "Sets new Finder windows to open to your home directory instead of Recents.", CategorySystem, RiskLow, "com.apple.finder", "NewWindowTarget", "string", "PfHm", "Finder"))

	// Input & Trackpad
	Register(NewCommandTweak(
		"trackpad-tap-click",
		"Trackpad: Enable Tap to Click",
		"Enables single-finger tap to click for both internal and external Trackpads.",
		CategorySystem,
		RiskLow,
		`defaults read com.apple.driver.AppleBluetoothMultitouch.trackpad Clicking 2>/dev/null | grep -q '1' && echo "true" || echo "false"`,
		`defaults write com.apple.driver.AppleBluetoothMultitouch.trackpad Clicking -bool true && defaults write com.apple.AppleMultitouchTrackpad Clicking -bool true && defaults -currentHost write NSGlobalDomain com.apple.mouse.tapBehavior -int 1 && defaults write NSGlobalDomain com.apple.mouse.tapBehavior -int 1`,
		`defaults write com.apple.driver.AppleBluetoothMultitouch.trackpad Clicking -bool false && defaults write com.apple.AppleMultitouchTrackpad Clicking -bool false && defaults -currentHost write NSGlobalDomain com.apple.mouse.tapBehavior -int 0 && defaults write NSGlobalDomain com.apple.mouse.tapBehavior -int 0`,
		false,
	))

	// Keyboard & Input
	Register(NewDefaultsTweak("disable-press-and-hold", "Disable Press-and-Hold", "Disables press-and-hold for special characters, favoring key repeat.", CategoryOther, RiskLow, "NSGlobalDomain", "ApplePressAndHoldEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("fast-key-repeat", "Fast Key Repeat Rate", "Sets a very fast key repeat rate.", CategoryOther, RiskLow, "NSGlobalDomain", "KeyRepeat", "int", 2, ""))
	Register(NewDefaultsTweak("short-key-delay", "Short Key Repeat Delay", "Sets a short delay until key repeat starts.", CategoryOther, RiskLow, "NSGlobalDomain", "InitialKeyRepeat", "int", 15, ""))
	Register(NewDefaultsTweak("disable-autocorrect", "Disable Auto-Correct", "Turns off automatic spelling correction system-wide.", CategoryOther, RiskLow, "NSGlobalDomain", "NSAutomaticSpellingCorrectionEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("disable-autocapitalize", "Disable Auto-Capitalization", "Prevents automatic capitalization of first letters.", CategoryOther, RiskLow, "NSGlobalDomain", "NSAutomaticCapitalizationEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("disable-smart-quotes", "Disable Smart Quotes", "Uses straight quotes instead of curly quotes.", CategoryOther, RiskLow, "NSGlobalDomain", "NSAutomaticQuoteSubstitutionEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("disable-smart-dashes", "Disable Smart Dashes", "Uses regular hyphens instead of em/en dashes.", CategoryOther, RiskLow, "NSGlobalDomain", "NSAutomaticDashSubstitutionEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("disable-period-substitution", "Disable Period Substitution", "Disables double-space to period conversion.", CategoryOther, RiskLow, "NSGlobalDomain", "NSAutomaticPeriodSubstitutionEnabled", "bool", false, ""))

	// Three-finger drag (requires special handling)
	Register(NewCommandTweak(
		"three-finger-drag",
		"Trackpad: Enable Three-Finger Drag",
		"Enables dragging windows/items with three fingers on the trackpad.",
		CategorySystem,
		RiskLow,
		`defaults read com.apple.AppleMultitouchTrackpad TrackpadThreeFingerDrag 2>/dev/null | grep -q '1' && echo "true" || echo "false"`,
		`defaults write com.apple.AppleMultitouchTrackpad TrackpadThreeFingerDrag -bool true && defaults write com.apple.driver.AppleBluetoothMultitouch.trackpad TrackpadThreeFingerDrag -bool true && defaults -currentHost write NSGlobalDomain com.apple.trackpad.threeFingerSwipeGesture -int 1`,
		`defaults write com.apple.AppleMultitouchTrackpad TrackpadThreeFingerDrag -bool false && defaults write com.apple.driver.AppleBluetoothMultitouch.trackpad TrackpadThreeFingerDrag -bool false && defaults -currentHost delete NSGlobalDomain com.apple.trackpad.threeFingerSwipeGesture`,
		false,
	))

	// Dock & UI
	Register(NewDefaultsTweak("dock-autohide", "Auto-hide Dock", "Automatically hides and shows the Dock.", CategorySystem, RiskLow, "com.apple.dock", "autohide", "bool", true, "Dock"))
	Register(NewDefaultsTweak("dock-nodly", "Remove Auto-hide Delay", "Removes the hover delay before the Dock appears.", CategorySystem, RiskLow, "com.apple.dock", "autohide-delay", "float", float64(0), "Dock"))
	Register(NewDefaultsTweak("dock-fastanim", "Fast Dock Animation", "Speeds up the Dock sliding animation.", CategorySystem, RiskLow, "com.apple.dock", "autohide-time-modifier", "float", float64(0.2), "Dock"))
	Register(NewDefaultsTweak("expose-fast", "Fast Mission Control", "Speeds up Mission Control animations.", CategorySystem, RiskLow, "com.apple.dock", "expose-animation-duration", "float", float64(0.1), "Dock"))
	Register(NewDefaultsTweak("win-anim-off", "Disable Window Animations", "Disables standard macOS window opening and closing animations.", CategorySystem, RiskLow, "NSGlobalDomain", "NSAutomaticWindowAnimationsEnabled", "bool", false, ""))
	Register(NewDefaultsTweak("battery-percent", "Battery: Show Percentage", "Shows battery percentage in the menu bar.", CategorySystem, RiskLow, "com.apple.menuextra.battery", "ShowPercent", "string", "YES", "SystemUIServer"))
	Register(NewDefaultsTweak("dock-scale-effect", "Dock: Use Scale Effect", "Changes the minimize effect from Genie to Scale (faster).", CategorySystem, RiskLow, "com.apple.dock", "mineffect", "string", "scale", "Dock"))
	Register(NewDefaultsTweak("dock-show-recent", "Dock: Hide Recent Apps", "Hides the recent applications section in the Dock.", CategorySystem, RiskLow, "com.apple.dock", "show-recents", "bool", false, "Dock"))
	Register(NewDefaultsTweak("dock-minimize-to-app", "Dock: Minimize Into App Icon", "Minimizes windows into their application icon instead of the right side.", CategorySystem, RiskLow, "com.apple.dock", "minimize-to-application", "bool", true, "Dock"))
	Register(NewDefaultsTweak("dock-icon-size-small", "Dock: Small Icon Size", "Sets Dock icon size to 36 pixels (compact).", CategorySystem, RiskLow, "com.apple.dock", "tilesize", "int", 36, "Dock"))
	Register(NewDefaultsTweak("launchpad-reset-grid", "Launchpad: 7x5 Grid", "Sets Launchpad to show more apps (7 columns, 5 rows).", CategorySystem, RiskLow, "com.apple.dock", "springboard-columns", "int", 7, "Dock"))

	// Screenshots
	Register(NewCommandTweak(
		"screenshot-location",
		"Screenshots: Save to ~/Pictures/Screenshots",
		"Changes the default screenshot save location instead of cluttering the Desktop.",
		CategorySystem,
		RiskLow,
		`defaults read com.apple.screencapture location 2>/dev/null | grep -q 'Screenshots' && echo "true" || echo "false"`,
		`mkdir -p ~/Pictures/Screenshots && defaults write com.apple.screencapture location ~/Pictures/Screenshots && killall SystemUIServer`,
		`defaults delete com.apple.screencapture location && killall SystemUIServer`,
		false,
	))
	Register(NewDefaultsTweak("screenshot-shadow", "Screenshots: Disable Shadow", "Disables the drop shadow effect on individual window screenshots.", CategorySystem, RiskLow, "com.apple.screencapture", "disable-shadow", "bool", true, "SystemUIServer"))
	Register(NewDefaultsTweak("screenshot-format-jpg", "Screenshots: Save as JPG", "Changes screenshot format from PNG to JPG for smaller file sizes.", CategorySystem, RiskLow, "com.apple.screencapture", "type", "string", "jpg", "SystemUIServer"))
	Register(NewDefaultsTweak("screenshot-include-date", "Screenshots: Include Date in Filename", "Adds date/time to screenshot filenames.", CategorySystem, RiskLow, "com.apple.screencapture", "include-date", "bool", true, "SystemUIServer"))

	// Network
	Register(NewDefaultsTweak("disable-captive", "Disable Captive Portal", "Prevents macOS from automatically detecting and launching the captive portal login (useful for spoofing).", CategoryNetwork, RiskMedium, "/Library/Preferences/SystemConfiguration/com.apple.captive.control", "Active", "bool", false, ""))

	Register(NewCommandTweak(
		"wifi-power",
		"Keep Wi-Fi Enabled",
		"Toggles the main Wi-Fi interface power status (On/Off).",
		CategoryNetwork,
		RiskLow,
		`IFACE=$(networksetup -listallhardwareports | awk '/Wi-Fi/{getline; print $2}'); networksetup -getairportpower $IFACE | grep -q 'On' && echo "true" || echo "false"`,
		`IFACE=$(networksetup -listallhardwareports | awk '/Wi-Fi/{getline; print $2}'); networksetup -setairportpower $IFACE on`,
		`IFACE=$(networksetup -listallhardwareports | awk '/Wi-Fi/{getline; print $2}'); networksetup -setairportpower $IFACE off`,
		false,
	))

	Register(NewCommandTweak(
		"flush-dns",
		"Flush DNS Cache",
		"Clears the mDNSResponder and dscacheutil DNS cache (Revert does nothing) requires Admin.",
		CategoryNetwork,
		RiskMedium,
		`echo "false"`, // Always false so it can be 'Applied' to run it, but isn't persisting
		`dscacheutil -flushcache && killall -HUP mDNSResponder`,
		`echo "Cannot revert a cache flush"`,
		true,
	))

	Register(NewCommandTweak(
		"cloudflare-dns",
		"Use Cloudflare DNS (1.1.1.1)",
		"Sets your Wi-Fi DNS servers to Cloudflare's fast resolvers (Requires Admin).",
		CategoryNetwork,
		RiskLow,
		`networksetup -getdnsservers Wi-Fi 2>/dev/null | grep -q '1.1.1.1' && echo "true" || echo "false"`,
		`networksetup -setdnsservers Wi-Fi 1.1.1.1 1.0.0.1`,
		`networksetup -setdnsservers Wi-Fi empty`,
		true,
	))

	Register(NewCommandTweak(
		"google-dns",
		"Use Google DNS (8.8.8.8)",
		"Sets your Wi-Fi DNS servers to Google's public resolvers (Requires Admin).",
		CategoryNetwork,
		RiskLow,
		`networksetup -getdnsservers Wi-Fi 2>/dev/null | grep -q '8.8.8.8' && echo "true" || echo "false"`,
		`networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4`,
		`networksetup -setdnsservers Wi-Fi empty`,
		true,
	))

	Register(NewCommandTweak(
		"network-tcp-nodelay",
		"Disable TCP Delayed ACK (Lower Latency)",
		"Disables TCP delayed ACKs. Equivalent to disabling interrupt coalescing for faster network response (requires Admin).",
		CategoryNetwork,
		RiskLow,
		`/usr/sbin/sysctl net.inet.tcp.delayed_ack 2>/dev/null | grep -q '0' && echo "true" || echo "false"`,
		`/usr/sbin/sysctl -w net.inet.tcp.delayed_ack=0`,
		`/usr/sbin/sysctl -w net.inet.tcp.delayed_ack=3`,
		true,
	))

	// Low Level & Kernel
	Register(NewCommandTweak(
		"mute-chime",
		"Mute Startup Chime",
		"Disables the classic Mac startup sound (requires Admin).",
		CategoryLowLevel,
		RiskMedium,
		`nvram StartupMute 2>/dev/null | grep -q '%01' && echo "true" || echo "false"`,
		`nvram StartupMute=%01`,
		`nvram -d StartupMute`,
		true,
	))

	Register(NewCommandTweak(
		"verbose-boot",
		"Enable Verbose Boot",
		"Shows verbose text on startup instead of the Apple logo (requires Admin).",
		CategoryLowLevel,
		RiskMedium,
		`nvram boot-args 2>/dev/null | grep -q '\-v' && echo "true" || echo "false"`,
		`nvram boot-args="-v"`,
		`nvram boot-args=""`,
		true,
	))

	Register(NewCommandTweak(
		"server-perf-mode",
		"Enable Server Performance Mode",
		"Tunes the XNU CPU scheduler and memory manager for server workloads instead of desktop UI. (requires Admin).",
		CategoryLowLevel,
		RiskHigh,
		`nvram boot-args 2>/dev/null | grep -q 'serverperfmode=1' && echo "true" || echo "false"`,
		`nvram boot-args="serverperfmode=1"`,
		`nvram boot-args=""`,
		true,
	))

	Register(NewCommandTweak(
		"firewall-stealth",
		"Enable Firewall Stealth Mode",
		"Drops ICMP ping requests and doesn't answer closed TCP/UDP ports (requires Admin).",
		CategorySystem,
		RiskMedium,
		`/usr/libexec/ApplicationFirewall/socketfilterfw --getstealthmode | grep -q 'enabled' && echo "true" || echo "false"`,
		`/usr/libexec/ApplicationFirewall/socketfilterfw --setstealthmode on`,
		`/usr/libexec/ApplicationFirewall/socketfilterfw --setstealthmode off`,
		true,
	))

	Register(NewCommandTweak(
		"disable-gatekeeper",
		"Disable Gatekeeper",
		"Allows applications from anywhere to run seamlessly. Use cautiously! (requires Admin).",
		CategorySystem,
		RiskHigh,
		`spctl --status | grep -q 'disabled' && echo "true" || echo "false"`,
		`spctl --master-disable`,
		`spctl --master-enable`,
		true,
	))

	Register(NewCommandTweak(
		"disable-spotlight",
		"Disable Spotlight Indexing",
		"Globally turns off all mdutil Spotlight file indexing everywhere (requires Admin).",
		CategoryDisk,
		RiskHigh,
		`mdutil -s / | grep -q 'disabled' && echo "true" || echo "false"`,
		`mdutil -i off -a`,
		`mdutil -i on -a`,
		true,
	))

	// Time Machine
	Register(NewDefaultsTweak("tm-no-prompt", "Time Machine: Don't Prompt for New Disks", "Prevents Time Machine from asking to use newly connected drives for backups.", CategoryDisk, RiskLow, "com.apple.TimeMachine", "DoNotOfferNewDisksForBackup", "bool", true, ""))
	Register(NewCommandTweak(
		"tm-disable-local-snapshots",
		"Time Machine: Disable Local Snapshots",
		"Disables local Time Machine snapshots to save disk space (requires Admin).",
		CategoryDisk,
		RiskMedium,
		`tmutil listlocalsnapshots / 2>/dev/null | wc -l | grep -q '^0$' && echo "true" || echo "false"`,
		`tmutil disablelocal`,
		`tmutil enablelocal`,
		true,
	))
	Register(NewCommandTweak(
		"tm-no-throttle",
		"Time Machine: Disable Backup Throttling",
		"Prevents Time Machine from throttling backups when on battery (faster backups).",
		CategoryDisk,
		RiskLow,
		`sysctl debug.lowpri_throttle_enabled 2>/dev/null | grep -q '0' && echo "true" || echo "false"`,
		`sysctl debug.lowpri_throttle_enabled=0`,
		`sysctl debug.lowpri_throttle_enabled=1`,
		true,
	))

	// Apps
	Register(NewDefaultsTweak("safari-dev", "Safari: Enable Develop Menu", "Shows the web inspector and developer tools in Safari.", CategoryApps, RiskLow, "com.apple.Safari", "IncludeDevelopMenu", "bool", true, "Safari"))
	Register(NewDefaultsTweak("safari-no-autoopen", "Safari: Disable Auto-Open", "Stops Safari from automatically opening downloaded 'safe' files.", CategoryApps, RiskLow, "com.apple.Safari", "AutoOpenSafeDownloads", "bool", false, "Safari"))
	Register(NewDefaultsTweak("safari-show-full-url", "Safari: Show Full URL", "Displays the complete website address in the address bar.", CategoryApps, RiskLow, "com.apple.Safari", "ShowFullURLInSmartSearchField", "bool", true, "Safari"))
	Register(NewDefaultsTweak("safari-no-autoplay", "Safari: Prevent Auto-Play Videos", "Stops videos from auto-playing on websites.", CategoryApps, RiskLow, "com.apple.Safari", "WebKitMediaPlaybackAllowsInline", "bool", false, "Safari"))
	Register(NewDefaultsTweak("safari-backspace-back", "Safari: Backspace Goes Back", "Enables using Backspace key to navigate back in Safari.", CategoryApps, RiskLow, "com.apple.Safari", "com.apple.Safari.ContentPageGroupIdentifier.WebKit2BackspaceKeyNavigationEnabled", "bool", true, "Safari"))
	Register(NewDefaultsTweak("textedit-plain", "TextEdit: Default to Plain Text", "Sets TextEdit to use plain text for new documents instead of Rich Text.", CategoryApps, RiskLow, "com.apple.TextEdit", "RichText", "int", 0, "TextEdit"))
	Register(NewDefaultsTweak("chrome-no-swipe", "Chrome: Disable Swipe Navigation", "Prevents accidentally navigating back/forward using trackpad swipes in Chrome.", CategoryApps, RiskLow, "com.google.Chrome", "AppleEnableSwipeNavigateWithScrolls", "bool", false, "Google Chrome"))
	Register(NewDefaultsTweak("messages-readreq", "Messages: Disable Read Receipts", "Globally disables sending read receipts in iMessage.", CategoryApps, RiskLow, "com.apple.iChat", "smcReadReceiptsEnabled", "bool", false, "Messages"))
	Register(NewDefaultsTweak("mail-copy-plain", "Mail: Copy Addresses as Plain Text", "Copies email addresses without names when copying from Mail.", CategoryApps, RiskLow, "com.apple.mail", "AddressesIncludeNameOnPasteboard", "bool", false, "Mail"))
	Register(NewDefaultsTweak("mail-disable-inline", "Mail: Disable Inline Attachments", "Always shows attachments at the end instead of inline.", CategoryApps, RiskLow, "com.apple.mail", "DisableInlineAttachmentViewing", "bool", true, "Mail"))

	Register(NewDefaultsTweak("am-main-window", "Activity Monitor: Show Main Window", "Always opens the main window when launching Activity Monitor.", CategoryApps, RiskLow, "com.apple.ActivityMonitor", "OpenMainWindow", "bool", true, ""))
	Register(NewDefaultsTweak("am-sort-cpu", "Activity Monitor: Sort by CPU Usage", "Sorts Activity Monitor results by CPU usage by default.", CategoryApps, RiskLow, "com.apple.ActivityMonitor", "SortColumn", "string", "CPUUsage", ""))
	Register(NewDefaultsTweak("am-sort-desc", "Activity Monitor: Sort Descending", "Sorts Activity Monitor lists in descending order.", CategoryApps, RiskLow, "com.apple.ActivityMonitor", "SortDirection", "int", 0, ""))
	Register(NewDefaultsTweak("am-show-all-processes", "Activity Monitor: Show All Processes", "Shows all processes by default, not just current user.", CategoryApps, RiskLow, "com.apple.ActivityMonitor", "ShowCategory", "int", 100, ""))
	Register(NewDefaultsTweak("am-update-fast", "Activity Monitor: Fast Update Frequency", "Updates Activity Monitor stats every 1 second (default is 5).", CategoryApps, RiskLow, "com.apple.ActivityMonitor", "UpdatePeriod", "int", 1, ""))
	Register(NewDefaultsTweak("console-show-debug", "Console: Show Debug Messages", "Enables debug-level log messages in Console.app.", CategoryApps, RiskLow, "com.apple.Console", "ShowDebugMessages", "bool", true, "Console"))

	// Power Management
	Register(NewCommandTweak(
		"prevent-sleep-lid-closed",
		"Power: Prevent Sleep When Lid Closed (External Display)",
		"Keeps Mac awake when lid is closed and external monitor is connected (requires Admin).",
		CategorySystem,
		RiskMedium,
		`pmset -g | grep -q 'lidwake.*1' && echo "true" || echo "false"`,
		`pmset -a lidwake 1 && pmset -a sleep 0`,
		`pmset -a lidwake 0`,
		true,
	))

	Register(NewCommandTweak(
		"low-power-mode",
		"Power: Enable Low Power Mode (CPU Core Parking)",
		"Forces aggressive P-Core parking and disables Turbo Boost for battery savings at the cost of performance.",
		CategorySystem,
		RiskLow,
		`pmset -g | grep -q 'lowpowermode.*1' && echo "true" || echo "false"`,
		`pmset -a lowpowermode 1`,
		`pmset -a lowpowermode 0`,
		true,
	))

	Register(NewDefaultsTweak("disable-sudden-motion", "Disable Sudden Motion Sensor", "Disables the sudden motion sensor (SSD don't need it, saves CPU).", CategorySystem, RiskLow, "com.apple.PowerManagement", "Sudden-Motion-Sensor", "bool", false, ""))
	Register(NewCommandTweak(
		"disable-power-chime",
		"Disable Power Adapter Chime",
		"Prevents the chime sound when connecting the power adapter.",
		CategorySystem,
		RiskLow,
		`defaults read com.apple.PowerChime ChimeOnAllHardware 2>/dev/null | grep -q '0' && echo "true" || echo "false"`,
		`defaults write com.apple.PowerChime ChimeOnAllHardware -bool false && killall PowerChime`,
		`defaults write com.apple.PowerChime ChimeOnAllHardware -bool true && open /System/Library/CoreServices/PowerChime.app`,
		false,
	))

	// Developer Tools
	Register(NewCommandTweak(
		"xcode-faster-build",
		"Xcode: Use All CPU Cores for Building",
		"Enables parallel builds using all available CPU cores in Xcode.",
		CategoryApps,
		RiskLow,
		`defaults read com.apple.dt.Xcode IDEBuildOperationMaxNumberOfConcurrentCompileTasks 2>/dev/null | grep -qE '[0-9]+' && echo "true" || echo "false"`,
		`defaults write com.apple.dt.Xcode IDEBuildOperationMaxNumberOfConcurrentCompileTasks $(sysctl -n hw.ncpu)`,
		`defaults delete com.apple.dt.Xcode IDEBuildOperationMaxNumberOfConcurrentCompileTasks`,
		false,
	))
	Register(NewDefaultsTweak("xcode-show-build-time", "Xcode: Show Build Time", "Displays build duration in the Xcode activity viewer.", CategoryApps, RiskLow, "com.apple.dt.Xcode", "ShowBuildOperationDuration", "bool", true, "Xcode"))

	// Security & Privacy
	Register(NewDefaultsTweak("quarantine-no-warning", "Disable Quarantine Warning", "Skips the 'Downloaded from Internet' warning (use cautiously!).", CategorySystem, RiskHigh, "com.apple.LaunchServices", "LSQuarantine", "bool", false, ""))
	Register(NewCommandTweak(
		"clear-quarantine",
		"Clear All Quarantine Attributes",
		"Removes quarantine attributes from ~/Downloads (one-time action).",
		CategorySystem,
		RiskMedium,
		`echo "false"`, // Always show as not applied since it's a one-time action
		`xattr -rd com.apple.quarantine ~/Downloads`,
		`echo "Cannot revert a one-time cleanup"`,
		false,
	))

	// Bluetooth
	Register(NewDefaultsTweak("bluetooth-quality", "Bluetooth: Improve Audio Quality", "Increases Bluetooth bitrate for better audio quality (may reduce battery).", CategorySystem, RiskLow, "com.apple.BluetoothAudioAgent", "Apple Bitpool Min (editable)", "int", 40, ""))

	// Hidden CLI Tools
	Register(NewCommandTweak(
		"cli-airport",
		"Symlink 'airport' CLI",
		"Symlinks the hidden Wi-Fi 'airport' utility to /usr/local/bin/airport (requires Admin).",
		CategoryHiddenCLI,
		RiskLow,
		`[ -L /usr/local/bin/airport ] && [ "$(readlink /usr/local/bin/airport)" = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport" ] && echo "true" || echo "false"`,
		`mkdir -p /usr/local/bin && ln -sf /System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport /usr/local/bin/airport`,
		`rm -f /usr/local/bin/airport`,
		true,
	))

	Register(NewCommandTweak(
		"cli-lsregister",
		"Symlink 'lsregister' CLI",
		"Symlinks the Launch Services 'lsregister' utility to /usr/local/bin/lsregister (requires Admin).",
		CategoryHiddenCLI,
		RiskLow,
		`[ -L /usr/local/bin/lsregister ] && [ "$(readlink /usr/local/bin/lsregister)" = "/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister" ] && echo "true" || echo "false"`,
		`mkdir -p /usr/local/bin && ln -sf /System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister /usr/local/bin/lsregister`,
		`rm -f /usr/local/bin/lsregister`,
		true,
	))

	Register(NewCommandTweak(
		"cli-kickstart",
		"Symlink 'kickstart' CLI",
		"Symlinks the Apple Remote Desktop 'kickstart' utility to /usr/local/bin/kickstart (requires Admin).",
		CategoryHiddenCLI,
		RiskLow,
		`[ -L /usr/local/bin/kickstart ] && [ "$(readlink /usr/local/bin/kickstart)" = "/System/Library/CoreServices/RemoteManagement/ARDAgent.app/Contents/Resources/kickstart" ] && echo "true" || echo "false"`,
		`mkdir -p /usr/local/bin && ln -sf /System/Library/CoreServices/RemoteManagement/ARDAgent.app/Contents/Resources/kickstart /usr/local/bin/kickstart`,
		`rm -f /usr/local/bin/kickstart`,
		true,
	))

	Register(NewCommandTweak(
		"cli-plistbuddy",
		"Symlink 'PlistBuddy' CLI",
		"Symlinks the hidden plutil companion 'PlistBuddy' to /usr/local/bin/PlistBuddy (requires Admin).",
		CategoryHiddenCLI,
		RiskLow,
		`[ -L /usr/local/bin/PlistBuddy ] && [ "$(readlink /usr/local/bin/PlistBuddy)" = "/usr/libexec/PlistBuddy" ] && echo "true" || echo "false"`,
		`mkdir -p /usr/local/bin && ln -sf /usr/libexec/PlistBuddy /usr/local/bin/PlistBuddy`,
		`rm -f /usr/local/bin/PlistBuddy`,
		true,
	))

	// Other
	Register(NewDefaultsTweak("crash-reporter", "Disable Crash Reporter", "Disables the 'Application quit unexpectedly' dialog box.", CategoryOther, RiskLow, "com.apple.CrashReporter", "DialogType", "string", "none", ""))
	Register(NewDefaultsTweak("disable-help-viewer", "Disable Help Viewer on Top", "Prevents Help Viewer from always floating on top of other windows.", CategoryOther, RiskLow, "com.apple.helpviewer", "DevMode", "bool", true, ""))
	Register(NewDefaultsTweak("menubar-flash-clock", "Menu Bar: Flash Time Separators", "Makes the clock in the menu bar flash the time separators.", CategorySystem, RiskLow, "com.apple.menuextra.clock", "FlashDateSeparators", "bool", true, "SystemUIServer"))
	Register(NewDefaultsTweak("expand-save-dialog", "Always Expand Save/Open Dialogs", "Shows expanded file picker dialogs by default.", CategorySystem, RiskLow, "NSGlobalDomain", "NSNavPanelExpandedStateForSaveMode2", "bool", true, ""))
}

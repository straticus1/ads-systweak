package tweaks

type Preset struct {
	Name        string
	Description string
	TweakIDs    []string
}

var Presets = []Preset{
	{
		Name:        "Developer Mode",
		Description: "Essential tweaks for macOS developers: show hidden files, path bar, Safari dev menu, text editing defaults, and Xcode optimizations.",
		TweakIDs:    []string{"show-hidden-files", "show-extensions", "path-bar", "status-bar", "safari-dev", "win-anim-off", "textedit-plain", "show-library", "disable-autocorrect", "disable-smart-quotes", "disable-smart-dashes", "xcode-faster-build", "xcode-show-build-time", "am-show-all-processes", "console-show-debug"},
	},
	{
		Name:        "Privacy & Hardening",
		Description: "Disables captive portals, crash reporters, enables stealth firewall, blocks USB tracking, and hardens security settings.",
		TweakIDs:    []string{"ds-store-usb", "ds-store-network", "disable-captive", "crash-reporter", "firewall-stealth", "messages-readreq", "safari-no-autoopen", "safari-no-autoplay"},
	},
	{
		Name:        "Maximum Performance",
		Description: "Reduces interface animation delays without disabling security, search, backups, or integrity verification.",
		TweakIDs:    []string{"win-anim-off", "dock-nodly", "dock-fastanim", "expose-fast", "dock-scale-effect"},
	},
	{
		Name:        "Clean UI & Productivity",
		Description: "Streamlines the interface, improves Finder, removes clutter, and enhances workflow efficiency.",
		TweakIDs:    []string{"finder-list-view", "finder-default-path", "path-bar", "status-bar", "show-extensions", "trash-no-warning", "no-ext-warning", "save-panel-expanded", "print-panel-expanded", "dock-show-recent", "dock-minimize-to-app", "screenshot-location", "save-to-disk"},
	},
	{
		Name:        "Power User Keyboard & Input",
		Description: "Optimizes keyboard, trackpad, and input settings for power users and developers.",
		TweakIDs:    []string{"fast-key-repeat", "short-key-delay", "disable-press-and-hold", "trackpad-tap-click", "three-finger-drag", "disable-autocorrect", "disable-autocapitalize", "disable-smart-quotes", "disable-smart-dashes", "disable-period-substitution"},
	},
}

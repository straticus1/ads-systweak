package ui

import (
	"errors"
	"fmt"
	"image/color"
	"sort"
	"strings"

	"ads-systweak/pkg/analytics"
	"ads-systweak/pkg/backup"
	"ads-systweak/pkg/state"
	"ads-systweak/pkg/tweaks"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func BuildLayout(win fyne.Window) fyne.CanvasObject {
	cfg, err := state.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &state.Config{DesiredState: make(map[string]bool)}
	}

	tabs := container.NewAppTabs()
	tabs.SetTabLocation(container.TabLocationLeading)

	// Built-in Tweaks Tab (organized by category)
	tweaksTab := buildTweaksTab(win, cfg)
	tabs.Append(container.NewTabItem("Built-in Tweaks", tweaksTab))

	// Defaults Editor Tab (matching web interface)
	defaultsTab := buildDefaultsEditorTab(win)
	tabs.Append(container.NewTabItem("Defaults Editor", defaultsTab))

	// Presets Tab
	tabs.Append(container.NewTabItem("Presets", container.NewVScroll(buildPresetsList(win, cfg))))

	// Pro Features Tab
	tabs.Append(container.NewTabItem("Pro", container.NewVScroll(buildProFeaturesTab(win))))

	search := widget.NewEntry()
	search.SetPlaceHolder("Search tweaks...")

	contentArea := container.NewStack(tabs)

	search.OnChanged = func(s string) {
		if s == "" {
			contentArea.Objects = []fyne.CanvasObject{tabs}
			contentArea.Refresh()
			return
		}

		s = strings.ToLower(s)
		var results []tweaks.Tweak
		for _, t := range tweaks.Registry {
			if strings.Contains(strings.ToLower(t.Name()), s) || strings.Contains(strings.ToLower(t.Description()), s) {
				results = append(results, t)
			}
		}

		resView := container.NewVScroll(buildCategoryList(win, results, cfg))
		contentArea.Objects = []fyne.CanvasObject{resView}
		contentArea.Refresh()
	}

	applyBtn := widget.NewButtonWithIcon("Apply Staged Changes", nil, func() {
		var errs []string
		appliedCount := 0
		for _, item := range tweaks.BuildPlan(tweaks.Registry, cfg.DesiredState) {
			tw := item.Tweak
			if item.Action == tweaks.PlanBlocked {
				errs = append(errs, tw.Name()+": state is "+string(item.Probe.State)+": "+fmt.Sprint(item.Probe.Err))
				continue
			}
			var applyErr error
			if item.Action == tweaks.PlanApply {
				applyErr = tw.Apply()
			} else {
				applyErr = tw.Revert()
			}
			if applyErr != nil {
				errs = append(errs, tw.Name()+": "+applyErr.Error())
			} else {
				appliedCount++
			}
		}

		_ = state.SaveConfig(cfg) // Persist desired state

		if len(errs) > 0 {
			dialog.ShowError(errors.New(strings.Join(errs, "\n")), win)
		} else {
			dialog.ShowInformation("Success", fmt.Sprintf("Applied %d staged changes successfully.", appliedCount), win)
		}
	})
	applyBtn.Importance = widget.HighImportance

	bottomBar := container.NewHBox(layout.NewSpacer(), applyBtn)

	return container.NewBorder(container.NewPadded(search), container.NewPadded(bottomBar), nil, nil, contentArea)
}

func buildTweaksTab(win fyne.Window, cfg *state.Config) fyne.CanvasObject {
	// Create sub-tabs for each category
	categoryTabs := container.NewAppTabs()
	categoryTabs.SetTabLocation(container.TabLocationTop)

	categories := []tweaks.TweakCategory{
		tweaks.CategorySystem,
		tweaks.CategoryDisk,
		tweaks.CategoryNetwork,
		tweaks.CategoryNetworkStorage,
		tweaks.CategoryApps,
		tweaks.CategoryLowLevel,
		tweaks.CategoryMemory,
		tweaks.CategoryKernel,
		tweaks.CategoryOther,
	}

	// Add "All" tab showing all tweaks
	allTweaks := tweaks.Registry
	allView := container.NewVScroll(buildCategoryList(win, allTweaks, cfg))
	categoryTabs.Append(container.NewTabItem("All", allView))

	// Add category tabs
	for _, cat := range categories {
		catTweaks := tweaks.GetByCategory(cat)
		if len(catTweaks) == 0 {
			continue
		}
		vbox := container.NewVScroll(buildCategoryList(win, catTweaks, cfg))
		categoryTabs.Append(container.NewTabItem(string(cat), vbox))
	}

	return categoryTabs
}

func buildDefaultsEditorTab(win fyne.Window) fyne.CanvasObject {
	// Left sidebar: domain list
	domainList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Loading...")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)

	var domains []string
	var selectedDomain string
	var currentKeys map[string]interface{}

	// Right side: keys table
	keysTable := widget.NewTable(
		func() (int, int) {
			if currentKeys == nil {
				return 0, 3
			}
			return len(currentKeys), 3
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Key")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if currentKeys == nil {
				return
			}

			keys := make([]string, 0, len(currentKeys))
			for k := range currentKeys {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			if id.Row >= len(keys) {
				return
			}

			key := keys[id.Row]
			value := currentKeys[key]

			switch id.Col {
			case 0:
				label.SetText(key)
			case 1:
				label.SetText(fmt.Sprintf("%T", value))
			case 2:
				label.SetText(fmt.Sprintf("%v", value))
			}
		},
	)
	keysTable.SetColumnWidth(0, 200)
	keysTable.SetColumnWidth(1, 100)
	keysTable.SetColumnWidth(2, 300)

	// Load domains function
	loadDomains := func() {
		out, err := tweaks.RunShell("defaults domains")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load domains: %w", err), win)
			return
		}

		domains = strings.Split(out, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
		sort.Strings(domains)

		domainList.Length = func() int { return len(domains) }
		domainList.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("domain")
		}
		domainList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(domains[id])
		}
		domainList.Refresh()
	}

	// Load keys for selected domain
	loadKeys := func(domain string) {
		out, err := tweaks.RunShell(fmt.Sprintf("defaults export %s -", domain))
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read domain: %w", err), win)
			return
		}

		// Parse plist output (simplified - just show raw for now)
		currentKeys = make(map[string]interface{})
		// For now, just indicate we have data
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			if len(line) > 0 {
				currentKeys[fmt.Sprintf("Key_%d", i)] = line
			}
		}

		keysTable.Refresh()
	}

	domainList.OnSelected = func(id widget.ListItemID) {
		selectedDomain = domains[id]
		loadKeys(selectedDomain)
	}

	// Domain search
	domainSearch := widget.NewEntry()
	domainSearch.SetPlaceHolder("Filter domains...")

	// Header with restore button
	restoreBtn := widget.NewButton("Restore All Backups", func() {
		dialog.ShowConfirm("Restore All Backups",
			"This will restore all backed-up default values. Are you sure?",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				if err := backup.RestoreAll(); err != nil {
					dialog.ShowError(err, win)
				} else {
					dialog.ShowInformation("Success", "All backups restored successfully", win)
				}
			}, win)
	})
	restoreBtn.Importance = widget.DangerImportance

	header := container.NewBorder(nil, nil, nil, restoreBtn,
		widget.NewLabelWithStyle("macOS Defaults Editor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	// Left panel with search and domain list
	leftPanel := container.NewBorder(
		container.NewVBox(domainSearch),
		nil, nil, nil,
		container.NewScroll(domainList))

	// Right panel with keys table
	rightPanel := container.NewBorder(
		widget.NewLabel("Select a domain to view keys"),
		nil, nil, nil,
		container.NewScroll(keysTable))

	// Split layout
	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.3

	// Main container
	mainContainer := container.NewBorder(header, nil, nil, nil, split)

	// Load domains on creation
	loadDomains()

	return mainContainer
}

func buildPresetsList(win fyne.Window, cfg *state.Config) fyne.CanvasObject {
	items := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Optimization Presets", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	}

	for _, p := range tweaks.Presets {
		pr := p
		btn := widget.NewButton("Stage Preset", func() {
			for _, tID := range pr.TweakIDs {
				cfg.DesiredState[tID] = true
			}
			_ = state.SaveConfig(cfg)
			dialog.ShowInformation("Staged", "Preset '"+pr.Name+"' has been staged. Click 'Apply Staged Changes' to execute.", win)
			// Trigger UI refresh would be ideal, but for now simple dialog works.
		})

		title := widget.NewLabelWithStyle(pr.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		desc := widget.NewLabel(pr.Description)
		desc.Wrapping = fyne.TextWrapWord
		desc.TextStyle = fyne.TextStyle{Italic: true}

		row := container.NewVBox(
			container.NewBorder(nil, nil, nil, btn, title),
			container.NewPadded(desc),
		)
		items = append(items, row, widget.NewSeparator())
	}

	return container.NewPadded(container.NewVBox(items...))
}

func buildCategoryList(win fyne.Window, list []tweaks.Tweak, cfg *state.Config) fyne.CanvasObject {
	items := []fyne.CanvasObject{
		container.NewPadded(widget.NewLabelWithStyle("Tweaks (Check to stage state)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewSeparator(),
	}

	for _, t := range list {
		tw := t
		probe := tw.Probe()
		actual := probe.Applied
		known := probe.State == tweaks.ProbeApplied || probe.State == tweaks.ProbeOff

		desired, exists := cfg.DesiredState[tw.ID()]
		isChecked := actual
		if exists {
			isChecked = desired
		} else if known {
			// Populate initial desired state from actual to avoid unexpected diffs
			cfg.DesiredState[tw.ID()] = actual
		}

		check := widget.NewCheck(tw.Name(), func(b bool) {
			cfg.DesiredState[tw.ID()] = b
			_ = state.SaveConfig(cfg)
		})
		check.Checked = isChecked
		if !known {
			check.Disable()
		}

		// Risk level badge
		riskBadge := createRiskBadge(tw.RiskLevel())

		// Category badge
		categoryBadge := widget.NewLabel(string(tw.Category()))
		categoryBadge.TextStyle = fyne.TextStyle{Italic: true}

		// Status indicator
		statusText := "✓ Applied"
		statusColor := color.NRGBA{R: 0, G: 150, B: 0, A: 255} // Green
		if !known {
			statusText = "! " + string(probe.State)
			statusColor = color.NRGBA{R: 180, G: 40, B: 40, A: 255}
		} else if exists && desired != actual {
			if desired {
				statusText = "⋯ Pending (will apply)"
				statusColor = color.NRGBA{R: 255, G: 140, B: 0, A: 255} // Orange
			} else {
				statusText = "⋯ Pending (will revert)"
				statusColor = color.NRGBA{R: 255, G: 140, B: 0, A: 255} // Orange
			}
		} else if !actual {
			statusText = "○ Not Applied"
			statusColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Gray
		}

		statusLabel := canvas.NewText(statusText, statusColor)
		statusLabel.TextSize = 11

		// Header with checkbox, badges, and status
		headerRow := container.NewHBox(
			check,
			layout.NewSpacer(),
			riskBadge,
			statusLabel,
		)

		desc := widget.NewLabel(tw.Description())
		desc.Wrapping = fyne.TextWrapWord
		desc.TextStyle = fyne.TextStyle{Italic: true}

		row := container.NewVBox(
			headerRow,
			container.NewPadded(desc),
		)
		items = append(items, row, widget.NewSeparator())
	}

	return container.NewPadded(container.NewVBox(items...))
}

func createRiskBadge(risk tweaks.RiskLevel) fyne.CanvasObject {
	var bgColor color.Color
	var textColor color.Color = color.White

	switch risk {
	case tweaks.RiskLow:
		bgColor = color.NRGBA{R: 40, G: 167, B: 69, A: 255} // Green
	case tweaks.RiskMedium:
		bgColor = color.NRGBA{R: 255, G: 193, B: 7, A: 255} // Yellow
		textColor = color.Black
	case tweaks.RiskHigh:
		bgColor = color.NRGBA{R: 220, G: 53, B: 69, A: 255} // Red
	default:
		bgColor = color.Gray{Y: 128}
	}

	badge := canvas.NewRectangle(bgColor)
	badge.SetMinSize(fyne.NewSize(60, 20))

	text := canvas.NewText(string(risk), textColor)
	text.TextSize = 10
	text.Alignment = fyne.TextAlignCenter

	return container.NewStack(badge, container.NewCenter(text))
}

func buildProFeaturesTab(win fyne.Window) fyne.CanvasObject {
	items := []fyne.CanvasObject{
		widget.NewLabelWithStyle("After Dark Pro Tools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Enterprise-grade system management for teams and developers"),
		widget.NewSeparator(),
	}

	// Fleet Manager
	fleetTitle := widget.NewLabelWithStyle("Fleet Manager", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fleetDesc := widget.NewLabel("Deploy system configurations across hundreds of Macs. Perfect for dev teams and IT departments.")
	fleetDesc.Wrapping = fyne.TextWrapWord
	fleetFeatures := widget.NewLabel("• Centralized configuration management\n• Enforce compliance policies\n• Audit trail of all changes\n• Real-time status monitoring")
	fleetFeatures.Wrapping = fyne.TextWrapWord

	learnMoreFleet := widget.NewButton("Learn More", func() {
		analytics.TrackProLinkClicked("fleet")
		// Open URL in browser
		exec := fyne.CurrentApp().Driver().(interface {
			OpenURL(url string) error
		})
		_ = exec.OpenURL("https://afterdarksys.com/fleet")
	})
	learnMoreFleet.Importance = widget.LowImportance

	fleetSection := container.NewVBox(
		container.NewBorder(nil, nil, nil, learnMoreFleet, fleetTitle),
		fleetDesc,
		container.NewPadded(fleetFeatures),
	)

	// Custom Tweak Builder
	builderTitle := widget.NewLabelWithStyle("Custom Tweak Builder", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	builderDesc := widget.NewLabel("Create proprietary configuration packages for your organization.")
	builderDesc.Wrapping = fyne.TextWrapWord
	builderFeatures := widget.NewLabel("• Visual editor for custom tweaks\n• Package as distributable apps\n• Code signing and notarization\n• Version control integration")
	builderFeatures.Wrapping = fyne.TextWrapWord

	learnMoreBuilder := widget.NewButton("Learn More", func() {
		analytics.TrackProLinkClicked("builder")
		exec := fyne.CurrentApp().Driver().(interface {
			OpenURL(url string) error
		})
		_ = exec.OpenURL("https://afterdarksys.com/builder")
	})
	learnMoreBuilder.Importance = widget.LowImportance

	builderSection := container.NewVBox(
		container.NewBorder(nil, nil, nil, learnMoreBuilder, builderTitle),
		builderDesc,
		container.NewPadded(builderFeatures),
	)

	// Priority Support
	supportTitle := widget.NewLabelWithStyle("Priority Support", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	supportDesc := widget.NewLabel("Direct access to our engineering team for custom solutions.")
	supportDesc.Wrapping = fyne.TextWrapWord
	supportFeatures := widget.NewLabel("• Custom tweak development\n• Integration consulting\n• 24/7 emergency support\n• Dedicated Slack channel")
	supportFeatures.Wrapping = fyne.TextWrapWord

	contactSupport := widget.NewButton("Contact Sales", func() {
		analytics.TrackProLinkClicked("contact")
		exec := fyne.CurrentApp().Driver().(interface {
			OpenURL(url string) error
		})
		_ = exec.OpenURL("https://afterdarksys.com/contact")
	})
	contactSupport.Importance = widget.LowImportance

	supportSection := container.NewVBox(
		container.NewBorder(nil, nil, nil, contactSupport, supportTitle),
		supportDesc,
		container.NewPadded(supportFeatures),
	)

	// GitHub & Community
	communityTitle := widget.NewLabelWithStyle("Open Source & Community", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	communityDesc := widget.NewLabel("This tool is free forever. Star us on GitHub and join the community!")
	communityDesc.Wrapping = fyne.TextWrapWord

	githubBtn := widget.NewButton("Star on GitHub", func() {
		analytics.TrackProLinkClicked("github")
		exec := fyne.CurrentApp().Driver().(interface {
			OpenURL(url string) error
		})
		_ = exec.OpenURL("https://github.com/afterdarksys/systweak")
	})
	githubBtn.Importance = widget.HighImportance

	communitySection := container.NewVBox(
		communityTitle,
		communityDesc,
		container.NewPadded(githubBtn),
	)

	items = append(items,
		fleetSection,
		widget.NewSeparator(),
		builderSection,
		widget.NewSeparator(),
		supportSection,
		widget.NewSeparator(),
		communitySection,
	)

	return container.NewPadded(container.NewVBox(items...))
}

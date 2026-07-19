package ui

import (
	"errors"
	"fmt"
	"image/color"
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

	var applyBtn *widget.Button
	executePlan := func(plan []tweaks.PlanItem) {
		applyBtn.Disable()
		go func() {
			var errs []string
			appliedCount := 0
			for _, item := range plan {
				tw := item.Tweak
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

			fyne.Do(func() {
				applyBtn.Enable()
				if len(errs) > 0 {
					dialog.ShowError(errors.New(strings.Join(errs, "\n")), win)
				} else {
					dialog.ShowInformation("Success", fmt.Sprintf("Applied %d staged changes successfully.", appliedCount), win)
				}
			})
		}()
	}
	applyBtn = widget.NewButtonWithIcon("Apply Staged Changes", nil, func() {
		plan := tweaks.BuildPlan(tweaks.Registry, cfg.DesiredState)
		if len(plan) == 0 {
			dialog.ShowInformation("No Changes", "The system already matches the staged state.", win)
			return
		}
		summary := SummarizePlan(plan)
		if summary.Blocked > 0 {
			dialog.ShowError(fmt.Errorf("%d change(s) are blocked because their current state could not be determined", summary.Blocked), win)
			return
		}
		message := fmt.Sprintf("Apply %d and revert %d setting(s)?", summary.Apply, summary.Revert)
		dialog.ShowConfirm("Review Execution Plan", message, func(confirmed bool) {
			if !confirmed {
				return
			}
			if summary.HighRisk > 0 {
				dialog.ShowConfirm("High-Risk Changes", fmt.Sprintf("This plan contains %d high-risk change(s). Continue?", summary.HighRisk), func(highRiskConfirmed bool) {
					if highRiskConfirmed {
						executePlan(plan)
					}
				}, win)
				return
			}
			executePlan(plan)
		}, win)
	})
	applyBtn.Importance = widget.HighImportance

	bottomBar := container.NewHBox(layout.NewSpacer(), applyBtn)

	return container.NewBorder(container.NewPadded(search), container.NewPadded(bottomBar), nil, nil, contentArea)
}

func buildTweaksTab(win fyne.Window, cfg *state.Config) fyne.CanvasObject {
	// Create sub-tabs for each category
	categoryTabs := container.NewAppTabs()
	categoryTabs.SetTabLocation(container.TabLocationTop)

	categories := TweakCategories()

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
	var domains []string
	var filteredDomains []string
	var selectedDomain string
	var currentRows []DefaultRow
	selectedRow := -1

	domainList := widget.NewList(
		func() int { return len(filteredDomains) },
		func() fyne.CanvasObject { return widget.NewLabel("domain") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(filteredDomains) {
				obj.(*widget.Label).SetText(filteredDomains[id])
			}
		},
	)
	keysTable := widget.NewTable(
		func() (int, int) { return len(currentRows), 3 },
		func() fyne.CanvasObject { return widget.NewLabel("Key") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row < 0 || id.Row >= len(currentRows) {
				label.SetText("")
				return
			}
			row := currentRows[id.Row]
			switch id.Col {
			case 0:
				label.SetText(row.Key)
			case 1:
				label.SetText(row.Type)
			case 2:
				label.SetText(fmt.Sprintf("%v", row.Value))
			}
		},
	)
	keysTable.SetColumnWidth(0, 200)
	keysTable.SetColumnWidth(1, 100)
	keysTable.SetColumnWidth(2, 300)
	keysTable.OnSelected = func(id widget.TableCellID) { selectedRow = id.Row }

	var loadKeys func(string)
	loadDomains := func() {
		out, err := tweaks.RunCommand("/usr/bin/defaults", "domains")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load domains: %w", err), win)
			return
		}
		domains = strings.Split(out, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
		filteredDomains = FilterDomains(domains, "")
		domainList.Refresh()
	}
	loadKeys = func(domain string) {
		out, err := tweaks.RunCommand("/usr/bin/defaults", "export", domain, "-")
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read domain: %w", err), win)
			return
		}
		rows, err := DecodeDefaultsExport([]byte(out))
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to parse domain: %w", err), win)
			return
		}
		currentRows = rows
		selectedRow = -1
		keysTable.Refresh()
	}
	domainList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(filteredDomains) {
			selectedDomain = filteredDomains[id]
			loadKeys(selectedDomain)
		}
	}
	domainSearch := widget.NewEntry()
	domainSearch.SetPlaceHolder("Filter domains...")
	domainSearch.OnChanged = func(query string) {
		filteredDomains = FilterDomains(domains, query)
		domainList.Refresh()
	}

	editBtn := widget.NewButton("Edit Selected", func() {
		if selectedDomain == "" || selectedRow < 0 || selectedRow >= len(currentRows) {
			dialog.ShowInformation("Select a Key", "Select a defaults key first.", win)
			return
		}
		row := currentRows[selectedRow]
		domain := selectedDomain
		if row.Type == "array" || row.Type == "dict" {
			dialog.ShowError(fmt.Errorf("editing %s values is not supported safely", row.Type), win)
			return
		}
		entry := widget.NewEntry()
		entry.SetText(fmt.Sprintf("%v", row.Value))
		form := dialog.NewForm("Edit "+row.Key, "Save", "Cancel", []*widget.FormItem{widget.NewFormItem(row.Type, entry)}, func(ok bool) {
			if !ok {
				return
			}
			flag, value, err := ParseScalarInput(row.Type, entry.Text)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			go func() {
				if err := backup.SaveBackup(domain, row.Key); err == nil {
					_, err = tweaks.RunCommand("/usr/bin/defaults", "write", domain, row.Key, flag, value)
				}
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					loadKeys(domain)
				})
			}()
		}, win)
		form.Show()
	})

	deleteBtn := widget.NewButton("Delete Selected", func() {
		if selectedDomain == "" || selectedRow < 0 || selectedRow >= len(currentRows) {
			dialog.ShowInformation("Select a Key", "Select a defaults key first.", win)
			return
		}
		row := currentRows[selectedRow]
		domain := selectedDomain
		dialog.ShowConfirm("Delete Defaults Key", "Delete "+row.Key+" from "+domain+"?", func(ok bool) {
			if !ok {
				return
			}
			go func() {
				err := backup.SaveBackup(domain, row.Key)
				if err == nil {
					_, err = tweaks.RunCommand("/usr/bin/defaults", "delete", domain, row.Key)
				}
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					loadKeys(domain)
				})
			}()
		}, win)
	})
	deleteBtn.Importance = widget.DangerImportance

	restoreBtn := widget.NewButton("Restore All Backups", func() {
		dialog.ShowConfirm("Restore All Backups",
			"This will restore all backed-up default values. Are you sure?",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				go func() {
					err := backup.RestoreAll()
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, win)
						} else {
							dialog.ShowInformation("Success", "All supported backups restored successfully", win)
						}
					})
				}()
			}, win)
	})
	restoreBtn.Importance = widget.DangerImportance
	headerActions := container.NewHBox(editBtn, deleteBtn, restoreBtn)
	header := container.NewBorder(nil, nil, nil, headerActions,
		widget.NewLabelWithStyle("macOS Defaults Editor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	leftPanel := container.NewBorder(
		container.NewVBox(domainSearch),
		nil, nil, nil,
		container.NewScroll(domainList))
	rightPanel := container.NewBorder(
		widget.NewLabel("Select a domain, then select a key to edit or delete"),
		nil, nil, nil,
		container.NewScroll(keysTable))
	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.3
	mainContainer := container.NewBorder(header, nil, nil, nil, split)
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
			if err := state.SaveConfig(cfg); err != nil {
				dialog.ShowError(err, win)
				return
			}
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
		desired, exists := cfg.DesiredState[tw.ID()]

		check := widget.NewCheck(tw.Name(), func(b bool) {
			cfg.DesiredState[tw.ID()] = b
			if err := state.SaveConfig(cfg); err != nil {
				dialog.ShowError(err, win)
			}
		})
		if exists {
			check.Checked = desired
		}
		check.Disable()

		// Risk level badge
		riskBadge := createRiskBadge(tw.RiskLevel())

		statusLabel := canvas.NewText("… Checking", color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		statusLabel.TextSize = 11
		go func() {
			probe := tw.Probe()
			fyne.Do(func() {
				known := probe.State == tweaks.ProbeApplied || probe.State == tweaks.ProbeOff
				if !known {
					statusLabel.Text = "! " + string(probe.State)
					statusLabel.Color = color.NRGBA{R: 180, G: 40, B: 40, A: 255}
					statusLabel.Refresh()
					return
				}
				check.Enable()
				if !exists {
					check.Checked = probe.Applied
					check.Refresh()
				}
				switch {
				case exists && desired != probe.Applied && desired:
					statusLabel.Text = "⋯ Pending (will apply)"
					statusLabel.Color = color.NRGBA{R: 255, G: 140, B: 0, A: 255}
				case exists && desired != probe.Applied:
					statusLabel.Text = "⋯ Pending (will revert)"
					statusLabel.Color = color.NRGBA{R: 255, G: 140, B: 0, A: 255}
				case probe.Applied:
					statusLabel.Text = "✓ Applied"
					statusLabel.Color = color.NRGBA{R: 0, G: 150, B: 0, A: 255}
				default:
					statusLabel.Text = "○ Not Applied"
					statusLabel.Color = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
				}
				statusLabel.Refresh()
			})
		}()

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

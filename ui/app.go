package ui

import (
	"ads-systweak/pkg/analytics"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func StartApp() {
	// Initialize analytics (disabled by default, can be enabled via env var)
	analyticsEnabled := os.Getenv("ADS_ANALYTICS_ENABLED") == "true"
	analyticsEndpoint := os.Getenv("ADS_ANALYTICS_ENDPOINT")
	if analyticsEndpoint == "" {
		analyticsEndpoint = "https://analytics.afterdarksys.com/events" // Default endpoint
	}
	analytics.Initialize(analyticsEndpoint, analyticsEnabled)
	analytics.TrackAppLaunched("gui")

	a := app.New()

	// Apply custom macOS theme
	a.Settings().SetTheme(newMacTheme())

	w := a.NewWindow("After Dark System Tools")
	w.Resize(fyne.NewSize(800, 600))

	// Build UI populated from `pkg/tweaks.Registry`
	w.SetContent(BuildLayout(w))

	w.ShowAndRun()
}

package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type macTheme struct {
	fyne.Theme
}

func (m *macTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Custom macOS accent blue
	if n == theme.ColorNamePrimary {
		return color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}
	}
	if n == theme.ColorNameFocus {
		return color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x40} // Light blue outline
	}
	return m.Theme.Color(n, v)
}

func (m *macTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNamePadding {
		return 8
	}
	return m.Theme.Size(n)
}

func newMacTheme() fyne.Theme {
	return &macTheme{Theme: theme.DefaultTheme()}
}

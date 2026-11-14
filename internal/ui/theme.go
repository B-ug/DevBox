package ui

import (
	"devbox/internal/model"

	"github.com/egoist/mygo"
)

// ThemeSource is the MyGo appearance for the stored theme. System following
// is left for the settings page, so an unknown value stays dark.
func ThemeSource(a *model.App) mygo.ThemeSource {
	if a.Theme == model.ThemeLight {
		return mygo.ThemeLight
	}
	return mygo.ThemeDark
}

// ApplyTheme pushes the stored theme into MyGo. A test click skips this:
// SetSource from a goroutine other than the main thread waits for the app
// loop, and the loop is not running.
func ApplyTheme(a *model.App) {
	if !mygo.App.IsReady() {
		return
	}
	mygo.Theme.SetSource(ThemeSource(a))
}

// SyncTheme reads the segmented control's index after a change and applies it.
func SyncTheme(a *model.App) {
	if a.ThemeIndex == model.ThemeLightIndex {
		a.SetTheme(model.ThemeLight)
	} else {
		a.SetTheme(model.ThemeDark)
	}
	ApplyTheme(a)
}

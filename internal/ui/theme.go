package ui

import (
	"devbox/internal/model"

	"github.com/egoist/mygo"
)

// ThemeSource is the MyGo appearance for the stored theme. An unknown
// value stays dark.
func ThemeSource(a *model.App) mygo.ThemeSource {
	switch a.Theme {
	case model.ThemeLight:
		return mygo.ThemeLight
	case model.ThemeSystem:
		return mygo.ThemeSystem
	default:
		return mygo.ThemeDark
	}
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

package ui

import (
	"devbox/internal/model"

	"github.com/egoist/mygo"
)

// ThemeButtonLabel is the appearance the toolbar button will switch to.
func ThemeButtonLabel(a *model.App) string {
	if a.Theme == model.ThemeLight {
		return "深色"
	}
	return "浅色"
}

// ToggleTheme switches between dark and light. The running app follows
// immediately; tests keep only the state, because the theme service is
// not up until the application is ready.
func ToggleTheme(a *model.App) {
	if a.Theme == model.ThemeLight {
		a.Theme = model.ThemeDark
	} else {
		a.Theme = model.ThemeLight
	}
	if !mygo.App.IsReady() {
		return
	}
	if a.Theme == model.ThemeLight {
		mygo.Theme.SetSource(mygo.ThemeLight)
		return
	}
	mygo.Theme.SetSource(mygo.ThemeDark)
}

// Package app starts DevBox and owns the main window.
package app

import (
	"devbox/internal/model"
	appui "devbox/internal/ui"

	"github.com/egoist/mygo"
)

// Run opens the main window and runs the application.
func Run() error {
	a := model.New()
	mygo.Theme.SetSource(appui.ThemeSource(a))
	mygo.App.WhenReady(func() { openWindow(a) })
	return mygo.App.Run()
}

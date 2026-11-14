// Package app starts DevBox and owns the main window.
package app

import (
	"devbox/internal/model"
	appui "devbox/internal/ui"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
	"github.com/egoist/mygo/plugins/updater/native"
)

// Run opens the main window and runs the application.
func Run() error {
	// The update window is native UI and speaks the app's language.
	mygo.Use(native.New(updater.Options{Language: "zh-Hans"}))
	prepareIdentity()
	a := model.New()
	useConfig(a)
	mygo.Theme.SetSource(appui.ThemeSource(a))
	mygo.App.WhenReady(func() { openWindow(a) })
	return mygo.App.Run()
}

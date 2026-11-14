package app

import (
	"devbox/internal/model"
	appui "devbox/internal/ui"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const (
	windowWidth     = 1200
	windowHeight    = 800
	windowMinWidth  = 900
	windowMinHeight = 600
)

func openWindow(a *model.App) {
	editors := appui.NewEditors()
	mygo.NewWindow(windowOptions(ui.View(func(c *ui.Context) {
		appui.Layout(a, editors, c)
	})))
}

func windowOptions(content mygo.Content) mygo.WindowOptions {
	return mygo.WindowOptions{
		Title:     "DevBox",
		Width:     windowWidth,
		Height:    windowHeight,
		MinWidth:  windowMinWidth,
		MinHeight: windowMinHeight,
		StateKey:  "main",
		Content:   content,
	}
}

package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func toolbar(a *model.App, c *myui.Context) {
	myui.Toolbar(c.Key("toolbar"), func() {
		myui.Text(c, "DevBox").FontSize(16).Bold()
		myui.Spacer(c)
		if myui.Button(c, ThemeButtonLabel(a)).Clicked() {
			ToggleTheme(a)
		}
		if myui.Button(c, "设置").Clicked() {
			a.Status = model.StatusSettings
		}
	}).Label("工具栏")
}

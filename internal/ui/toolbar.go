package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func toolbar(a *model.App, c *myui.Context) {
	myui.Toolbar(c.Key("toolbar"), func() {
		myui.Text(c, "DevBox").FontSize(16).Bold()
		myui.Spacer(c)
		if myui.Segmented(c, &a.ThemeIndex, "深色", "浅色").Label("主题").Changed() {
			SyncTheme(a)
		}
		myui.Button(c, "设置").Tooltip("设置将在后续提供")
	}).Label("工具栏")
}

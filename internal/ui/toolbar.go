package ui

import myui "github.com/egoist/mygo/ui"

func toolbar(c *myui.Context) {
	myui.Toolbar(c.Key("toolbar"), func() {
		myui.Text(c, "DevBox").FontSize(16).Bold()
	}).Label("工具栏")
}

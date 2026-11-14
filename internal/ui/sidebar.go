package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

const sidebarWidth = 220

func sidebar(a *model.App, c *myui.Context) {
	t := c.Theme()
	myui.Column(c).Width(sidebarWidth).Shrink(0).Background(t.Surface).Children(func() {
		myui.Text(c, "工具导航").FontSize(12).Bold().TextColor(t.TextMuted).Padding(12, 16, 4)
		myui.Sidebar(c.Key("nav"), &a.CurrentTool, func() {
			myui.SidebarSection(c, "工具", nil, func() {
				for _, tool := range model.Tools() {
					myui.SidebarItem(c, tool.ID, nil, tool.Name)
				}
			})
			myui.SidebarSection(c, "最近使用", nil, func() {})
			myui.SidebarSection(c, "收藏工具", nil, func() {})
		}).Grow(1).Label("工具导航")
	})
}

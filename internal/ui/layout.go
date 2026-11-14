package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

// Layout builds the main window: toolbar, sidebar, workspace, and status bar.
func Layout(a *model.App, c *myui.Context) {
	t := c.Theme()
	myui.Column(c).Fill().Children(func() {
		toolbar(a, c)
		myui.Divider(c)
		myui.Row(c).Grow(1).MinHeight(0).AlignItems(myui.Stretch).Children(func() {
			sidebar(a, c)
			myui.Divider(c)
			workspace(a, c, t)
		})
		myui.Divider(c)
		statusbar(a, c)
	})
}

func workspace(a *model.App, c *myui.Context, t *myui.Theme) {
	title := "未选择工具"
	summary := ""
	if tool, ok := model.Lookup(a.CurrentTool); ok {
		title = tool.Name
		summary = tool.Summary
	}
	myui.Column(c).Grow(1).MinWidth(0).Padding(24).Gap(8).Background(t.Background).Children(func() {
		myui.Text(c, title).FontSize(22).Bold()
		if summary != "" {
			myui.Text(c, summary).TextColor(t.TextMuted)
		}
	})
}

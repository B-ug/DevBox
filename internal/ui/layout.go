package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

// Layout builds the main window: toolbar, sidebar, workspace, and status bar.
func Layout(a *model.App, editors *Editors, c *myui.Context) {
	t := c.Theme()
	myui.Column(c).Fill().Children(func() {
		toolbar(a, c)
		myui.Divider(c)
		myui.Row(c).Grow(1).MinHeight(0).AlignItems(myui.Stretch).Children(func() {
			sidebar(a, c)
			myui.Divider(c)
			workspace(a, editors, c, t)
		})
		myui.Divider(c)
		syncToolStatus(a)
		statusbar(a, c)
	})
}

// syncToolStatus clears the previous tool's result when the selection changes.
func syncToolStatus(a *model.App) {
	if a.ShownTool == a.CurrentTool {
		return
	}
	a.ShownTool = a.CurrentTool
	a.Status = model.StatusReady
	a.Error = ""
	a.Elapsed = ""
}

func workspace(a *model.App, editors *Editors, c *myui.Context, t *myui.Theme) {
	title := "未选择工具"
	summary := "从左侧选择一个工具"
	if tool, ok := model.Lookup(a.CurrentTool); ok {
		title = tool.Name
		summary = tool.Summary
	}
	myui.Column(c).Grow(1).MinWidth(0).Background(t.Background).Children(func() {
		myui.Column(c).Shrink(0).Padding(20, 24, 12).Gap(4).Children(func() {
			myui.Text(c, title).FontSize(22).Bold()
			myui.Text(c, summary).TextColor(t.TextMuted)
		})
		myui.Divider(c)
		switch a.CurrentTool {
		case model.ToolJSON:
			jsonWorkspace(a, editors.JSON, c)
		case model.ToolBase64:
			base64Workspace(a, editors.Base64, c)
		case model.ToolTimestamp:
			datetimeWorkspace(a, editors.Clock, c)
		default:
			myui.Column(c).Grow(1).MinWidth(0).MinHeight(0)
		}
	})
}

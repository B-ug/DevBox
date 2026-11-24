package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

// Layout builds the main window: toolbar, sidebar, workspace, and status bar.
func Layout(a *model.App, editors *Editors, c *myui.Context) {
	t := c.Theme()
	myui.Column(c).Fill().Children(func() {
		toolbar(c)
		myui.Divider(c)
		myui.Row(c).Grow(1).MinHeight(0).AlignItems(myui.Stretch).Children(func() {
			sidebar(a, c)
			myui.Divider(c)
			workspace(a, editors, c, t)
		})
		myui.Divider(c)
		statusbar(a, c)
	})
}

// syncNavigation applies the sidebar choice. Settings and About keep the
// last tool so returning restores it. A new page clears the previous
// tool's status.
func syncNavigation(a *model.App) {
	if a.Selection == "" {
		a.Selection = a.CurrentTool
	}
	if isAppPage(a.Selection) {
		if a.ShownTool != a.Selection {
			a.ShownTool = a.Selection
			clearStatus(a)
		}
		return
	}
	if _, ok := model.Lookup(model.ToolOf(a.Selection)); !ok {
		a.Selection = a.CurrentTool
	}
	id := model.ToolOf(a.Selection)
	a.Selection = id
	if a.CurrentTool == id && a.ShownTool == id {
		return
	}
	a.CurrentTool = id
	a.ShownTool = id
	clearStatus(a)
	if a.Remember(id) {
		persist(a)
	}
}

func clearStatus(a *model.App) {
	a.Status = model.StatusReady
	a.Error = ""
	a.Elapsed = ""
}

// isAppPage reports a sidebar choice that is not a tool. Those pages leave
// CurrentTool alone.
func isAppPage(selection string) bool {
	return selection == model.NavSettings || selection == model.NavAbout
}

func workspace(a *model.App, editors *Editors, c *myui.Context, t *myui.Theme) {
	syncNavigation(a)
	if editors != nil && editors.Gold != nil && a.Selection != model.ToolGold {
		editors.Gold.Leave()
	}
	switch a.Selection {
	case model.NavSettings:
		settingsWorkspace(a, c, t)
		return
	case model.NavAbout:
		aboutWorkspace(c, t)
		return
	}
	title := "未选择工具"
	summary := "从左侧选择一个工具"
	if tool, ok := model.Lookup(a.CurrentTool); ok {
		title = tool.Name
		summary = tool.Summary
	}
	myui.Column(c).Grow(1).MinWidth(0).Background(t.Background).Children(func() {
		myui.Row(c).Shrink(0).Padding(20, 24, 12).Gap(12).AlignItems(myui.Center).Children(func() {
			myui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
				myui.Text(c, title).FontSize(22).Bold().MinWidth(0)
				myui.Text(c, summary).TextColor(t.TextMuted).MinWidth(0)
			})
			if _, ok := model.Lookup(a.CurrentTool); ok {
				label := "收藏"
				if a.IsFavorite(a.CurrentTool) {
					label = "取消收藏"
				}
				if myui.Button(c, label).Shrink(0).Clicked() {
					if a.ToggleFavorite(a.CurrentTool) {
						persist(a)
					}
				}
			}
		})
		myui.Divider(c)
		switch a.CurrentTool {
		case model.ToolJSON:
			jsonWorkspace(a, editors.JSON, c)
		case model.ToolBase64:
			base64Workspace(a, editors.Base64, c)
		case model.ToolTimestamp:
			datetimeWorkspace(a, editors.Clock, c)
		case model.ToolPDF:
			pdfWorkspace(a, editors, c)
		case model.ToolTranslate:
			translateWorkspace(a, editors, c)
		case model.ToolTunnel:
			tunnelWorkspace(a, editors, c)
		case model.ToolGold:
			goldWorkspace(a, editors, c)
		default:
			myui.Column(c).Grow(1).MinWidth(0).MinHeight(0)
		}
	})
}

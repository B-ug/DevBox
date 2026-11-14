package ui

import (
	"github.com/egoist/mygo"
	myui "github.com/egoist/mygo/ui"
)

// projectURL is the DevBox repository. The about page links to it.
const projectURL = "https://github.com/B-ug/DevBox"

func aboutWorkspace(c *myui.Context, t *myui.Theme) {
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Background(t.Background).Children(func() {
		myui.Column(c).Shrink(0).Padding(20, 24, 12).Gap(4).Children(func() {
			myui.Text(c, "关于").FontSize(22).Bold()
			myui.Text(c, "DevBox").TextColor(t.TextMuted)
		})
		myui.Divider(c)
		myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Padding(20, 24).Gap(16).Children(func() {
			myui.Text(c, "Windows 上的开发者工具箱。").MinWidth(0)
			aboutFact(c, t, "版本", aboutVersion())
			aboutLink(c, t, "项目地址", projectURL)
		})
	})
}

func aboutFact(c *myui.Context, t *myui.Theme, label, value string) {
	myui.Column(c).Shrink(0).Gap(4).MinWidth(0).Children(func() {
		myui.Text(c, label).TextColor(t.TextMuted)
		myui.Text(c, value).MinWidth(0)
	})
}

func aboutLink(c *myui.Context, t *myui.Theme, label, url string) {
	myui.Column(c).Shrink(0).Gap(4).MinWidth(0).Children(func() {
		myui.Text(c, label).TextColor(t.TextMuted)
		myui.Link(c, url, url).MinWidth(0)
	})
}

// aboutVersion is the version MyGo recorded for this run. A test, or a
// binary that never set one, has none.
func aboutVersion() string {
	if v := mygo.App.Version(); v != "" {
		return v
	}
	return "—"
}

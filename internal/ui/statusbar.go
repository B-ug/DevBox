package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func statusbar(a *model.App, c *myui.Context) {
	t := c.Theme()
	elapsed := "—"
	if a.Elapsed != "" {
		elapsed = a.Elapsed
	}
	myui.Row(c).Shrink(0).Padding(8, 16).Gap(16).AlignItems(myui.Center).Background(t.Surface).Children(func() {
		myui.Text(c, "状态："+a.Status).FontSize(12).TextColor(t.TextMuted).Role(myui.RoleStatus).MinWidth(0)
		if a.Error != "" {
			myui.Text(c, "错误："+a.Error).FontSize(12).TextColor(t.Danger).Shrink(1).MinWidth(0)
		}
		myui.Spacer(c)
		myui.Text(c, "耗时："+elapsed).FontSize(12).TextColor(t.TextMuted).Shrink(0)
	})
}

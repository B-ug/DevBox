package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func statusbar(a *model.App, c *myui.Context) {
	t := c.Theme()
	myui.Row(c).Shrink(0).Padding(8, 16).AlignItems(myui.Center).Background(t.Surface).Children(func() {
		myui.Text(c, a.Status).FontSize(12).TextColor(t.TextMuted).Role(myui.RoleStatus).MinWidth(0)
	})
}

package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func settingsWorkspace(a *model.App, c *myui.Context, t *myui.Theme) {
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Background(t.Background).Children(func() {
		myui.Column(c).Shrink(0).Padding(20, 24, 12).Gap(4).Children(func() {
			myui.Text(c, "设置").FontSize(22).Bold()
			myui.Text(c, "外观").TextColor(t.TextMuted)
		})
		myui.Divider(c)
		myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Padding(20, 24).Gap(12).Children(func() {
			myui.Text(c, "主题").FontSize(16).Bold()
			if themeChoice(a, c) {
				a.SetTheme(a.Theme)
				ApplyTheme(a)
				persist(a)
			}
		})
	})
}

func themeChoice(a *model.App, c *myui.Context) bool {
	changed := false
	myui.RadioGroup(c.Key("theme"), func() {
		if myui.Radio(c, &a.Theme, model.ThemeDark, "深色").Changed() {
			changed = true
		}
		if myui.Radio(c, &a.Theme, model.ThemeLight, "浅色").Changed() {
			changed = true
		}
		if myui.Radio(c, &a.Theme, model.ThemeSystem, "跟随系统").Changed() {
			changed = true
		}
	}).Label("主题")
	return changed
}

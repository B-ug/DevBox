package ui

import (
	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func settingsWorkspace(a *model.App, c *myui.Context, t *myui.Theme) {
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Background(t.Background).Children(func() {
		myui.Column(c).Shrink(0).Padding(20, 24, 12).Gap(4).Children(func() {
			myui.Text(c, "设置").FontSize(22).Bold()
			myui.Text(c, "外观与翻译").TextColor(t.TextMuted)
		})
		myui.Divider(c)
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Padding(20, 24).Gap(12).MinWidth(0).Children(func() {
				myui.Text(c, "外观").FontSize(16).Bold()
				if themeChoice(a, c) {
					a.SetTheme(a.Theme)
					ApplyTheme(a)
					persist(a)
				}
				myui.Text(c, "翻译").FontSize(16).Bold()
				translateModelSettings(a, c, t)
			})
		})
	})
}

// translateModelSettings edits the model and prompt used by the translation tool.
func translateModelSettings(a *model.App, c *myui.Context, t *myui.Theme) {
	base := myui.TextInput(c.Key("translate-base"), &a.Translate.BaseURL).Placeholder(model.TranslateBaseURL).Label("Base URL").Grow(1).MinWidth(0)
	if base.Changed() {
		persist(a)
	}
	key := myui.TextInput(c.Key("translate-key"), &a.Translate.APIKey).Password().Placeholder("API Key").Label("API Key").Grow(1).MinWidth(0)
	if key.Changed() {
		persist(a)
	}
	name := myui.TextInput(c.Key("translate-model"), &a.Translate.Model).Placeholder(model.TranslateModel).Label("模型").Grow(1).MinWidth(0)
	if name.Changed() {
		persist(a)
	}
	prompt := myui.TextArea(c.Key("translate-prompt"), &a.Translate.Prompt).Placeholder("系统提示，可使用 {{target}}").Label("Prompt").MinHeight(120).MinWidth(0)
	if prompt.Changed() {
		persist(a)
	}
	myui.Text(c, "{{target}} 会替换成目标语言。原文作为用户消息发送。").FontSize(12).TextColor(t.TextMuted).MinWidth(0)
	if myui.Button(c, "恢复默认 Prompt").Clicked() {
		a.Translate.Prompt = model.TranslatePrompt
		persist(a)
		if a.Error != "设置未能保存" {
			a.Status = "已恢复默认 Prompt"
			a.Error = ""
			a.Elapsed = ""
		}
	}
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

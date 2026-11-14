package ui

import (
	"testing"

	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func TestLayout(t *testing.T) {
	a := model.New()
	tt := myui.NewTester(func(c *myui.Context) { Layout(a, c) }, 1200, 800)

	for _, text := range []string{"DevBox", "工具导航", "工具", "JSON", "Base64", "时间戳", "最近使用", "收藏工具", "格式化、压缩与校验", "就绪", "浅色", "设置"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}

	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	if a.CurrentTool != model.ToolBase64 || !tt.HasText("编码与解码") || tt.HasText("格式化、压缩与校验") {
		t.Fatalf("tool %s texts %q", a.CurrentTool, tt.Texts())
	}
	tt.Frame()
	if !tt.HasText("编码与解码") {
		t.Fatalf("texts after another frame %q", tt.Texts())
	}

	if err := tt.Click("时间戳"); err != nil {
		t.Fatal(err)
	}
	if a.CurrentTool != model.ToolTimestamp || !tt.HasText("秒与毫秒转换") {
		t.Fatalf("tool %s texts %q", a.CurrentTool, tt.Texts())
	}

	if err := tt.Click("浅色"); err != nil {
		t.Fatal(err)
	}
	if a.Theme != model.ThemeLight || !tt.HasText("深色") {
		t.Fatalf("theme %s texts %q", a.Theme, tt.Texts())
	}
	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	if a.Theme != model.ThemeDark || !tt.HasText("浅色") {
		t.Fatalf("theme %s texts %q", a.Theme, tt.Texts())
	}

	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	if a.Status != model.StatusSettings || !tt.HasText(model.StatusSettings) || tt.HasText(model.StatusReady) {
		t.Fatalf("status %s texts %q", a.Status, tt.Texts())
	}
}

func TestLayoutAtMinimumSize(t *testing.T) {
	a := model.New()
	tt := myui.NewTester(func(c *myui.Context) { Layout(a, c) }, 900, 600)
	for _, text := range []string{"DevBox", "JSON", "Base64", "时间戳", "就绪"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q at 900x600 in %q", text, tt.Texts())
		}
	}
}

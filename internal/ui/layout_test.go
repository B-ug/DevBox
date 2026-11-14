package ui

import (
	"testing"

	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func layoutTester(a *model.App, width, height int) (*Editors, *myui.Tester) {
	editors := NewEditors()
	tt := myui.NewTester(func(c *myui.Context) { Layout(a, editors, c) }, width, height)
	return editors, tt
}

func TestLayout(t *testing.T) {
	a := model.New()
	_, tt := layoutTester(a, 1200, 800)

	for _, text := range []string{
		"DevBox", "工具导航", "工具", "JSON", "Base64", "时间戳",
		"最近使用", "收藏工具", "格式化、压缩与校验",
		"状态：就绪", "耗时：—", "深色", "浅色", "设置",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if a.Theme != model.ThemeDark || a.ThemeIndex != model.ThemeDarkIndex {
		t.Fatalf("initial theme %s index %d", a.Theme, a.ThemeIndex)
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
	tt.Frame()
	if a.Theme != model.ThemeLight || a.ThemeIndex != model.ThemeLightIndex {
		t.Fatalf("theme %s index %d texts %q", a.Theme, a.ThemeIndex, tt.Texts())
	}
	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Theme != model.ThemeDark || a.ThemeIndex != model.ThemeDarkIndex {
		t.Fatalf("theme %s index %d texts %q", a.Theme, a.ThemeIndex, tt.Texts())
	}

	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	if a.Status != model.StatusReady || !tt.HasText("状态：就绪") {
		t.Fatalf("status %s texts %q", a.Status, tt.Texts())
	}
}

func TestStatusBarDetails(t *testing.T) {
	a := model.New()
	a.Error = "无效输入"
	a.Elapsed = "12ms"
	_, tt := layoutTester(a, 1200, 800)
	for _, text := range []string{"状态：就绪", "错误：无效输入", "耗时：12ms"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
}

func TestLayoutAtMinimumSize(t *testing.T) {
	a := model.New()
	_, tt := layoutTester(a, 900, 600)
	for _, text := range []string{"DevBox", "JSON", "Base64", "时间戳", "状态：就绪", "耗时：—"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q at 900x600 in %q", text, tt.Texts())
		}
	}
}

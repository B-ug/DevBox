package ui

import (
	"testing"

	"devbox/internal/config"
	"devbox/internal/model"

	"github.com/egoist/mygo"
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
		"状态：就绪", "耗时：—", "设置", "应用",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if tt.HasText("浅色") || tt.HasText("系统") {
		t.Fatalf("theme switch still in the toolbar: %q", tt.Texts())
	}
	if a.Theme != model.ThemeDark {
		t.Fatalf("initial theme %s", a.Theme)
	}

	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	if a.CurrentTool != model.ToolBase64 || !tt.HasText("标准与 URL Safe 编解码") || tt.HasText("格式化、压缩与校验") {
		t.Fatalf("tool %s texts %q", a.CurrentTool, tt.Texts())
	}
	tt.Frame()
	if !tt.HasText("标准与 URL Safe 编解码") {
		t.Fatalf("texts after another frame %q", tt.Texts())
	}

	if err := tt.Click("时间戳"); err != nil {
		t.Fatal(err)
	}
	if a.CurrentTool != model.ToolTimestamp || !tt.HasText("秒与毫秒转换") {
		t.Fatalf("tool %s texts %q", a.CurrentTool, tt.Texts())
	}

	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	if a.Selection != model.NavSettings || a.CurrentTool != model.ToolTimestamp {
		t.Fatalf("selection %s tool %s", a.Selection, a.CurrentTool)
	}

	if err := tt.Click("浅色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Theme != model.ThemeLight {
		t.Fatalf("theme %s texts %q", a.Theme, tt.Texts())
	}
	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Theme != model.ThemeDark {
		t.Fatalf("theme %s texts %q", a.Theme, tt.Texts())
	}
	if err := tt.Click("跟随系统"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Theme != model.ThemeSystem || a.Status != model.StatusReady || !tt.HasText("状态：就绪") {
		t.Fatalf("theme %s status %s texts %q", a.Theme, a.Status, tt.Texts())
	}
}

func TestSettingsPage(t *testing.T) {
	a := model.New()
	a.Error = "无效输入"
	a.Elapsed = "3ms"
	_, tt := layoutTester(a, 1200, 800)
	if !tt.HasText("错误：无效输入") {
		t.Fatalf("texts %q", tt.Texts())
	}

	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, text := range []string{"外观", "主题", "深色", "浅色", "跟随系统", "状态：就绪"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	for _, text := range []string{"当前外观", "选择左侧的工具即可返回", "深色和浅色会固定外观"} {
		if tt.HasText(text) {
			t.Fatalf("removed %q still in %q", text, tt.Texts())
		}
	}
	if a.Error != "" || a.Elapsed != "" || tt.HasText("错误：无效输入") || tt.HasText("格式化、压缩与校验") {
		t.Fatalf("error %q elapsed %q texts %q", a.Error, a.Elapsed, tt.Texts())
	}

	if err := tt.Click("跟随系统"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Theme != model.ThemeSystem {
		t.Fatalf("theme %s texts %q", a.Theme, tt.Texts())
	}

	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Selection != model.ToolJSON || a.CurrentTool != model.ToolJSON || !tt.HasText("格式化、压缩与校验") || tt.HasText("跟随系统") {
		t.Fatalf("selection %s tool %s texts %q", a.Selection, a.CurrentTool, tt.Texts())
	}
	if a.Theme != model.ThemeSystem {
		t.Fatalf("theme changed to %s", a.Theme)
	}
}

func TestSettingsSaveTheme(t *testing.T) {
	a := model.New()
	a.ConfigDir = t.TempDir()
	_, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("浅色"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	file, err := config.Load(a.ConfigDir)
	if err != nil || file.Theme != model.ThemeLight {
		t.Fatalf("file %+v err %v theme %s", file, err, a.Theme)
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
	for _, text := range []string{"DevBox", "JSON", "Base64", "时间戳", "状态：就绪", "耗时：—", "设置"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q at 900x600 in %q", text, tt.Texts())
		}
	}
}

func TestThemeSource(t *testing.T) {
	a := model.New()
	if ThemeSource(a) != mygo.ThemeDark {
		t.Fatalf("dark source %s", ThemeSource(a))
	}
	a.SetTheme(model.ThemeLight)
	if ThemeSource(a) != mygo.ThemeLight {
		t.Fatalf("light source %s", ThemeSource(a))
	}
	a.SetTheme(model.ThemeSystem)
	if ThemeSource(a) != mygo.ThemeSystem {
		t.Fatalf("system source %s", ThemeSource(a))
	}
}

package ui

import (
	"slices"
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
		"DevBox", "工具导航", "工具", "JSON", "Base64", "时间戳", "PDF", "翻译", "快速隧道",
		"最近使用", "收藏工具", "格式化、压缩与校验",
		"状态：就绪", "耗时：—", "设置", "关于", "应用",
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
	if !tt.HasText("JSON 内容") || !tt.HasText("内容：") || tt.HasText("JSON 输出") || tt.HasText("输入：") {
		t.Fatalf("editor texts %q", tt.Texts())
	}

	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	if a.CurrentTool != model.ToolBase64 || !tt.HasText("标准与 URL Safe 编解码") || !tt.HasText("Base64 内容") || tt.HasText("JSON 内容") || tt.HasText("格式化、压缩与校验") {
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

func TestAboutPage(t *testing.T) {
	mygo.App.SetVersion("0.1.0")
	a := model.New()
	a.Error = "无效输入"
	a.Elapsed = "3ms"
	_, tt := layoutTester(a, 1200, 800)
	settingsAt, aboutAt := -1, -1
	for i, text := range tt.Texts() {
		if text == "设置" && settingsAt < 0 {
			settingsAt = i
		}
		if text == "关于" && aboutAt < 0 {
			aboutAt = i
		}
	}
	if settingsAt < 0 || aboutAt < settingsAt {
		t.Fatalf("about button order settings %d about %d in %q", settingsAt, aboutAt, tt.Texts())
	}
	if tt.HasText("Windows 上的开发者工具箱。") || tt.HasText("检查更新") {
		t.Fatalf("about page open on start: %q", tt.Texts())
	}

	if err := tt.Click("关于"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Selection != model.NavAbout || a.CurrentTool != model.ToolJSON || len(a.Recent) != 0 {
		t.Fatalf("selection %s tool %s recent %v", a.Selection, a.CurrentTool, a.Recent)
	}
	for _, text := range []string{"关于", "DevBox", "Windows 上的开发者工具箱。", "版本", aboutVersion(), "检查更新", "项目地址", projectURL, "状态：就绪"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	checks := 0
	previousCheck := checkForUpdates
	checkForUpdates = func() { checks++ }
	t.Cleanup(func() { checkForUpdates = previousCheck })
	if err := tt.Click("检查更新"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if checks != 1 || a.Selection != model.NavAbout || !tt.HasText("检查更新") {
		t.Fatalf("checks %d selection %s texts %q", checks, a.Selection, tt.Texts())
	}
	if err := tt.Click(projectURL); err != nil {
		t.Fatal(err)
	}
	if got := tt.OpenedURLs(); !slices.Equal(got, []string{projectURL}) || a.Selection != model.NavAbout {
		t.Fatalf("opened %v selection %s", got, a.Selection)
	}
	if countExact(tt, aboutVersion()) < 1 || tt.HasText("外观") || tt.HasText("跟随系统") || tt.HasText("格式化、压缩与校验") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if a.Error != "" || a.Elapsed != "" || tt.HasText("错误：无效输入") {
		t.Fatalf("error %q elapsed %q texts %q", a.Error, a.Elapsed, tt.Texts())
	}

	tt.Frame()
	if a.Selection != model.NavAbout || !tt.HasText("Windows 上的开发者工具箱。") {
		t.Fatalf("left the about page: %s %q", a.Selection, tt.Texts())
	}

	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Selection != model.NavSettings || a.CurrentTool != model.ToolJSON || !tt.HasText("外观") || tt.HasText("Windows 上的开发者工具箱。") || tt.HasText(projectURL) || tt.HasText("检查更新") {
		t.Fatalf("selection %s tool %s texts %q", a.Selection, a.CurrentTool, tt.Texts())
	}

	if err := tt.Click("关于"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if checks != 1 || a.Selection != model.NavAbout || !tt.HasText("Windows 上的开发者工具箱。") || !tt.HasText("检查更新") || tt.HasText("外观") {
		t.Fatalf("checks %d selection %s texts %q", checks, a.Selection, tt.Texts())
	}

	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if checks != 1 || a.Selection != model.ToolJSON || a.CurrentTool != model.ToolJSON || !tt.HasText("格式化、压缩与校验") || tt.HasText("Windows 上的开发者工具箱。") || tt.HasText(projectURL) || tt.HasText("检查更新") || !tt.HasText("关于") {
		t.Fatalf("selection %s tool %s texts %q", a.Selection, a.CurrentTool, tt.Texts())
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

func TestRecentAndFavorites(t *testing.T) {
	a := model.New()
	a.ConfigDir = t.TempDir()
	_, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !slices.Equal(a.Recent, []string{model.ToolBase64}) || countExact(tt, "Base64") < 3 {
		t.Fatalf("recent %v texts %q", a.Recent, tt.Texts())
	}

	if err := tt.Click("收藏"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !slices.Equal(a.Favorites, []string{model.ToolBase64}) || countExact(tt, "取消收藏") != 1 {
		t.Fatalf("favorites %v texts %q", a.Favorites, tt.Texts())
	}
	file, err := config.Load(a.ConfigDir)
	if err != nil || file.Theme != model.ThemeDark || !slices.Equal(file.Recent, a.Recent) || !slices.Equal(file.Favorites, a.Favorites) {
		t.Fatalf("file %+v err %v", file, err)
	}

	if err := tt.Click("取消收藏"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.Favorites) != 0 || countExact(tt, "收藏") != 1 || countExact(tt, "取消收藏") != 0 {
		t.Fatalf("favorites %v texts %q", a.Favorites, tt.Texts())
	}

	a.Selection = model.RecentID(model.ToolJSON)
	tt.Frame()
	if a.Selection != model.ToolJSON || a.CurrentTool != model.ToolJSON || !slices.Equal(a.Recent, []string{model.ToolJSON, model.ToolBase64}) {
		t.Fatalf("selection %s tool %s recent %v", a.Selection, a.CurrentTool, a.Recent)
	}
}

func countExact(tt *myui.Tester, text string) int {
	n := 0
	for _, item := range tt.Texts() {
		if item == text {
			n++
		}
	}
	return n
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
	for _, text := range []string{"DevBox", "JSON", "Base64", "时间戳", "PDF", "翻译", "状态：就绪", "耗时：—", "设置", "关于"} {
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

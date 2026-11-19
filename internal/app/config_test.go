package app

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"devbox/internal/config"
	"devbox/internal/model"
)

func TestUseConfigDir(t *testing.T) {
	dir := t.TempDir()
	a := model.New()
	if err := useConfigDir(a, dir); err != nil {
		t.Fatal(err)
	}
	if a.ConfigDir != dir || a.Theme != model.ThemeDark {
		t.Fatalf("%+v", a)
	}

	if err := config.Save(dir, config.File{Theme: model.ThemeSystem}); err != nil {
		t.Fatal(err)
	}
	a = model.New()
	if err := useConfigDir(a, dir); err != nil {
		t.Fatal(err)
	}
	if a.Theme != model.ThemeSystem || a.ConfigDir != dir {
		t.Fatalf("%+v", a)
	}

	if err := config.Save(dir, config.File{Theme: "blue"}); err != nil {
		t.Fatal(err)
	}
	a = model.New()
	if err := useConfigDir(a, dir); err != nil || a.Theme != model.ThemeDark {
		t.Fatalf("err %v theme %s", err, a.Theme)
	}

	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	a = model.New()
	if err := useConfigDir(a, dir); err == nil || a.ConfigDir != dir || a.Theme != model.ThemeDark {
		t.Fatalf("err %v app %+v", err, a)
	}
}

func TestUseConfigTranslate(t *testing.T) {
	dir := t.TempDir()
	err := config.Save(dir, config.File{
		Translate: config.Translate{APIKey: "sk-test", Prompt: "custom {{target}}", Target: "nope"},
	})
	if err != nil {
		t.Fatal(err)
	}
	a := model.New()
	if err := useConfigDir(a, dir); err != nil {
		t.Fatal(err)
	}
	if a.Translate.APIKey != "sk-test" || a.Translate.Prompt != "custom {{target}}" || a.Translate.Target != model.TranslateTarget {
		t.Fatalf("%+v", a.Translate)
	}
	if a.Translate.BaseURL != model.TranslateBaseURL || a.Translate.Model != model.TranslateModel {
		t.Fatalf("defaults %+v", a.Translate)
	}

	if err := config.Save(dir, config.File{Theme: model.ThemeLight}); err != nil {
		t.Fatal(err)
	}
	a = model.New()
	if err := useConfigDir(a, dir); err != nil {
		t.Fatal(err)
	}
	if a.Translate.APIKey != "" || a.Translate.Prompt != model.TranslatePrompt || a.Translate.Model != model.TranslateModel {
		t.Fatalf("old file %+v", a.Translate)
	}
}

func TestUseConfigLists(t *testing.T) {
	dir := t.TempDir()
	err := config.Save(dir, config.File{
		Theme:     model.ThemeLight,
		Recent:    []string{model.ToolBase64, "nope", model.ToolBase64, model.ToolJSON},
		Favorites: []string{model.ToolTimestamp, model.ToolTimestamp, "nope"},
	})
	if err != nil {
		t.Fatal(err)
	}
	a := model.New()
	if err := useConfigDir(a, dir); err != nil {
		t.Fatal(err)
	}
	if a.Theme != model.ThemeLight {
		t.Fatalf("theme %s", a.Theme)
	}
	if !slices.Equal(a.Recent, []string{model.ToolBase64, model.ToolJSON}) {
		t.Fatalf("recent %v", a.Recent)
	}
	if !slices.Equal(a.Favorites, []string{model.ToolTimestamp}) {
		t.Fatalf("favorites %v", a.Favorites)
	}
}

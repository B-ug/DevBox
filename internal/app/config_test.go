package app

import (
	"os"
	"path/filepath"
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

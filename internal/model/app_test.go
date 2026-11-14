package model

import "testing"

func TestSetTheme(t *testing.T) {
	a := New()
	if a.Theme != ThemeDark || a.Selection != ToolJSON {
		t.Fatalf("initial %+v", a)
	}

	a.SetTheme(ThemeLight)
	if a.Theme != ThemeLight {
		t.Fatalf("light %+v", a)
	}
	a.SetTheme(ThemeSystem)
	if a.Theme != ThemeSystem {
		t.Fatalf("system %+v", a)
	}
	a.SetTheme("other")
	if a.Theme != ThemeDark {
		t.Fatalf("unknown %+v", a)
	}
}

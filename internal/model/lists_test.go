package model

import (
	"slices"
	"testing"
)

func TestRemember(t *testing.T) {
	a := New()
	if a.Remember("missing") {
		t.Fatal("unknown tool was recorded")
	}
	if !a.Remember(ToolBase64) || !slices.Equal(a.Recent, []string{ToolBase64}) {
		t.Fatalf("recent %v", a.Recent)
	}
	if a.Remember(ToolBase64) {
		t.Fatal("recorded the tool that was already first")
	}
	if !a.Remember(ToolJSON) || !a.Remember(ToolTimestamp) || !a.Remember(ToolBase64) {
		t.Fatal("expected later tools to move to the front")
	}
	if !slices.Equal(a.Recent, []string{ToolBase64, ToolTimestamp, ToolJSON}) {
		t.Fatalf("recent %v", a.Recent)
	}
}

func TestKeepToolsLimit(t *testing.T) {
	got := keepTools([]string{ToolJSON, "nope", ToolJSON, ToolBase64, ToolTimestamp}, 2)
	if !slices.Equal(got, []string{ToolJSON, ToolBase64}) {
		t.Fatalf("kept %v", got)
	}
}

func TestToggleFavorite(t *testing.T) {
	a := New()
	if a.ToggleFavorite("missing") || a.IsFavorite(ToolJSON) {
		t.Fatal("unknown tool changed favorites")
	}
	if !a.ToggleFavorite(ToolJSON) || !a.ToggleFavorite(ToolBase64) {
		t.Fatal("pin failed")
	}
	if !slices.Equal(a.Favorites, []string{ToolJSON, ToolBase64}) {
		t.Fatalf("favorites %v", a.Favorites)
	}
	if !a.ToggleFavorite(ToolJSON) || !slices.Equal(a.Favorites, []string{ToolBase64}) || a.IsFavorite(ToolJSON) {
		t.Fatalf("favorites %v", a.Favorites)
	}
}

func TestToolOf(t *testing.T) {
	if ToolOf(RecentID(ToolJSON)) != ToolJSON || ToolOf(FavoriteID(ToolBase64)) != ToolBase64 || ToolOf(ToolTimestamp) != ToolTimestamp {
		t.Fatalf("json %s base64 %s time %s", ToolOf(RecentID(ToolJSON)), ToolOf(FavoriteID(ToolBase64)), ToolOf(ToolTimestamp))
	}
}

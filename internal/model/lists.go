package model

import (
	"slices"
	"strings"
)

const (
	// RecentPrefix marks a sidebar shortcut into Recent. The selection is
	// resolved back to the tool id so the catalog entry stays the choice.
	RecentPrefix = "recent:"
	// FavoritePrefix marks a sidebar shortcut into Favorites.
	FavoritePrefix = "favorite:"
	// MaxRecent is how many tools the recent list keeps.
	MaxRecent = 8
)

// RecentID is the sidebar id of a recent shortcut for tool id.
func RecentID(id string) string { return RecentPrefix + id }

// FavoriteID is the sidebar id of a favorite shortcut for tool id.
func FavoriteID(id string) string { return FavoritePrefix + id }

// ToolOf returns the tool id for a catalog entry or a recent or favorite shortcut.
func ToolOf(selection string) string {
	for _, prefix := range []string{RecentPrefix, FavoritePrefix} {
		if id, ok := strings.CutPrefix(selection, prefix); ok {
			return id
		}
	}
	return selection
}

// Remember moves id to the front of Recent. An unknown id is ignored.
// It reports whether the list changed.
func (a *App) Remember(id string) bool {
	if _, ok := Lookup(id); !ok {
		return false
	}
	next := keepTools(append([]string{id}, a.Recent...), MaxRecent)
	if slices.Equal(next, a.Recent) {
		return false
	}
	a.Recent = next
	return true
}

// IsFavorite reports whether id is pinned.
func (a *App) IsFavorite(id string) bool {
	return slices.Contains(a.Favorites, id)
}

// ToggleFavorite pins id, or removes it when it is already pinned.
// An unknown id is ignored. It reports whether the list changed.
func (a *App) ToggleFavorite(id string) bool {
	if _, ok := Lookup(id); !ok {
		return false
	}
	if a.IsFavorite(id) {
		next := make([]string, 0, len(a.Favorites))
		for _, item := range a.Favorites {
			if item != id {
				next = append(next, item)
			}
		}
		a.Favorites = next
		return true
	}
	a.Favorites = append(append([]string{}, a.Favorites...), id)
	return true
}

// UseLists keeps known tool ids from saved lists, dropping duplicates.
func (a *App) UseLists(recent, favorites []string) {
	a.Recent = keepTools(recent, MaxRecent)
	a.Favorites = keepTools(favorites, 0)
}

// keepTools returns known ids in order, without duplicates. A positive
// limit stops after that many.
func keepTools(ids []string, limit int) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		if _, ok := Lookup(id); !ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
		if limit > 0 && len(out) == limit {
			break
		}
	}
	return out
}

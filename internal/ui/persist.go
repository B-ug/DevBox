package ui

import (
	"devbox/internal/config"
	"devbox/internal/model"
)

func persist(a *model.App) {
	if err := config.Save(a.ConfigDir, config.File{
		Theme:     a.Theme,
		Recent:    a.Recent,
		Favorites: a.Favorites,
	}); err != nil {
		a.Status = "失败"
		a.Error = "设置未能保存"
	}
}

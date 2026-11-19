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
		Translate: config.Translate{
			BaseURL: a.Translate.BaseURL,
			APIKey:  a.Translate.APIKey,
			Model:   a.Translate.Model,
			Prompt:  a.Translate.Prompt,
			Target:  a.Translate.Target,
		},
	}); err != nil {
		a.Status = "失败"
		a.Error = "设置未能保存"
	}
}

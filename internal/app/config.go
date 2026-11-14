package app

import (
	"errors"
	"log"
	"os"

	"devbox/internal/config"
	"devbox/internal/model"

	"github.com/egoist/mygo"
)

// prepareIdentity gives an unpackaged run, such as go run, the same name
// the packaged app uses, so its data directory stays DevBox. mygo dev
// ships as its own package, "DevBox Dev", and keeps a separate directory.
func prepareIdentity() {
	if mygo.App.IsPackaged() {
		return
	}
	mygo.App.SetName("DevBox")
	mygo.App.SetVersion("0.1.0")
}

// useConfig loads saved settings from MyGo's user data directory.
func useConfig(a *model.App) {
	dir, err := config.UserDir()
	if err != nil {
		log.Printf("配置目录: %v", err)
		return
	}
	if err := useConfigDir(a, dir); err != nil {
		log.Printf("读取配置: %v", err)
	}
}

// useConfigDir loads settings from dir and remembers it for later saves.
// A missing file keeps the defaults. A corrupt file is reported after the
// directory is remembered, so a later change can replace it.
func useConfigDir(a *model.App, dir string) error {
	a.ConfigDir = dir
	file, err := config.Load(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	switch file.Theme {
	case model.ThemeDark, model.ThemeLight, model.ThemeSystem:
		a.SetTheme(file.Theme)
	}
	a.UseLists(file.Recent, file.Favorites)
	return nil
}

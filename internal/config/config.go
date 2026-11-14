// Package config stores DevBox settings in MyGo's per-user data directory.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

// FileName is the settings file inside PathUserData.
const FileName = "config.json"

// File is the saved settings. New fields can be added without breaking
// older files.
type File struct {
	Theme     string   `json:"theme,omitempty"`
	Recent    []string `json:"recent,omitempty"`
	Favorites []string `json:"favorites,omitempty"`
}

// UserDir is MyGo's application data directory, created on first use.
// On Windows that is %AppData% joined with the application name.
func UserDir() (string, error) {
	return mygo.App.Path(mygo.PathUserData)
}

// Load reads the settings file. A missing file returns os.ErrNotExist.
func Load(dir string) (File, error) {
	b, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		return File{}, err
	}
	var file File
	if err := json.Unmarshal(b, &file); err != nil {
		return File{}, err
	}
	return file, nil
}

// Save writes the settings file. An empty dir skips the write, so tests
// and a failed directory lookup do not touch the disk.
func Save(dir string, file File) error {
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	path := filepath.Join(dir, FileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

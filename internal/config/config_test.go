package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	if err := Save("", File{Theme: "light"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, FileName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("empty dir wrote a file: %v", err)
	}

	if err := Save(dir, File{Theme: "system"}); err != nil {
		t.Fatal(err)
	}
	file, err := Load(dir)
	if err != nil || file.Theme != "system" {
		t.Fatalf("file %+v err %v", file, err)
	}
	if _, err := os.Stat(filepath.Join(dir, FileName+".tmp")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file left behind: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("corrupt file loaded")
	}
}

func TestSaveAndLoadTranslate(t *testing.T) {
	dir := t.TempDir()
	want := Translate{BaseURL: "https://example.test/v1", APIKey: "sk-test", Model: "demo", Prompt: "译成{{target}}", Target: "English"}
	if err := Save(dir, File{Translate: want}); err != nil {
		t.Fatal(err)
	}
	file, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if file.Translate != want {
		t.Fatalf("%+v", file.Translate)
	}
}

func TestSaveAndLoadTunnel(t *testing.T) {
	dir := t.TempDir()
	want := Tunnel{Origin: "http://127.0.0.1:3000", Emails: "a@example.com\n*@example.com", Cloudflared: `C:\Tools\cloudflared.exe`}
	if err := Save(dir, File{Tunnel: want}); err != nil {
		t.Fatal(err)
	}
	file, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if file.Tunnel != want {
		t.Fatalf("%+v", file.Tunnel)
	}
}

func TestSaveAndLoadGold(t *testing.T) {
	dir := t.TempDir()
	want := Gold{Weight: "oz", Currency: "EUR", Window: "5y"}
	if err := Save(dir, File{Gold: want}); err != nil {
		t.Fatal(err)
	}
	file, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if file.Gold != want {
		t.Fatalf("%+v", file.Gold)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load(t.TempDir())
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err %v", err)
	}
}

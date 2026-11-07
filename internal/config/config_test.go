package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadRejectsConfigOutsideDirectory(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG config directory")
	}
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)

	dir, err := EnsureConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	outside := filepath.Join(home, "outside.json")
	if err := os.WriteFile(outside, []byte(`{"project_id":"outside"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(outside, filepath.Join(dir, "config.json")); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("Load accepted config outside its directory")
	}
}

func TestSaveLoadConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG config directory")
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	initial, err := Load()
	if err != nil || initial.Region != "us" {
		t.Fatalf("missing config = %v, %v", initial, err)
	}

	want := Config{ProjectID: "42", Region: "eu"}
	if saveErr := want.Save(); saveErr != nil {
		t.Fatal(saveErr)
	}

	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if *got != want {
		t.Fatalf("Load = %v, want %v", got, want)
	}
}

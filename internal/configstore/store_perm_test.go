package configstore

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveUsesPrivatePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cs := &ConfigFile{Connections: []SavedConnection{{Name: "private", Host: "db.local"}}}
	if err := cs.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := DefaultPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("config mode = %o, want 600", perm)
		}
		dirInfo, err := os.Stat(filepath.Dir(path))
		if err != nil {
			t.Fatal(err)
		}
		if perm := dirInfo.Mode().Perm(); perm != 0o700 {
			t.Fatalf("config dir mode = %o, want 700", perm)
		}
	}
}

func TestSaveTightensExistingFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod semantics differ on Windows")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if err := os.MkdirAll(DefaultDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(DefaultPath(), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	cs := &ConfigFile{Connections: []SavedConnection{{Name: "private"}}}
	if err := cs.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	info, err := os.Stat(DefaultPath())
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("existing config mode = %o, want 600", perm)
	}
}

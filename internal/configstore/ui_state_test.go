package configstore

import (
	"path/filepath"
	"testing"
)

func TestUIStateRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := LoadUIState()
	if err != nil {
		t.Fatalf("LoadUIState() on missing file: %v", err)
	}
	if got != (UIState{}) {
		t.Fatalf("LoadUIState() = %+v, want zero state", got)
	}

	want := UIState{
		LastExportDir: filepath.Join(home, "exports"),
		LastImportDir: filepath.Join(home, "imports"),
	}
	if err := SaveUIState(want); err != nil {
		t.Fatalf("SaveUIState() error = %v", err)
	}
	got, err = LoadUIState()
	if err != nil {
		t.Fatalf("LoadUIState() error = %v", err)
	}
	if got != want {
		t.Fatalf("LoadUIState() = %+v, want %+v", got, want)
	}
}

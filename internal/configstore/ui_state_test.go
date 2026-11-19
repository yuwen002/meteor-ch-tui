package configstore

import (
	"os"
	"path/filepath"
	"reflect"
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
	if !reflect.DeepEqual(got, UIState{}) {
		t.Fatalf("LoadUIState() = %+v, want zero state", got)
	}

	want := UIState{
		LastExportDir: filepath.Join(home, "exports"),
		LastImportDir: filepath.Join(home, "imports"),
		ResultColumnWidths: map[string]map[string]int{
			"name\x1fvalue": {"name": 18, "value": 10},
		},
		ResultHiddenColumns: map[string][]string{
			"name\x1fvalue": {"value"},
		},
	}
	if err := SaveUIState(want); err != nil {
		t.Fatalf("SaveUIState() error = %v", err)
	}
	got, err = LoadUIState()
	if err != nil {
		t.Fatalf("LoadUIState() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadUIState() = %+v, want %+v", got, want)
	}
}

func TestSaveUIStateWriteFailure(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", blocker)
	t.Setenv("USERPROFILE", blocker)

	if err := SaveUIState(UIState{LastExportDir: "x"}); err == nil {
		t.Fatal("SaveUIState() should fail when the home path is a file")
	}
}

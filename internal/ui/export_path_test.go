package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPathInput(t *testing.T) {
	home := isolateTestHome(t)
	t.Setenv("METEOR_TEST_DIR", filepath.Join(home, "from-env"))

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"home", "~", home},
		{"home child", "~/docs", filepath.Join(home, "docs")},
		{"unix env", "$METEOR_TEST_DIR/x", filepath.Join(home, "from-env", "x")},
		{"windows env", "%METEOR_TEST_DIR%/x", filepath.Join(home, "from-env", "x")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := expandPathInput(tc.in)
			if err != nil {
				t.Fatalf("expandPathInput(%q) error = %v", tc.in, err)
			}
			want := tc.want
			if want != "" {
				want = filepath.Clean(want)
			}
			if got != want {
				t.Fatalf("expandPathInput(%q) = %q, want %q", tc.in, got, want)
			}
		})
	}
}

func TestExportFilePathExpandsInput(t *testing.T) {
	home := isolateTestHome(t)
	m := newImportTestModel(t)
	m.initExportPathInput()
	m.exportPathInput.SetValue("~/exports")

	got, err := m.exportFilePath("demo.json")
	if err != nil {
		t.Fatalf("exportFilePath() error = %v", err)
	}
	wantDir := filepath.Join(home, "exports")
	if got != filepath.Join(wantDir, "demo.json") {
		t.Fatalf("exportFilePath() = %q, want %q", got, filepath.Join(wantDir, "demo.json"))
	}
	if info, err := os.Stat(wantDir); err != nil || !info.IsDir() {
		t.Fatalf("export directory stat = %v, isDir = %v", err, err == nil && info.IsDir())
	}
	if m.lastExportDir != wantDir {
		t.Fatalf("lastExportDir = %q, want %q", m.lastExportDir, wantDir)
	}
}

func TestDirBrowserStartsFromExpandedPath(t *testing.T) {
	home := isolateTestHome(t)
	sub := filepath.Join(home, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	m := newImportTestModel(t)
	m.initExportPathInput()
	m.exportPathInput.SetValue("~/sub")
	m.openDirBrowser()

	if m.dirBrowserPath != sub {
		t.Fatalf("dirBrowserPath = %q, want %q", m.dirBrowserPath, sub)
	}
}

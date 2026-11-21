package configstore

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
)

type memoryBackend struct {
	mu    sync.Mutex
	data  map[string]string
	fail  bool
	count int
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{data: map[string]string{}}
}

func (b *memoryBackend) Load(key string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.fail {
		return "", errors.New("backend unavailable")
	}
	v, ok := b.data[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (b *memoryBackend) Store(key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.fail {
		return errors.New("backend unavailable")
	}
	b.data[key] = value
	b.count++
	return nil
}

func (b *memoryBackend) Remove(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.data, key)
	return nil
}

func useMockBackend(t *testing.T, b CredentialBackend) {
	t.Helper()
	SetCredentialBackend(b)
	t.Cleanup(func() { SetCredentialBackend(nil) })
}

func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestSaveExternalizesPasswordToBackend(t *testing.T) {
	isolateHome(t)
	b := newMemoryBackend()
	useMockBackend(t, b)

	cs := &ConfigFile{Connections: []SavedConnection{
		{Name: "prod", Host: "db.local", Password: "s3cr3t"},
	}}
	if err := cs.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(DefaultPath())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "s3cr3t") {
		t.Fatalf("password leaked into config.json: %s", data)
	}
	if !strings.Contains(string(data), KeyringPassword) {
		t.Fatalf("config.json missing keyring placeholder: %s", data)
	}
	if b.data["prod"] != "s3cr3t" {
		t.Fatalf("backend[prod] = %q, want s3cr3t", b.data["prod"])
	}
}

func TestLoadReadsPasswordBackFromBackend(t *testing.T) {
	isolateHome(t)
	b := newMemoryBackend()
	useMockBackend(t, b)

	if err := (&ConfigFile{Connections: []SavedConnection{
		{Name: "prod", Host: "db.local", Password: "s3cr3t"},
	}}).Save(); err != nil {
		t.Fatal(err)
	}
	// 磁盘 config.json 已是占位符，密码只存在于 mock backend 中。
	if b.data["prod"] != "s3cr3t" {
		t.Fatalf("backend should hold the password: %+v", b.data)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Connections) != 1 || got.Connections[0].Password != "s3cr3t" {
		t.Fatalf("loaded password = %q, want s3cr3t", got.Connections[0].Password)
	}
}

func TestLoadMigratesLegacyPlaintext(t *testing.T) {
	home := isolateHome(t)
	b := newMemoryBackend()
	useMockBackend(t, b)

	// 写入旧格式明文配置（模拟升级前的数据）。
	if err := os.MkdirAll(DefaultDir(), 0700); err != nil {
		t.Fatal(err)
	}
	legacy := `{"connections":[{"name":"legacy","host":"db.local","password":"oldpwd"}]}`
	if err := os.WriteFile(DefaultPath(), []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	_ = home

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Connections) != 1 || got.Connections[0].Password != "oldpwd" {
		t.Fatalf("loaded legacy password = %q", got.Connections[0].Password)
	}
	if b.data["legacy"] != "oldpwd" {
		t.Fatalf("legacy password not migrated to backend: %+v", b.data)
	}

	// 再次保存后，明文应从 config.json 中消失。
	if err := got.Save(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(DefaultPath())
	if strings.Contains(string(data), "oldpwd") {
		t.Fatalf("legacy plaintext still in config.json: %s", data)
	}
}

func TestSaveFallsBackToPlaintextWhenBackendFails(t *testing.T) {
	isolateHome(t)
	b := newMemoryBackend()
	b.fail = true
	useMockBackend(t, b)

	cs := &ConfigFile{Connections: []SavedConnection{
		{Name: "prod", Host: "db.local", Password: "s3cr3t"},
	}}
	if err := cs.Save(); err != nil {
		t.Fatalf("Save() should degrade gracefully, got %v", err)
	}
	data, _ := os.ReadFile(DefaultPath())
	if !strings.Contains(string(data), "s3cr3t") {
		t.Fatalf("password not preserved on backend failure: %s", data)
	}
}

func TestSaveRemovesBackendEntryForDeletedConnection(t *testing.T) {
	isolateHome(t)
	b := newMemoryBackend()
	useMockBackend(t, b)

	cs := &ConfigFile{Connections: []SavedConnection{
		{Name: "keep", Host: "a.local", Password: "pw1"},
		{Name: "drop", Host: "b.local", Password: "pw2"},
	}}
	if err := cs.Save(); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.data["drop"]; !ok {
		t.Fatal("drop password should exist before deletion")
	}

	cs.Connections = cs.Connections[:1]
	if err := cs.Save(); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.data["drop"]; ok {
		t.Fatal("backend entry for deleted connection should be removed")
	}
	if b.data["keep"] != "pw1" {
		t.Fatal("kept connection password should remain")
	}
}

func TestOSKeychainBackendRoundTrip(t *testing.T) {
	if !CredentialBackendIsOS() {
		t.Skip("OS backend replaced by test setup")
	}
	key := "meteor-ch-tui-probe-" + t.Name()
	secret := "probe-secret"
	b := osKeychainBackend{}
	_ = b.Remove(key)
	t.Cleanup(func() { _ = b.Remove(key) })

	if err := b.Store(key, secret); err != nil {
		t.Skipf("OS keychain unavailable on this host: %v", err)
	}
	got, err := b.Load(key)
	if err != nil || got != secret {
		t.Fatalf("OS keychain round trip = %q, %v; want %q", got, err, secret)
	}
}

func TestSaveRenamedConnectionMigratesBackendEntry(t *testing.T) {
	isolateHome(t)
	b := newMemoryBackend()
	useMockBackend(t, b)

	cs := &ConfigFile{Connections: []SavedConnection{
		{Name: "old", Host: "db.local", Password: "s3cr3t"},
	}}
	if err := cs.Save(); err != nil {
		t.Fatal(err)
	}

	conn := *cs.Find("old")
	conn.Name = "new"
	if !cs.Replace("old", conn) {
		t.Fatal("Replace failed")
	}
	// 内存连接持有明文密码，Save 以新名写入，并对账删除旧名条目。
	if err := cs.Save(); err != nil {
		t.Fatal(err)
	}
	if b.data["new"] != "s3cr3t" {
		t.Fatalf("backend[new] = %q, want s3cr3t", b.data["new"])
	}
	if _, ok := b.data["old"]; ok {
		t.Fatal("backend entry for old name should be removed after rename")
	}
}

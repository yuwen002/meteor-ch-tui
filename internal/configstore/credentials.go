package configstore

import "github.com/zalando/go-keyring"

// keyringService 是系统密钥环中的服务名，用于隔离本应用与其他程序的凭据。
const keyringService = "meteor-ch-tui"

// CredentialBackend 抽象系统凭据存储，便于在无密钥环环境下降级或在测试中替换。
type CredentialBackend interface {
	Load(key string) (string, error)
	Store(key, value string) error
	Remove(key string) error
}

// osKeychainBackend 基于 go-keyring，按平台映射到：
// Windows Credential Manager、macOS Keychain、Linux Secret Service（D-Bus）。
type osKeychainBackend struct{}

func (osKeychainBackend) Load(key string) (string, error) {
	return keyring.Get(keyringService, key)
}

func (osKeychainBackend) Store(key, value string) error {
	return keyring.Set(keyringService, key, value)
}

func (osKeychainBackend) Remove(key string) error {
	return keyring.Delete(keyringService, key)
}

var activeBackend CredentialBackend = osKeychainBackend{}

// SetCredentialBackend 替换凭据后端，传入 nil 恢复系统默认。仅供测试使用。
func SetCredentialBackend(b CredentialBackend) {
	if b == nil {
		activeBackend = osKeychainBackend{}
		return
	}
	activeBackend = b
}

// CredentialBackendIsOS 报告当前是否使用系统密钥环后端。
func CredentialBackendIsOS() bool {
	_, ok := activeBackend.(osKeychainBackend)
	return ok
}

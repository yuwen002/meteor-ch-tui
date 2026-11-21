package configstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

type SavedConnection struct {
	Name     string `json:"name"`
	Group    string `json:"group"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Protocol string `json:"protocol,omitempty"`
}

// DefaultGroupName 是未填写分组的连接所处的保留分组名。
const DefaultGroupName = "(默认)"

// RedactedPassword 表示导出时已脱敏的密码；覆盖导入时会保留现有密码。
const RedactedPassword = "<redacted>"

// KeyringPassword 表示密码已外置到系统密钥环，config.json 中仅保留此占位符。
// 内存中仍保存明文供连接与导出使用。
const KeyringPassword = "<stored-in-keyring>"

type ConfigFile struct {
	Connections []SavedConnection `json:"connections"`
	LastUsed    string            `json:"last_used"`
	GroupOrder  []string          `json:"group_order,omitempty"`
}

func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".meteor-ch-tui")
}

func DefaultPath() string {
	return filepath.Join(DefaultDir(), "config.json")
}

func (c SavedConnection) ToCfg() ch.Config {
	return ch.Config{
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		Database: c.Database,
		Protocol: ch.NormalizeProtocol(c.Protocol),
	}
}

func FromCfg(name, group string, cfg ch.Config) SavedConnection {
	return SavedConnection{
		Name:     name,
		Group:    group,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
		Protocol: ch.NormalizeProtocol(cfg.Protocol),
	}
}

func Load() (*ConfigFile, error) {
	path := DefaultPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{Connections: []SavedConnection{}}, nil
		}
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	if cfg.Connections == nil {
		cfg.Connections = []SavedConnection{}
	}
	for i := range cfg.Connections {
		pw := cfg.Connections[i].Password
		name := cfg.Connections[i].Name
		switch {
		case pw == KeyringPassword:
			// 密码已外置：从系统密钥环读回明文填入内存。读取失败时保留占位符。
			if secret, err := activeBackend.Load(name); err == nil {
				cfg.Connections[i].Password = secret
			}
		case pw != "" && pw != RedactedPassword:
			// 旧格式明文密码：惰性迁移到密钥环。内存保留明文，下次 Save 会写入占位符。
			_ = activeBackend.Store(name, pw)
		}
	}
	return &cfg, nil
}

func (c *ConfigFile) Save() error {
	dir := DefaultDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	// 已存在的目录/文件不会因 MkdirAll/WriteFile 改变权限，显式收紧以免升级后仍是旧权限。
	// Windows 上 chmod 仅影响只读属性，忽略其错误，避免影响主流程。
	_ = os.Chmod(dir, 0700)

	// 写盘前读取旧配置，用于写成功后对账清理已删除连接的密钥环条目。
	previous := persistedConnectionNames(DefaultPath())

	// 序列化用影子副本：真实密码写入系统密钥环，config.json 仅保留占位符。
	// 密钥环写入失败时保留明文（降级），避免凭据丢失。
	out := *c
	out.Connections = make([]SavedConnection, len(c.Connections))
	copy(out.Connections, c.Connections)
	for i := range out.Connections {
		pw := out.Connections[i].Password
		if pw == "" || pw == KeyringPassword || pw == RedactedPassword {
			continue
		}
		if err := activeBackend.Store(out.Connections[i].Name, pw); err == nil {
			out.Connections[i].Password = KeyringPassword
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&out); err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	data := buf.Bytes()
	path := DefaultPath()
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	_ = os.Chmod(path, 0600)

	// 写盘成功后再清理，避免写失败导致凭据与配置不一致。
	live := make(map[string]bool, len(out.Connections))
	for _, conn := range out.Connections {
		live[conn.Name] = true
	}
	for _, name := range previous {
		if !live[name] {
			_ = activeBackend.Remove(name)
		}
	}
	return nil
}

// persistedConnectionNames 读取当前落盘配置中的连接名；读取或解析失败时返回空。
func persistedConnectionNames(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var old ConfigFile
	if err := json.Unmarshal(data, &old); err != nil {
		return nil
	}
	names := make([]string, 0, len(old.Connections))
	for _, conn := range old.Connections {
		names = append(names, conn.Name)
	}
	return names
}

func (c *ConfigFile) Add(conn SavedConnection) {
	c.Connections = append(c.Connections, conn)
}

func (c *ConfigFile) Remove(name string) bool {
	for i, conn := range c.Connections {
		if conn.Name == name {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			return true
		}
	}
	return false
}

// MoveConnectionsToGroup 把指定连接移动到目标分组，返回实际移动数量。
func (c *ConfigFile) MoveConnectionsToGroup(names []string, group string) int {
	set := map[string]bool{}
	for _, name := range names {
		set[name] = true
	}
	group = strings.TrimSpace(group)
	if group == DefaultGroupName {
		group = ""
	}
	moved := 0
	for i := range c.Connections {
		if set[c.Connections[i].Name] && c.Connections[i].Group != group {
			c.Connections[i].Group = group
			moved++
		}
	}
	return moved
}

// RemoveConnections 批量删除连接，返回实际删除数量。
func (c *ConfigFile) RemoveConnections(names []string) int {
	set := map[string]bool{}
	for _, name := range names {
		set[name] = true
	}
	kept := c.Connections[:0]
	removed := 0
	for _, conn := range c.Connections {
		if set[conn.Name] {
			removed++
			if conn.Name == c.LastUsed {
				c.LastUsed = ""
			}
			continue
		}
		kept = append(kept, conn)
	}
	c.Connections = kept
	return removed
}

func (c *ConfigFile) Update(conn SavedConnection) bool {
	for i, existing := range c.Connections {
		if existing.Name == conn.Name {
			c.Connections[i] = conn
			return true
		}
	}
	return false
}

// Replace 按原始名称定位并替换连接，允许同时修改名称。
func (c *ConfigFile) Replace(oldName string, conn SavedConnection) bool {
	for i := range c.Connections {
		if c.Connections[i].Name == oldName {
			c.Connections[i] = conn
			return true
		}
	}
	return false
}

func (c *ConfigFile) Find(name string) *SavedConnection {
	for i := range c.Connections {
		if c.Connections[i].Name == name {
			return &c.Connections[i]
		}
	}
	return nil
}

func (c *ConfigFile) SetLastUsed(name string) {
	c.LastUsed = name
}

func groupName(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return DefaultGroupName
	}
	return group
}

// Groups 先按 GroupOrder 返回分组，再把未排序的分组按名称补在后面。
// 旧配置没有 GroupOrder 时退化为按名称排序。
func (c *ConfigFile) Groups() []string {
	set := map[string]struct{}{}
	for _, conn := range c.Connections {
		set[groupName(conn.Group)] = struct{}{}
	}
	// 新建的空分组还没有连接，也要保留在列表里。
	for _, g := range c.GroupOrder {
		if strings.TrimSpace(g) != "" {
			set[g] = struct{}{}
		}
	}
	groups := make([]string, 0, len(set))
	seen := map[string]bool{}
	for _, g := range c.GroupOrder {
		if _, ok := set[g]; ok && !seen[g] {
			groups = append(groups, g)
			seen[g] = true
		}
	}
	extra := make([]string, 0, len(set))
	for g := range set {
		if !seen[g] {
			extra = append(extra, g)
		}
	}
	sort.Strings(extra)
	return append(groups, extra...)
}

func (c *ConfigFile) ByGroup(group string) []SavedConnection {
	if group == DefaultGroupName {
		group = ""
	}
	var out []SavedConnection
	for _, conn := range c.Connections {
		if strings.TrimSpace(conn.Group) == group {
			out = append(out, conn)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// AddGroup 新建一个分组；默认组和已存在的分组不能重复创建。
func (c *ConfigFile) AddGroup(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == DefaultGroupName {
		return false
	}
	for _, g := range c.Groups() {
		if g == name {
			return false
		}
	}
	// 以当前顺序为基准追加，让新分组排在末尾而不是跳到最前。
	c.GroupOrder = append(append([]string(nil), c.Groups()...), name)
	return true
}

// RenameGroup 重命名分组，并把该分组下的连接一并迁移。
func (c *ConfigFile) RenameGroup(old, name string) bool {
	old, name = strings.TrimSpace(old), strings.TrimSpace(name)
	if old == "" || old == DefaultGroupName || name == "" || name == DefaultGroupName {
		return false
	}
	found := false
	for _, g := range c.Groups() {
		if g == old {
			found = true
		} else if g == name {
			return false
		}
	}
	if !found {
		return false
	}
	for i := range c.Connections {
		if groupName(c.Connections[i].Group) == old {
			c.Connections[i].Group = name
		}
	}
	for i, g := range c.GroupOrder {
		if g == old {
			c.GroupOrder[i] = name
		}
	}
	return true
}

// RemoveGroup 删除分组，并把该分组下的连接移回默认组（连接本身不删除）。
func (c *ConfigFile) RemoveGroup(name string) bool {
	if name == "" || name == DefaultGroupName {
		return false
	}
	found := false
	for _, g := range c.Groups() {
		if g == name {
			found = true
		}
	}
	if !found {
		return false
	}
	for i := range c.Connections {
		if groupName(c.Connections[i].Group) == name {
			c.Connections[i].Group = ""
		}
	}
	kept := c.GroupOrder[:0]
	for _, g := range c.GroupOrder {
		if g != name {
			kept = append(kept, g)
		}
	}
	c.GroupOrder = kept
	return true
}

// MoveGroup 在显示顺序中上移或下移分组。
func (c *ConfigFile) MoveGroup(name string, delta int) bool {
	groups := c.Groups()
	if len(groups) < 2 {
		return false
	}
	idx := -1
	for i, g := range groups {
		if g == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	next := idx + delta
	if next < 0 || next >= len(groups) {
		return false
	}
	order := append([]string(nil), groups...)
	order[idx], order[next] = order[next], order[idx]
	c.GroupOrder = order
	return true
}

func (c *ConfigFile) AllFlat() []SavedConnection {
	out := make([]SavedConnection, len(c.Connections))
	copy(out, c.Connections)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Clone 返回配置的深拷贝，调用方可以在副本上试算，成功后再替换原对象。
func (c *ConfigFile) Clone() *ConfigFile {
	return &ConfigFile{
		Connections: append([]SavedConnection(nil), c.Connections...),
		LastUsed:    c.LastUsed,
		GroupOrder:  append([]string(nil), c.GroupOrder...),
	}
}

// ExportJSON 导出配置为 JSON；includePasswords=false 时密码替换为 RedactedPassword。
func (c *ConfigFile) ExportJSON(includePasswords bool) (string, error) {
	out := *c
	out.Connections = make([]SavedConnection, len(c.Connections))
	copy(out.Connections, c.Connections)
	if !includePasswords {
		for i := range out.Connections {
			if out.Connections[i].Password != "" {
				out.Connections[i].Password = RedactedPassword
			}
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&out); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

func ImportJSON(data string) (*ConfigFile, error) {
	var cfg ConfigFile
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, fmt.Errorf("导入解析失败: %w", err)
	}
	return normalizeImportedConfig(&cfg)
}

// Merge 合并导入配置；overwrite=true 时同名连接被替换，否则跳过。
// 返回新增和覆盖的连接数。
func (c *ConfigFile) Merge(other *ConfigFile, overwrite bool) (added, updated int) {
	for _, conn := range other.Connections {
		existing := c.Find(conn.Name)
		if existing == nil {
			if conn.Password == RedactedPassword {
				conn.Password = ""
			}
			c.Connections = append(c.Connections, conn)
			added++
			continue
		}
		if !overwrite {
			continue
		}
		if conn.Password == RedactedPassword {
			conn.Password = existing.Password
		}
		*existing = conn
		updated++
	}
	return added, updated
}

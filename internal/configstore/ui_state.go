package configstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// UIState 保存与连接配置无关的界面偏好。
type UIState struct {
	LastExportDir string `json:"last_export_dir,omitempty"`
	LastImportDir string `json:"last_import_dir,omitempty"`
}

func uiStatePath() string {
	return filepath.Join(DefaultDir(), "ui-state.json")
}

// LoadUIState 读取界面偏好；文件不存在时返回零值。
func LoadUIState() (UIState, error) {
	data, err := os.ReadFile(uiStatePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UIState{}, nil
		}
		return UIState{}, fmt.Errorf("读取界面状态失败: %w", err)
	}
	var state UIState
	if err := json.Unmarshal(data, &state); err != nil {
		return UIState{}, fmt.Errorf("解析界面状态失败: %w", err)
	}
	return state, nil
}

// SaveUIState 写入界面偏好。
func SaveUIState(state UIState) error {
	dir := DefaultDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建界面状态目录失败: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化界面状态失败: %w", err)
	}
	if err := os.WriteFile(uiStatePath(), data, 0644); err != nil {
		return fmt.Errorf("写入界面状态失败: %w", err)
	}
	return nil
}

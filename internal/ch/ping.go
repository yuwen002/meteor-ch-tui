package ch

import (
	"context"
	"fmt"
	"strings"
)

// Ping 按给定配置建立一次短连接并返回服务端版本，供连接表单的「测试连接」使用。
// 不修改任何会话状态，连接用完即关。
func Ping(ctx context.Context, cfg Config) (string, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return "", fmt.Errorf("Host 不能为空")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return "", fmt.Errorf("Port 必须在 1-65535 之间")
	}
	client, err := Dial(ctx, cfg, cfg.Addr())
	if err != nil {
		return "", err
	}
	defer client.Close()

	res, err := client.Query(ctx, "SELECT version()", 1)
	if err != nil {
		return "", err
	}
	if len(res.Rows) == 0 || len(res.Rows[0]) == 0 {
		return "", nil
	}
	return FormatValue(res.Rows[0][0]), nil
}

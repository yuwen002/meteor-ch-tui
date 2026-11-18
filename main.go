package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
	"github.com/yuwen002/meteor-ch-tui/internal/configstore"
	"github.com/yuwen002/meteor-ch-tui/internal/ui"
)

var (
	Version   = "dev"
	BuildTime = "-"
	GoVersion = "-"
)

func main() {
	var cfg ch.Config
	var forcePlain bool
	var debug bool
	var connName string
	var quickConnect bool
	var showVersion bool

	flag.StringVar(&cfg.Host, "host", "", "ClickHouse 主机 (可选)")
	flag.IntVar(&cfg.Port, "port", 9000, "ClickHouse Native 端口")
	flag.StringVar(&cfg.Username, "user", "default", "用户名")
	flag.StringVar(&cfg.Password, "password", "", "密码")
	flag.StringVar(&cfg.Database, "database", "", "数据库 (留空使用服务端默认库，仍可浏览账号可见的全部库)")
	flag.StringVar(&cfg.Protocol, "protocol", "native", "连接协议: native(默认, 9000) | http(8123) | https(8443)")
	flag.StringVar(&connName, "name", "", "连接名称 (配合 -host 使用)")
	flag.BoolVar(&quickConnect, "connect", false, "启动直接用命令行参数连接 (跳过管理器)")
	flag.BoolVar(&forcePlain, "plain", false, "强制禁用 AltScreen/鼠标 (用于不支持真 TTY 的环境, 如 GoLand 输出控制台)")
	flag.BoolVar(&debug, "debug", false, "打印调试日志到 stderr")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息并退出")
	flag.Parse()
	if showVersion {
		fmt.Printf("meteor-ch-tui %s (build %s, %s)\n", Version, BuildTime, GoVersion)
		return
	}
	cfg.Protocol = ch.NormalizeProtocol(cfg.Protocol)

	log := func(format string, args ...any) {
		if debug {
			fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
		}
	}

	cs, err := configstore.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "加载配置失败:", err)
		cs = &configstore.ConfigFile{}
	}

	if cfg.Host != "" {
		if connName == "" {
			connName = cfg.Host
		}
		cs.Add(configstore.FromCfg(connName, "命令行", cfg))
		if err := cs.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "保存配置失败:", err)
		}
		if quickConnect {
			cs.SetLastUsed(connName)
			_ = cs.Save()
		}
	}

	stdinIsTTY := isatty.IsTerminal(os.Stdin.Fd())
	stdoutIsTTY := isatty.IsTerminal(os.Stdout.Fd())

	log("stdin.isatty=%v  stdout.isatty=%v  forcePlain=%v", stdinIsTTY, stdoutIsTTY, forcePlain)

	useAltScreen := !forcePlain && stdinIsTTY && stdoutIsTTY
	log("useAltScreen=%v", useAltScreen)

	model := ui.NewModel(cs, stdinIsTTY && stdoutIsTTY)

	if quickConnect {
		if cfg.Host != "" {
			model = ui.NewModelQuickConnect(cfg, cs, stdinIsTTY && stdoutIsTTY)
		} else if conn := cs.Find(cs.LastUsed); conn != nil {
			model = ui.NewModelQuickConnect(conn.ToCfg(), cs, stdinIsTTY && stdoutIsTTY)
		} else {
			fmt.Fprintln(os.Stderr, "-connect: 未找到可用连接，请指定 -host 或先在连接管理器中新建")
		}
	}

	var opts []tea.ProgramOption
	opts = append(opts, tea.WithMouseCellMotion())
	if useAltScreen {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(model, opts...)
	log("tea.NewProgram 完成, 调用 Run()...")

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "运行失败:", err)
		os.Exit(1)
	}

	log("tea.Run() 正常返回, 程序退出")
}

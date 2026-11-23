# 人工验收清单

本清单用于在真实终端中完成 `PLAN.md` 阶段二的「真实终端人工点击验收」。按优先级排列，P0 为本轮新增、最需要验证的部分（`[ ]` 表示待验收，`[x]` 表示已通过）。建议在真实工程机上逐项勾选，并把结果与 `PLAN.md` 的发布前检查清单同步。

## P0：系统密钥环（核心改动）

- [ ] **首次保存不落明文**：新建连接填写密码 → 保存 → 打开 `~/.meteor-ch-tui/config.json`，确认该连接 password 字段是 `<stored-in-keyring>`，搜不到真实密码。
- [ ] **重启回读**：完全退出程序 → 重新打开 → 编辑该连接，确认密码框自动回填正确密码。
- [ ] **密码真能用**：用该连接执行一条 SQL（如 `SELECT 1`），确认认证成功。
- [ ] **凭据条目存在**：Windows 执行 `cmdkey /list`，确认有 `meteor-ch-tui` 相关条目（macOS/Linux 用钥匙串 / Secret Service 工具查）。
  <details><summary>详细验证步骤（点击展开）</summary>

  **Windows（凭据管理器）**
  1. 打开 PowerShell，执行：`cmdkey /list`
  2. 在输出里找 `Target:` 含 `meteor-ch-tui` 的条目。本机实测格式如下（已存在两条，对应你工程机上的真实连接）：
     ```
     Target: LegacyGeneric:target=meteor-ch-tui:临港测试
     Type: Generic
     User: 临港测试
     ```
     其中 `User:` 就是连接名，`target=meteor-ch-tui:<连接名>` 是密钥环键。
  3. 只要保存过带密码的连接，就应有对应条目；保存了 N 个带密码连接，应有 N 条 `meteor-ch-tui` 条目。
  4. （可选）按 `Win+R` 输入 `control /name Microsoft.CredentialManager` 打开图形界面，在「普通凭据」中同样能看到。

  **macOS（钥匙串）**
  - 命令行：`security find-generic-password -s meteor-ch-tui -a <连接名> -w`（能打印出密码即为存在）。
  - 或打开「钥匙串访问」App，搜索 `meteor-ch-tui`。

  **Linux（Secret Service / libsecret）**
  - 若装了 `secret-tool`：`secret-tool lookup service meteor-ch-tui username <连接名>`，输出密码即为存在。
  - 无 `secret-tool` 时，可在有 D-Bus 的桌面环境用「密码和密钥」（Seahorse）查看。

  > 注意：无头 Linux 无 D-Bus/Secret Service 时会走降级路径（见第 14 项），此时**不会有**密钥环条目，密码以明文留在 config.json，属预期行为。

  </details>
- [ ] **重命名迁移**：把连接改名 → 检查凭据管理器，确认新名有条目、旧名条目已清除。
- [ ] **删除清理**：删除连接 → 确认凭据管理器中对应条目被移除。
- [ ] **明文导出确认**：导出弹窗按 `Tab` 开启「包含密码」→ 应弹「安全提醒」→ 按 `Esc` 保持关闭、按 `Enter` 开启；开启后导出的文件确实含明文密码。
- [ ] **降级路径**（可选，需无密钥环环境）：在 Linux 无头容器里保存带密码连接，确认降级为 config.json 明文且功能不中断。
  <details><summary>详细验证步骤（点击展开）</summary>

  目标：模拟「系统密钥环不可用」，确认程序**不崩溃、不丢密码**，自动降级为把明文写回 config.json。

  **方式一：WSL（本机可用，`docker-desktop` 发行版无 D-Bus，天然满足）**
  1. 交叉编译 Linux 版（在项目根目录 PowerShell 执行）：
     ```powershell
     $env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
     go build -o meteor-ch-tui-linux ./...
     Remove-Item Env:\GOOS, Env:\GOARCH, Env:\CGO_ENABLED
     ```
  2. 把二进制送进 WSL 并运行（WSL 无 `/mnt/c` 挂载时用管道传输）：
     ```powershell
     Get-Content .\meteor-ch-tui-linux -Encoding Byte -Raw | `
       wsl -d docker-desktop -- sh -c 'cat > /tmp/mct && chmod +x /tmp/mct && HOME=/tmp/mct-home /tmp/mct'
     ```
  3. 在 WSL 里新建带密码的连接并保存，然后查看 `/tmp/mct-home/.meteor-ch-tui/config.json`：应能看到 `password` 字段是**明文**（不是 `<stored-in-keyring>`），且退出重开后密码仍在、连接可用。
  4. 清理：`wsl -d docker-desktop -- rm -f /tmp/mct`。

  **方式二：Docker 容器（更接近 CI/服务器）**
  1. 启动一个无 D-Bus 的 Linux 容器并挂载项目编译产物：
     ```powershell
     docker run -it --rm -v "$PWD:/work" alpine:3.20 sh
     ```
  2. 容器内运行 `/work/meteor-ch-tui-linux`，同样新建带密码连接并保存。
  3. `cat /root/.meteor-ch-tui/config.json` 确认密码为明文、`<stored-in-keyring>` 占位符不出现；重启进程确认密码可读回。

  **方式三：只需验证逻辑（无需 Linux）**
  - 单元测试已覆盖该分支：`internal/configstore/credentials_test.go` 的 `TestSaveFallsBackToPlaintextWhenBackendFails` 用 `fail=true` 的 mock 后端模拟密钥环不可用，断言 `Save()` 不返回错误且密码以明文留在 config.json。运行：`go test ./internal/configstore/... -run TestSaveFallsBackToPlaintextWhenBackendFails -count=1 -v`。

  **预期结果（三种方式一致）**
  - 程序不崩溃、保存成功；
  - config.json 中该连接 password 为**真实明文**（非占位符）；
  - 退出重开后编辑该连接，密码框自动回填、连接认证成功；
  - 密钥环中**不会**产生对应条目。

  > 说明：降级是刻意设计——无头/CI 环境没有 Secret Service，优先保证功能可用。明文文件的权限仍为 `0600`（目录 `0700`）。因此生产部署建议提供 D-Bus/密钥环，或接受明文存储风险。

  </details>

## P0：批量连接操作

- [ ] **标记显示**：列表按 `空格` 标记多个连接，行首出现 `[x]`。
- [ ] **全选**：`Ctrl+A` 后所有连接都被标记。
- [ ] **批量移动**：`m` 打开弹窗 → ↑↓ 选分组 → `Enter`，确认连接已移到目标组，标记清空。
- [ ] **批量删除**：标记 → `Del` → 确认框显示「确定删除选中的 N 个连接」→ `Enter` 删除；`Esc` 取消则不变。
- [ ] **失败不变**：写操作保存失败时，内存状态不被污染。
- [ ] **弹出组选择列表**：新建/编辑连接时分组字段按 `Ctrl+G`（或点击输入框）弹出分组列表，显示已有分组及数量，`↑`/`↓` 选择、`Enter` 填入、`Esc` 取消，选中项高亮；表单下方不再铺列表。批量移动（`m`）与分组管理（`g`）弹窗可见全部已有分组。

## P1：查询结果列设置

- [ ] SQL 用 `Ctrl+S` 或 `F5` 执行成功；`Alt+Enter` 不再触发执行（避免终端全屏冲突）。
- [ ] 结果表格按 `c` 打开列设置，Space 隐藏列、`[` / `]` 调列宽、`R` 重置、`A` 全显。
- [ ] 隐藏列**只影响显示**，导出的 TSV/CSV/JSON/XLSX 仍包含该列。
- [ ] 退出重开后列显示 / 列宽设置保持不变。

## P1：导入导出与浏览器

- [ ] 四种格式（JSON/CSV/TSV/XLSX）导出文件可正常打开；损坏 XLSX 导入显示失败状态。
- [ ] 文件浏览器 `/` 搜索、`s` 切换名称 / 时间排序，重启后排序模式保持。
- [ ] 路径支持 `~`、`$VAR`、`%VAR%`；目录浏览器可切换保存目录。

## P1：基础回归

- [ ] 新建 / 编辑 / 删除连接、分组管理（`g`）、测试连接（`Ctrl+T`）。
- [ ] `?` / `F1` 帮助浮层可打开，作者「浮笙」与邮箱「yuwen002@163.com」显示正确。
- [ ] `./meteor-ch-tui.exe -version` 输出版本与 `git rev-parse --short HEAD` 一致且无 `-dirty`。

## 验收后动作

全部通过后，把 `PLAN.md` 阶段二第 57 行「真实终端人工点击验收」勾选为 `[x]`，计划即全部完成。

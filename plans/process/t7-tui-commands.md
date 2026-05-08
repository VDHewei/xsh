# T7: TUI 自定义 Commands + Config + i18n - 完成

**完成时间**: 2026-05-08

## 子任务进度

| 子任务 | 状态 | 通过/总数 | 备注 |
|--------|------|-----------|------|
| T7.1 Types (CustomCommand) | 已完成 | - | internal/types/types.go |
| T7.2 CommandLoader | 已完成 | - | pkg/llm/command_loader.go |
| T7.3 Example commands | 已完成 | - | commands/deploy.md, health.md, migration.md |
| T7.4 Config module | 已完成 | 5/5 | internal/config/config.go |
| T7.5 Locale detection | 已完成 | - | internal/config/locale.go + locale_unix.go + locale_windows.go |
| T7.6 i18n module | 已完成 | 3/3 | internal/i18n/i18n.go |
| T7.7 Default config file | 已完成 | - | internal/config/default.toml (embedded) |
| T7.8 TUI refactoring | 已完成 | - | Two-column layout + commands + config-driven |
| T7.9 Tests | 已完成 | 22/22 | config(6), i18n(4), tui(4), existing(8) |
| T7.10 Progress files | 已完成 | - | 本文件 |

## 新增文件

| 文件 | 描述 |
|------|------|
| `internal/config/config.go` | Config, ColorScheme, LayoutConfig, I18NConfig structs; Load() with priority cwd>home>install |
| `internal/config/default.toml` | 嵌入式默认配置 (zh-CN) |
| `internal/config/locale.go` | DetectLocale(), mapLocale() |
| `internal/config/locale_unix.go` | Unix locale detection via $LANG |
| `internal/config/locale_windows.go` | Windows locale detection via GetUserDefaultLocaleName |
| `internal/config/config_test.go` | 6 tests: defaults, locale, mapping, merge, paths, mergeString |
| `internal/i18n/i18n.go` | i18n.Manager with T() and convenience methods |
| `internal/i18n/i18n_test.go` | 4 tests: New, static strings, format strings |
| `internal/tui/tui_test.go` | 4 tests: NewApp, parseColor, errTag, okTag |
| `.xsh.toml` | 项目级默认配置文件 |

## 修改文件

| 文件 | 变更 |
|------|------|
| `internal/types/types.go` | 添加 CustomCommand + CommandLoader interface |
| `internal/tui/tui.go` | 完全重构: config-driven, i18n, 两列布局, 命令列表, Tab焦点 |
| `go.mod` | 添加 github.com/BurntSushi/toml v1.5.0 |

## 配置加载优先级

1. `./xsh.toml` (当前工作目录)
2. `~/.xsh.toml` (用户家目录)
3. `<install_dir>/.xsh.toml` (xsh 安装目录)

先找到的优先，支持部分覆盖默认值。

## 支持的语言

| 语言 | 检测方式 |
|------|----------|
| zh-CN (简体中文, 默认) | Windows: GetUserDefaultLocaleName; Unix: $LANG |
| zh-TW (繁体中文) | 同上 |
| en (英文) | 同上 |

## TUI 布局

```
+----------------------+----------------------------------+
|         header (span 2, from i18n)                         |
+----------------------+----------------------------------+
| cmdList (25 chars)   | taskList                           |
+----------------------+----------------------------------+
|                      | progress                           |
+----------------------+----------------------------------+
|                      | inputLabel | inputField             |
+----------------------+----------------------------------+
```

## 焦点管理

| 按键 | 操作 |
|------|------|
| Tab | 切换焦点: inputField ↔ commandList |
| Enter (commandList) | 加载选中命令 |
| Enter (task area) | 执行下一个任务 |

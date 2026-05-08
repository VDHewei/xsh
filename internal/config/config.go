package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

//go:embed default.toml
var defaultConfigData []byte

// ColorScheme UI 颜色配置
type ColorScheme struct {
	Header    string `toml:"header"`
	Success   string `toml:"success"`
	Error     string `toml:"error"`
	Progress  string `toml:"progress"`
	Border    string `toml:"border"`
	Selection string `toml:"selection"`
}

// LayoutConfig 布局配置
type LayoutConfig struct {
	CommandListWidth int `toml:"command_list_width"`
}

// I18NConfig i18n 字符串配置
type I18NConfig struct {
	Header           string `toml:"header"`
	TaskListTitle    string `toml:"task_list_title"`
	ProgressTitle    string `toml:"progress_title"`
	InputLabel       string `toml:"input_label"`
	InputTitle       string `toml:"input_title"`
	InputPlaceholder string `toml:"input_placeholder"`
	CmdListTitle     string `toml:"cmd_list_title"`
	WaitingText      string `toml:"waiting_text"`
	TasksLoadedFmt   string `toml:"tasks_loaded_fmt"`
	TasksLoadedCmdFmt string `toml:"tasks_loaded_cmd_fmt"`
	TaskExecutingFmt string `toml:"task_executing_fmt"`
	AllDoneText      string `toml:"all_done_text"`
	InvalidPathText  string `toml:"invalid_path_text"`
	ParseFailedFmt   string `toml:"parse_failed_fmt"`
	ContinueButton   string `toml:"continue_button"`
	SkipButton       string `toml:"skip_button"`
	ExitButton       string `toml:"exit_button"`
	AskTitle         string `toml:"ask_title"`
	CheckTitle       string `toml:"check_title"`
}

// Config 应用配置
type Config struct {
	Language string       `toml:"language"`
	Style    ColorScheme  `toml:"style"`
	Layout   LayoutConfig `toml:"layout"`
	I18N     I18NConfig   `toml:"i18n"`
}

// defaults 返回默认配置 (zh-CN)
func defaults() *Config {
	return &Config{
		Language: "",
		Style: ColorScheme{
			Header:    "green",
			Success:   "green",
			Error:     "red",
			Progress:  "yellow",
			Border:    "white",
			Selection: "cyan",
		},
		Layout: LayoutConfig{
			CommandListWidth: 25,
		},
		I18N: I18NConfig{
			Header: "xsh - 任务执行工具",
			TaskListTitle:     "任务列表",
			ProgressTitle:     "执行进度",
			InputLabel:        "加载任务文件: ",
			InputTitle:        "输入文件路径",
			InputPlaceholder:  "请输入任务文件路径 (如: tests/data/prod-migration-form-uat.txt)",
			CmdListTitle:      "命令列表",
			WaitingText:       "等待加载任务文件...",
			TasksLoadedFmt:    "已加载 %d 个任务，按 Enter 开始执行",
			TasksLoadedCmdFmt: "已加载 command: %s, %d 个任务",
			TaskExecutingFmt:  "正在执行 [%d/%d]: %s",
			AllDoneText:       "所有任务执行完成!",
			InvalidPathText:   "请输入文件路径",
			ParseFailedFmt:    "解析文件失败: %v",
			ContinueButton:    "继续",
			SkipButton:        "跳过",
			ExitButton:        "退出",
			AskTitle:          "Ask LLM 分析",
			CheckTitle:        "Check 检查结果",
		},
	}
}

// Load 按优先级加载配置: cwd > home > install dir
func Load() *Config {
	cfg := defaults()

	// 自动检测语言
	if cfg.Language == "" {
		cfg.Language = DetectLocale()
	}

	// 按优先级查找配置文件
	paths := configPaths()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := cfg.MergeFromFile(p); err == nil {
				return cfg
			}
		}
	}

	return cfg
}

// MergeFromFile 从文件合并配置, 覆盖默认值
func (c *Config) MergeFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	var fileCfg Config
	if err := toml.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("parse config file %s: %w", path, err)
	}

	// 合并 Style
	if fileCfg.Style.Header != "" {
		c.Style.Header = fileCfg.Style.Header
	}
	if fileCfg.Style.Success != "" {
		c.Style.Success = fileCfg.Style.Success
	}
	if fileCfg.Style.Error != "" {
		c.Style.Error = fileCfg.Style.Error
	}
	if fileCfg.Style.Progress != "" {
		c.Style.Progress = fileCfg.Style.Progress
	}
	if fileCfg.Style.Border != "" {
		c.Style.Border = fileCfg.Style.Border
	}
	if fileCfg.Style.Selection != "" {
		c.Style.Selection = fileCfg.Style.Selection
	}

	// 合并 Layout
	if fileCfg.Layout.CommandListWidth > 0 {
		c.Layout.CommandListWidth = fileCfg.Layout.CommandListWidth
	}

	// 合并 Language
	if fileCfg.Language != "" {
		c.Language = fileCfg.Language
	}

	// 合并 I18N
	mergeString(&c.I18N.Header, fileCfg.I18N.Header)
	mergeString(&c.I18N.TaskListTitle, fileCfg.I18N.TaskListTitle)
	mergeString(&c.I18N.ProgressTitle, fileCfg.I18N.ProgressTitle)
	mergeString(&c.I18N.InputLabel, fileCfg.I18N.InputLabel)
	mergeString(&c.I18N.InputTitle, fileCfg.I18N.InputTitle)
	mergeString(&c.I18N.InputPlaceholder, fileCfg.I18N.InputPlaceholder)
	mergeString(&c.I18N.CmdListTitle, fileCfg.I18N.CmdListTitle)
	mergeString(&c.I18N.WaitingText, fileCfg.I18N.WaitingText)
	mergeString(&c.I18N.TasksLoadedFmt, fileCfg.I18N.TasksLoadedFmt)
	mergeString(&c.I18N.TasksLoadedCmdFmt, fileCfg.I18N.TasksLoadedCmdFmt)
	mergeString(&c.I18N.TaskExecutingFmt, fileCfg.I18N.TaskExecutingFmt)
	mergeString(&c.I18N.AllDoneText, fileCfg.I18N.AllDoneText)
	mergeString(&c.I18N.InvalidPathText, fileCfg.I18N.InvalidPathText)
	mergeString(&c.I18N.ParseFailedFmt, fileCfg.I18N.ParseFailedFmt)
	mergeString(&c.I18N.ContinueButton, fileCfg.I18N.ContinueButton)
	mergeString(&c.I18N.SkipButton, fileCfg.I18N.SkipButton)
	mergeString(&c.I18N.ExitButton, fileCfg.I18N.ExitButton)
	mergeString(&c.I18N.AskTitle, fileCfg.I18N.AskTitle)
	mergeString(&c.I18N.CheckTitle, fileCfg.I18N.CheckTitle)

	return nil
}

func mergeString(target *string, source string) {
	if source != "" {
		*target = source
	}
}

// configPaths 返回配置文件搜索路径 (按优先级)
func configPaths() []string {
	var paths []string

	// 1. 当前目录
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".xsh.toml"))
	}

	// 2. 用户家目录
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".xsh.toml"))
	}

	// 3. xsh 可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		installDir := filepath.Dir(execPath)
		// 处理 Windows 符号链接
		if runtime.GOOS == "windows" {
			if resolved, err := filepath.EvalSymlinks(installDir); err == nil {
				installDir = resolved
			}
		}
		paths = append(paths, filepath.Join(installDir, ".xsh.toml"))
	}

	return paths
}

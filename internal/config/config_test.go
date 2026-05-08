package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.Style.Header != "green" {
		t.Errorf("expected header color green, got %s", cfg.Style.Header)
	}
	if cfg.Style.Success != "green" {
		t.Errorf("expected success color green, got %s", cfg.Style.Success)
	}
	if cfg.Style.Error != "red" {
		t.Errorf("expected error color red, got %s", cfg.Style.Error)
	}
	if cfg.Layout.CommandListWidth != 25 {
		t.Errorf("expected command list width 25, got %d", cfg.Layout.CommandListWidth)
	}
	if cfg.I18N.Header != "xsh - \u4efb\u52a1\u6267\u884c\u5de5\u5177" {
		t.Errorf("expected chinese header, got %s", cfg.I18N.Header)
	}
	if cfg.I18N.TasksLoadedFmt != "\u5df2\u52a0\u8f7d %d \u4e2a\u4efb\u52a1\uff0c\u6309 Enter \u5f00\u59cb\u6267\u884c" {
		t.Errorf("expected chinese task loaded format, got %s", cfg.I18N.TasksLoadedFmt)
	}
}

func TestDetectLocale(t *testing.T) {
	locale := DetectLocale()
	if locale == "" {
		t.Error("DetectLocale returned empty string")
	}
	// Must be one of supported languages
	switch locale {
	case "zh-CN", "zh-TW", "en":
		// valid
	default:
		t.Errorf("unexpected locale: %s", locale)
	}
}

func TestMapLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"zh-CN", "zh-CN"},
		{"zh_CN", "zh-CN"},
		{"zh-Hans-CN", "zh-CN"},
		{"zh-cn", "zh-CN"},
		{"zh-TW", "zh-TW"},
		{"zh_tw", "zh-TW"},
		{"zh-Hant-TW", "zh-TW"},
		{"zh-HK", "zh-TW"},
		{"en-US", "en"},
		{"en_GB", "en"},
		{"en", "en"},
		{"ja-JP", "zh-CN"},   // fallback to default
		{"ko-KR", "zh-CN"},   // fallback to default
		{"unknown", "zh-CN"}, // fallback to default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapLocale(tt.input)
			if result != tt.expected {
				t.Errorf("mapLocale(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMergeFromFile(t *testing.T) {
	cfg := defaults()

	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".xsh.toml")

	content := []byte(`
language = "en"

[style]
header = "cyan"
success = "blue"

[layout]
command_list_width = 40

[i18n]
header = "xsh - Task Tool"
task_list_title = "Tasks"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	if err := cfg.MergeFromFile(configPath); err != nil {
		t.Fatalf("MergeFromFile: %v", err)
	}

	// Language should be updated
	if cfg.Language != "en" {
		t.Errorf("expected language en, got %s", cfg.Language)
	}

	// Style should be merged
	if cfg.Style.Header != "cyan" {
		t.Errorf("expected header cyan, got %s", cfg.Style.Header)
	}
	if cfg.Style.Success != "blue" {
		t.Errorf("expected success blue, got %s", cfg.Style.Success)
	}
	// Error should remain default (not set in file)
	if cfg.Style.Error != "red" {
		t.Errorf("expected error red (default), got %s", cfg.Style.Error)
	}

	// Layout should be merged
	if cfg.Layout.CommandListWidth != 40 {
		t.Errorf("expected command list width 40, got %d", cfg.Layout.CommandListWidth)
	}

	// I18N should be partially merged
	if cfg.I18N.Header != "xsh - Task Tool" {
		t.Errorf("expected i18n header 'xsh - Task Tool', got %s", cfg.I18N.Header)
	}
	if cfg.I18N.TaskListTitle != "Tasks" {
		t.Errorf("expected task list title 'Tasks', got %s", cfg.I18N.TaskListTitle)
	}
	// Unset i18n keys should remain default
	if cfg.I18N.ProgressTitle != "\u6267\u884c\u8fdb\u5ea6" {
		t.Errorf("expected progress title unchanged, got %s", cfg.I18N.ProgressTitle)
	}
}

func TestConfigPaths(t *testing.T) {
	paths := configPaths()
	if len(paths) != 3 {
		t.Errorf("expected 3 config paths, got %d", len(paths))
	}
	// All paths should end with .xsh.toml
	for _, p := range paths {
		if filepath.Base(p) != ".xsh.toml" {
			t.Errorf("expected path ending with .xsh.toml, got %s", p)
		}
	}
}

func TestMergeString(t *testing.T) {
	target := "original"
	mergeString(&target, "")
	if target != "original" {
		t.Error("empty source should not override")
	}
	mergeString(&target, "new")
	if target != "new" {
		t.Error("non-empty source should override")
	}
}

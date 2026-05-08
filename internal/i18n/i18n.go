package i18n

import (
	"fmt"

	"github.com/VDHewei/xsh/internal/config"
)

// Manager i18n 字符串管理器
type Manager struct {
	lang    string
	strings config.I18NConfig
}

// New 创建 i18n 管理器
func New(lang string, strings config.I18NConfig) *Manager {
	return &Manager{
		lang:    lang,
		strings: strings,
	}
}

// Lang 返回当前语言 tag
func (m *Manager) Lang() string {
	return m.lang
}

// T 返回翻译后的字符串 (当前实现: 所有语言使用同一组 config 中的字符串)
// 当 config 支持多语言时, 此处根据 m.lang 选择对应语言的字符串
func (m *Manager) T(key string) string {
	return fmt.Sprintf("{{%s}}", key)
}

// --- 便捷方法 ---

func (m *Manager) Header() string            { return m.strings.Header }
func (m *Manager) TaskListTitle() string       { return m.strings.TaskListTitle }
func (m *Manager) ProgressTitle() string       { return m.strings.ProgressTitle }
func (m *Manager) InputLabel() string          { return m.strings.InputLabel }
func (m *Manager) InputTitle() string          { return m.strings.InputTitle }
func (m *Manager) InputPlaceholder() string    { return m.strings.InputPlaceholder }
func (m *Manager) CmdListTitle() string        { return m.strings.CmdListTitle }
func (m *Manager) WaitingText() string         { return m.strings.WaitingText }
func (m *Manager) AllDoneText() string         { return m.strings.AllDoneText }
func (m *Manager) InvalidPathText() string     { return m.strings.InvalidPathText }
func (m *Manager) ContinueButton() string      { return m.strings.ContinueButton }
func (m *Manager) SkipButton() string          { return m.strings.SkipButton }
func (m *Manager) ExitButton() string          { return m.strings.ExitButton }
func (m *Manager) AskTitle() string            { return m.strings.AskTitle }
func (m *Manager) CheckTitle() string          { return m.strings.CheckTitle }

func (m *Manager) TasksLoadedFmt(count int) string {
	return fmt.Sprintf(m.strings.TasksLoadedFmt, count)
}

func (m *Manager) TasksLoadedCmdFmt(name string, count int) string {
	return fmt.Sprintf(m.strings.TasksLoadedCmdFmt, name, count)
}

func (m *Manager) TaskExecutingFmt(current, total int, raw string) string {
	return fmt.Sprintf(m.strings.TaskExecutingFmt, current, total, raw)
}

func (m *Manager) ParseFailedFmt(err error) string {
	return fmt.Sprintf(m.strings.ParseFailedFmt, err)
}

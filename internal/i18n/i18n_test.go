package i18n

import (
	"testing"

	"github.com/VDHewei/xsh/internal/config"
)

func defaultI18NConfig() config.I18NConfig {
	return config.I18NConfig{
		Header:           "xsh - Test Tool",
		TaskListTitle:    "Task List",
		ProgressTitle:    "Progress",
		InputLabel:       "Load: ",
		InputTitle:       "Input",
		InputPlaceholder: "Enter path",
		CmdListTitle:     "Commands",
		WaitingText:      "Waiting...",
		TasksLoadedFmt:   "%d tasks loaded",
		TasksLoadedCmdFmt: "Command %s: %d tasks",
		TaskExecutingFmt: "[%d/%d] %s",
		AllDoneText:      "All done!",
		InvalidPathText:  "Invalid path",
		ParseFailedFmt:   "Parse error: %v",
		ContinueButton:   "Continue",
		SkipButton:       "Skip",
		ExitButton:       "Exit",
		AskTitle:         "Ask",
		CheckTitle:       "Check",
	}
}

func TestNew(t *testing.T) {
	cfg := defaultI18NConfig()
	m := New("en", cfg)

	if m.Lang() != "en" {
		t.Errorf("expected lang en, got %s", m.Lang())
	}
}

func TestStaticStrings(t *testing.T) {
	cfg := defaultI18NConfig()
	m := New("en", cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Header", m.Header(), "xsh - Test Tool"},
		{"TaskListTitle", m.TaskListTitle(), "Task List"},
		{"ProgressTitle", m.ProgressTitle(), "Progress"},
		{"InputLabel", m.InputLabel(), "Load: "},
		{"InputTitle", m.InputTitle(), "Input"},
		{"InputPlaceholder", m.InputPlaceholder(), "Enter path"},
		{"CmdListTitle", m.CmdListTitle(), "Commands"},
		{"WaitingText", m.WaitingText(), "Waiting..."},
		{"AllDoneText", m.AllDoneText(), "All done!"},
		{"InvalidPathText", m.InvalidPathText(), "Invalid path"},
		{"ContinueButton", m.ContinueButton(), "Continue"},
		{"SkipButton", m.SkipButton(), "Skip"},
		{"ExitButton", m.ExitButton(), "Exit"},
		{"AskTitle", m.AskTitle(), "Ask"},
		{"CheckTitle", m.CheckTitle(), "Check"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestFormatStrings(t *testing.T) {
	cfg := defaultI18NConfig()
	m := New("en", cfg)

	t.Run("TasksLoadedFmt", func(t *testing.T) {
		got := m.TasksLoadedFmt(5)
		if got != "5 tasks loaded" {
			t.Errorf("got %q, want '5 tasks loaded'", got)
		}
	})

	t.Run("TasksLoadedCmdFmt", func(t *testing.T) {
		got := m.TasksLoadedCmdFmt("deploy", 3)
		if got != "Command deploy: 3 tasks" {
			t.Errorf("got %q, want 'Command deploy: 3 tasks'", got)
		}
	})

	t.Run("TaskExecutingFmt", func(t *testing.T) {
		got := m.TaskExecutingFmt(2, 5, "[GET] /health")
		if got != "[2/5] [GET] /health" {
			t.Errorf("got %q, want '[2/5] [GET] /health'", got)
		}
	})
}

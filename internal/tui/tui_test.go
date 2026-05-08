package tui

import (
	"testing"

	"github.com/VDHewei/xsh/internal/config"
	"github.com/gdamore/tcell/v2"
)

func TestNewApp(t *testing.T) {
	cfg := config.Load()
	app := NewApp(cfg)

	if app.config != cfg {
		t.Error("config not set correctly")
	}
	if app.i18n == nil {
		t.Error("i18n manager not initialized")
	}
	if app.exec == nil {
		t.Error("executor not initialized")
	}
	if app.commandLoader == nil {
		t.Error("commandLoader not initialized")
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name     string
		expected tcell.Color
	}{
		{"red", tcell.ColorRed},
		{"green", tcell.ColorGreen},
		{"yellow", tcell.ColorYellow},
		{"blue", tcell.ColorBlue},
		{"magenta", tcell.ColorPurple},
		{"cyan", tcell.ColorLightCyan},
		{"white", tcell.ColorWhite},
		{"orange", tcell.ColorOrange},
		{"unknown", tcell.ColorWhite}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseColor(tt.name)
			if got != tt.expected {
				t.Errorf("parseColor(%q) = %d, want %d", tt.name, got, tt.expected)
			}
		})
	}
}

func TestErrTag(t *testing.T) {
	cfg := config.Load()
	cfg.Style.Error = "red"
	app := NewApp(cfg)

	got := app.errTag("error message")
	expected := "[red]error message[red]"
	if got != expected {
		t.Errorf("errTag = %q, want %q", got, expected)
	}
}

func TestOkTag(t *testing.T) {
	cfg := config.Load()
	cfg.Style.Success = "green"
	app := NewApp(cfg)

	got := app.okTag("success message")
	expected := "[green]success message[green]"
	if got != expected {
		t.Errorf("okTag = %q, want %q", got, expected)
	}
}

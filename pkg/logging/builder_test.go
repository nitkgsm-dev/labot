package logging

import (
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"
)

func Test_LogLevel_ToSlogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		l        LogLevel
		expected slog.Level
	}{
		{
			name:     "LevelDebug",
			l:        LevelDebug,
			expected: slog.LevelDebug,
		},
		{
			name:     "LevelInfo",
			l:        LevelInfo,
			expected: slog.LevelInfo,
		},
		{
			name:     "LevelWarn",
			l:        LevelWarn,
			expected: slog.LevelWarn,
		},
		{
			name:     "LevelError",
			l:        LevelError,
			expected: slog.LevelError,
		},
		{
			name:     "unknown",
			l:        10,
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.l.ToSlogLevel()
			if actual != tt.expected {
				t.Errorf("expected %v, but received %v", tt.expected, actual)
			}
		})
	}
}

func Test_DefaultBuilder(t *testing.T) {
	t.Parallel()

	actual := DefaultBuilder()
	if actual.level != LevelInfo {
		t.Errorf("expected level is LevelInfo, but received %v", actual.level)
	}
	if actual.dateFormat != time.DateTime {
		t.Errorf("expected dateFormat is time.DateTime, but received %v", actual.dateFormat)
	}
	if actual.logFormat != FormatText {
		t.Errorf("expected logFormat is FormatText, but received %v", actual.logFormat)
	}
	if !reflect.DeepEqual(actual.writer, os.Stdout) {
		t.Errorf("expected writer is os.Stdout, but received %T", actual.writer)
	}
	if actual.displaySource {
		t.Errorf("expected displaySource is false, but received %v", actual.displaySource)
	}
}

func Test_Builder_SetLevel(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	b = b.SetLevel(LevelDebug)
	if b.level != LevelDebug {
		t.Errorf("expected level is LevelDebug, but received %v", b.level)
	}
}

func Test_Builder_SetDateFormat(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	b = b.SetDateFormat("2006-01-02")
	if b.dateFormat != "2006-01-02" {
		t.Errorf("expected dateFormat is 2006-01-02, but received %v", b.dateFormat)
	}
}

func Test_Builder_SetLogFormat(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	b = b.SetLogFormat(FormatJSON)
	if b.logFormat != FormatJSON {
		t.Errorf("expected logFormat is FormatJSON, but received %v", b.logFormat)
	}
}

func Test_Builder_SetWriter(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	b = b.SetWriter(os.Stderr)
	if !reflect.DeepEqual(b.writer, os.Stderr) {
		t.Errorf("expected writer is os.Stderr, but received %T", b.writer)
	}
}

func Test_Builder_SetDisplaySource(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	b = b.SetDisplaySource(true)
	if !b.displaySource {
		t.Errorf("expected displaySource is true, but received %v", b.displaySource)
	}
}

func Test_Builder_Build(t *testing.T) {
	t.Parallel()

	jsonLogger := DefaultBuilder().SetLogFormat(FormatJSON).Build()
	if jsonLogger == nil {
		t.Error("expected jsonLogger is not nil, but received nil")
	}
	textLogger := DefaultBuilder().SetLogFormat(FormatText).Build()
	if textLogger == nil {
		t.Error("expected textLogger is not nil, but received nil")
	}
	if reflect.DeepEqual(jsonLogger, textLogger) {
		t.Error("expected jsonLogger and textLogger are different, but received same")
	}
}

package logging

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/m-mizutani/clog"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) ToSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type logFormat int

const (
	FormatText logFormat = iota
	FormatJSON
)

// Builder builds a slog.
type Builder struct {
	level         LogLevel
	dateFormat    string
	logFormat     logFormat
	displaySource bool
	writer        io.Writer
}

func DefaultBuilder() *Builder {
	return &Builder{
		level:         LevelInfo,
		dateFormat:    time.DateTime,
		displaySource: false,
		logFormat:     FormatText,
		writer:        os.Stdout,
	}
}

func (b *Builder) SetLevel(level LogLevel) *Builder {
	b.level = level
	return b
}

func (b *Builder) SetDateFormat(format string) *Builder {
	b.dateFormat = format
	return b
}

func (b *Builder) SetLogFormat(format logFormat) *Builder {
	b.logFormat = format
	return b
}

func (b *Builder) SetDisplaySource(display bool) *Builder {
	b.displaySource = display
	return b
}

func (b *Builder) SetWriter(writer io.Writer) *Builder {
	b.writer = writer
	return b
}

func (b *Builder) Build() *slog.Logger {
	var handler slog.Handler

	if b.logFormat == FormatJSON {
		handler = slog.NewJSONHandler(
			b.writer,
			&slog.HandlerOptions{
				AddSource: b.displaySource,
				Level:     b.level.ToSlogLevel(),
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == "time" {
						return slog.Attr{
							Key:   "time",
							Value: slog.StringValue(a.Value.Time().Format(b.dateFormat)),
						}
					}
					return a
				},
			},
		)
	} else {
		handler = clog.New(
			clog.WithLevel(b.level.ToSlogLevel()),
			clog.WithSource(b.displaySource),
			clog.WithWriter(b.writer),
			clog.WithTimeFmt(b.dateFormat),
			clog.WithPrinter(clog.IndentPrinter),
		)
	}
	return slog.New(handler)
}

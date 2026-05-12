package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level   string `yaml:"level"`
	Component string `yaml:"component"`
}

func (lc *Config) GetLevel() slog.Level {
	switch strings.ToLower(lc.Level) {
	case "debug":
		return slog.LevelDebug 
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func Init(cfg *Config) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{
		Level: cfg.GetLevel(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02|15:04:05.00")) // <-- your format
			}
			return a
		},
	}
	handler := NewContextHandler(
		slog.NewTextHandler(os.Stdout, handlerOpts),
	)
	l := slog.New(handler).With("component", cfg.Component)
	slog.SetDefault(l)
	return l
}



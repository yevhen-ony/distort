package logger

import (
	"log/slog"
	"os"
	"strings"
)

type LogConfig struct {
	Level   string `yaml:"level"`
	Service string `yaml:"service"`
}

func (lc *LogConfig) GetLevel() slog.Level {
	switch strings.ToLower(lc.Level) {
	case "debug":
		return slog.LevelDebug 
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func Init(cfg *LogConfig) *slog.Logger {
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
	handler := slog.NewTextHandler(os.Stdout, handlerOpts)
	l := slog.New(handler).With("service", cfg.Service)
	slog.SetDefault(l)
	return l
}

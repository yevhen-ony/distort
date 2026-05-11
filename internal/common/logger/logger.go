package logger

import (
	"context"
	"dos/internal/common/dosctx"
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

type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(next slog.Handler) slog.Handler {
	return &ContextHandler{
		Handler: next,
	}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if objectID, ok := dosctx.ObjectID(ctx); ok {
		r.AddAttrs(slog.String("object_id", string(objectID)))
	}
	if chunkID, ok := dosctx.ChunkID(ctx); ok {
		r.AddAttrs(slog.String("chunk_id", string(chunkID)))
	}
	if service, ok := dosctx.Service(ctx); ok {
		r.AddAttrs(slog.String("service", service))
	}
	if operation, ok := dosctx.Operation(ctx); ok {
		r.AddAttrs(slog.String("operation", operation))
	}
	return h.Handler.Handle(ctx, r)
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



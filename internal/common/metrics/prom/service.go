package prom

import (
	"context"
	"dos/internal/common/metrics"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	ScrapePort int    `yaml:"scrape_port"`
	ScrapePath string `yaml:"scrape_path"`
	Namespace  string `yaml:"namespace"`
}

func (c *Config) ListeningAddr() string {
	return fmt.Sprintf(":%d", c.ScrapePort)
}

type Service struct {
	registry *prometheus.Registry
	provider *Provider
	config   Config
}

func NewService(config Config) *Service {
	registry := prometheus.NewRegistry()
	provider := newProvider(registry, config.Namespace)

	s := &Service{
		registry: registry,
		provider: provider,
		config:   config,
	}
	return s
}

func (s *Service) Provider() metrics.Provider {
	return s.provider
}

func (s *Service) Serve(ctx context.Context) (err error) {
	opts := promhttp.HandlerOpts{Registry: s.registry}
	handler := promhttp.HandlerFor(s.registry, opts)

	mux := http.NewServeMux()
	mux.Handle(s.config.ScrapePath, handler)

	srv := http.Server{
		Addr:    s.config.ListeningAddr(),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(stopCtx)
		<-errCh
		return nil
	}
}

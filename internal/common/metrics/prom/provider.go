package prom

import (
	"dos/internal/common/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	registry  *prometheus.Registry
	namespace string
}

func newProvider(registry *prometheus.Registry, namespace string) *Provider {
	return &Provider{
		registry:  registry,
		namespace: namespace,
	}
}

func (p *Provider) Counter(opts metrics.CounterOpts) metrics.Counter {
	count := prometheus.NewCounter(prometheus.CounterOpts{
		Name: opts.Name,
		Help: opts.Help,
	})
	p.registry.MustRegister(count)
	return count
}

func (p *Provider) Gauge(opts metrics.GaugeOpts) metrics.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: opts.Name,
		Help: opts.Help,
	})
	p.registry.MustRegister(gauge)
	return gauge
}

func (p *Provider) Histogram(opts metrics.HistogramOpts) metrics.Histogram {
	hist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    opts.Name,
		Help:    opts.Help,
		Buckets: opts.Buckets,
	})
	p.registry.MustRegister(hist)
	return hist
}

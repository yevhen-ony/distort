package metrics

type Counter interface {
	Inc()
	Add(float64)
}

type Gauge interface {
	Set(float64)
	Add(float64)
}

type Histogram interface {
	Observe(float64)
}

type Provider interface {
	Counter(CounterOpts) Counter
	Gauge(GaugeOpts) Gauge
	Histogram(HistogramOpts) Histogram
}

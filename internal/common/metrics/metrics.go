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
  	Counter(name string, labels ...string) Counter
  	Gauge(name string, labels ...string) Gauge
  	Histogram(name string, labels ...string) Histogram
}

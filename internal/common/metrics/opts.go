package metrics

type CounterOpts struct {
	Name string
	Help string
}

type GaugeOpts struct {
	Name string
	Help string
}

type HistogramOpts struct {
	Name    string
	Help    string
	Buckets []float64
}

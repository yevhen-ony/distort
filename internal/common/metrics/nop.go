package metrics

type nopCounter struct{}

func (nopCounter) Inc()        {}
func (nopCounter) Add(float64) {}

type nopGauge struct{}

func (nopGauge) Set(float64) {}
func (nopGauge) Add(float64) {}

type nopHistogram struct{}

func (nopHistogram) Observe(float64) {}

type NopProvider struct{}

func (NopProvider) Counter(CounterOpts) Counter       { return nopCounter{} }
func (NopProvider) Gauge(GaugeOpts) Gauge             { return nopGauge{} }
func (NopProvider) Histogram(HistogramOpts) Histogram { return nopHistogram{} }

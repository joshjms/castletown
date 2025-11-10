package telemetry

type Metrics struct{}

func NewMetricsRegistry() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Inc(name string)                    {}
func (m *Metrics) Observe(name string, value float64) {}

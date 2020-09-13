package attacker

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"gopkg.in/yaml.v2"
)

// Metrics wraps "vegeta.Metrics" to avoid dependency on it.
type Metrics struct {
	Latencies LatencyMetrics `json:"latencies"`
}

type LatencyMetrics struct {
	Total time.Duration `json:"total"`
	Mean  time.Duration `json:"mean"`
	P50   time.Duration `json:"50th"`
	P90   time.Duration `json:"90th"`
	P95   time.Duration `json:"95th"`
	P99   time.Duration `json:"99th"`
	Max   time.Duration `json:"max"`
	Min   time.Duration `json:"min"`
}

func newMetrics(m *vegeta.Metrics) *Metrics {
	return &Metrics{
		Latencies: LatencyMetrics{
			Total: m.Latencies.Total,
			Mean:  m.Latencies.Mean,
			P50:   m.Latencies.P50,
			P90:   m.Latencies.P90,
			P95:   m.Latencies.P95,
			P99:   m.Latencies.P99,
			Max:   m.Latencies.Max,
			Min:   m.Latencies.Min,
		},
	}
}

func (m *Metrics) String() string {
	b, _ := yaml.Marshal(m)
	return string(b)
}

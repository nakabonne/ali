package attacker

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"gopkg.in/yaml.v2"
)

// Metrics wraps "vegeta.Metrics" to avoid dependency on it.
// TODO: Add more fields
type Metrics struct {
	// Latencies holds computed request latency metrics.
	Latencies vegeta.LatencyMetrics `json:"latencies"`
	// Histogram, only if requested
	//Histogram *vegeta.Histogram `json:"buckets,omitempty"`
	// BytesIn holds computed incoming byte metrics.
	BytesIn vegeta.ByteMetrics `json:"bytes_in"`
	// BytesOut holds computed outgoing byte metrics.
	BytesOut vegeta.ByteMetrics `json:"bytes_out"`
	// Earliest is the earliest timestamp in a Result set.
	Earliest time.Time `json:"earliest"`
	// Latest is the latest timestamp in a Result set.
	Latest time.Time `json:"latest"`
	// End is the latest timestamp in a Result set plus its latency.
	End time.Time `json:"end"`
	// Duration is the duration of the attack.
	Duration time.Duration `json:"duration"`
	// Wait is the extra time waiting for responses from targets.
	Wait time.Duration `json:"wait"`
	// Requests is the total number of requests executed.
	Requests uint64 `json:"requests"`
	// Rate is the rate of sent requests per second.
	Rate float64 `json:"rate"`
	// Throughput is the rate of successful requests per second.
	Throughput float64 `json:"throughput"`
	// Success is the percentage of non-error responses.
	Success float64 `json:"success"`
	// StatusCodes is a histogram of the responses' status codes.
	StatusCodes map[string]int `json:"status_codes"`
	// Errors is a set of unique errors returned by the targets during the attack.
	Errors []string `json:"errors"`
}

func newMetrics(m *vegeta.Metrics) *Metrics {
	return &Metrics{
		Latencies: m.Latencies,
		//Histogram:   m.Histogram,
		BytesIn:     m.BytesIn,
		BytesOut:    m.BytesOut,
		Earliest:    m.Earliest,
		Latest:      m.Latest,
		End:         m.End,
		Duration:    m.Duration,
		Wait:        m.Wait,
		Requests:    m.Requests,
		Rate:        m.Rate,
		Throughput:  m.Throughput,
		Success:     m.Success,
		StatusCodes: m.StatusCodes,
		Errors:      m.Errors,
	}
}

func (m *Metrics) String() string {
	b, _ := yaml.Marshal(m)
	return string(b)
}

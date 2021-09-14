package attacker

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Metrics wraps "vegeta.Metrics" to avoid dependency on it.
type Metrics struct {
	// Latencies holds computed request latency metrics.
	Latencies LatencyMetrics `json:"latencies"`
	// Histogram, only if requested
	// Histogram *vegeta.Histogram `json:"buckets,omitempty"`
	// BytesIn holds computed incoming byte metrics.
	BytesIn ByteMetrics `json:"bytes_in"`
	// BytesOut holds computed outgoing byte metrics.
	BytesOut ByteMetrics `json:"bytes_out"`
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

// LatencyMetrics holds computed request latency metrics.
type LatencyMetrics struct {
	// Total is the total latency sum of all requests in an attack.
	Total time.Duration `json:"total"`
	// Mean is the mean request latency.
	Mean time.Duration `json:"mean"`
	// P50 is the 50th percentile request latency.
	P50 time.Duration `json:"50th"`
	// P90 is the 90th percentile request latency.
	P90 time.Duration `json:"90th"`
	// P95 is the 95th percentile request latency.
	P95 time.Duration `json:"95th"`
	// P99 is the 99th percentile request latency.
	P99 time.Duration `json:"99th"`
	// Max is the maximum observed request latency.
	Max time.Duration `json:"max"`
	// Min is the minimum observed request latency.
	Min time.Duration `json:"min"`
}

// ByteMetrics holds computed byte flow metrics.
type ByteMetrics struct {
	// Total is the total number of flowing bytes in an attack.
	Total uint64 `json:"total"`
	// Mean is the mean number of flowing bytes per hit.
	Mean float64 `json:"mean"`
}

func newMetrics(m *vegeta.Metrics) *Metrics {
	statusCodes := make(map[string]int, len(m.StatusCodes))
	for k, v := range m.StatusCodes {
		statusCodes[k] = v
	}

	return &Metrics{
		Latencies: LatencyMetrics{
			Total: m.Latencies.Total,
			Mean:  m.Latencies.Mean,
			P50:   m.Latencies.Quantile(0.50),
			P90:   m.Latencies.Quantile(0.90),
			P95:   m.Latencies.Quantile(0.95),
			P99:   m.Latencies.Quantile(0.99),
			Max:   m.Latencies.Max,
			Min:   m.Latencies.Min,
		},
		//Histogram:   m.Histogram,
		BytesIn: ByteMetrics{
			Total: m.BytesIn.Total,
			Mean:  m.BytesIn.Mean,
		},
		BytesOut: ByteMetrics{
			Total: m.BytesOut.Total,
			Mean:  m.BytesOut.Mean,
		},
		Earliest:    m.Earliest,
		Latest:      m.Latest,
		End:         m.End,
		Duration:    m.Duration,
		Wait:        m.Wait,
		Requests:    m.Requests,
		Rate:        m.Rate,
		Throughput:  m.Throughput,
		Success:     m.Success,
		StatusCodes: statusCodes,
		Errors:      m.Errors,
	}
}

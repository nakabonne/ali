package attacker

import (
	"time"

	"github.com/nakabonne/ali/export"
)

func newSummary(targetURL, method string, rate int, duration time.Duration, metrics *Metrics) export.Summary {
	return export.Summary{
		Target: export.TargetSummary{
			URL:    targetURL,
			Method: method,
		},
		Parameters: export.ParametersSummary{
			Rate:            rate,
			DurationSeconds: duration.Seconds(),
		},
		Timing: export.TimingSummary{
			Earliest: metrics.Earliest,
			Latest:   metrics.Latest,
		},
		Requests: export.RequestsSummary{
			Count:        metrics.Requests,
			SuccessRatio: metrics.Success,
		},
		Throughput: metrics.Throughput,
		LatencyMS: export.LatencySummary{
			Total: durationToMillis(metrics.Latencies.Total),
			Mean:  durationToMillis(metrics.Latencies.Mean),
			P50:   durationToMillis(metrics.Latencies.P50),
			P90:   durationToMillis(metrics.Latencies.P90),
			P95:   durationToMillis(metrics.Latencies.P95),
			P99:   durationToMillis(metrics.Latencies.P99),
			Max:   durationToMillis(metrics.Latencies.Max),
			Min:   durationToMillis(metrics.Latencies.Min),
		},
		Bytes: export.BytesSummary{
			In: export.BytesFlowSummary{
				Total: metrics.BytesIn.Total,
				Mean:  metrics.BytesIn.Mean,
			},
			Out: export.BytesFlowSummary{
				Total: metrics.BytesOut.Total,
				Mean:  metrics.BytesOut.Mean,
			},
		},
		StatusCodes: export.StatusCodesSummary(metrics.StatusCodes),
	}
}

func durationToMillis(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

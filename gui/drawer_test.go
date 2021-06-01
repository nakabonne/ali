package gui

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.uber.org/atomic"

	"github.com/nakabonne/ali/attacker"
	"github.com/nakabonne/ali/storage"
)

func TestRedrawCharts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name             string
		storage          storage.Reader
		latencyChart     LineChart
		percentilesChart LineChart
	}{
		{
			name:    "two data points for each metric",
			storage: &storage.FakeStorage{Values: []float64{1, 2}},
			latencyChart: func() LineChart {
				l := NewMockLineChart(ctrl)
				l.EXPECT().Series("latency", []float64{1, 2}, gomock.Any()).AnyTimes()
				return l
			}(),
			percentilesChart: func() LineChart {
				l := NewMockLineChart(ctrl)
				l.EXPECT().Series("p50", []float64{1, 2}, gomock.Any()).AnyTimes()
				l.EXPECT().Series("p90", []float64{1, 2}, gomock.Any()).AnyTimes()
				l.EXPECT().Series("p95", []float64{1, 2}, gomock.Any()).AnyTimes()
				l.EXPECT().Series("p99", []float64{1, 2}, gomock.Any()).AnyTimes()
				return l
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			d := &drawer{
				redrawInterval: DefaultRedrawInterval,
				widgets:        &widgets{latencyChart: tt.latencyChart, percentilesChart: tt.percentilesChart},
				chartDrawing:   atomic.NewBool(false),
				storage:        tt.storage,
			}
			go d.redrawCharts(ctx)
		})
	}
}

func TestRedrawGauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name  string
		size  time.Duration
		count int
		gauge Gauge
	}{
		{
			name: "draw once",
			size: 1,
			gauge: func() Gauge {
				g := NewMockGauge(ctrl)
				g.EXPECT().Percent(0, gomock.Any()).AnyTimes()
				g.EXPECT().Percent(100, gomock.Any()).AnyTimes()
				return g
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &drawer{
				redrawInterval: DefaultRedrawInterval,
				widgets:        &widgets{progressGauge: tt.gauge},
			}
			go d.redrawGauge(ctx, tt.size)
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		metrics         *attacker.Metrics
		latenciesText   Text
		bytesText       Text
		othersText      Text
		statusCodesText Text
		errorsText      Text
	}{
		{
			name: "with errors",
			metrics: &attacker.Metrics{
				Latencies: attacker.LatencyMetrics{
					Total: 1,
					Mean:  1,
					P50:   1,
					P90:   1,
					P95:   1,
					P99:   1,
					Max:   1,
					Min:   1,
				},
				BytesIn: attacker.ByteMetrics{
					Total: 1,
					Mean:  1,
				},
				BytesOut: attacker.ByteMetrics{
					Total: 1,
					Mean:  1,
				},
				Earliest:    time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Latest:      time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				End:         time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:    1,
				Wait:        1,
				Requests:    1,
				Rate:        1,
				Throughput:  1,
				Success:     1,
				StatusCodes: map[string]int{"200": 2},
				Errors:      []string{"error1"},
			},
			latenciesText: func() Text {
				t := NewMockText(ctrl)
				t.EXPECT().Write(`Total: 1ns
Mean: 1ns
P50: 1ns
P90: 1ns
P95: 1ns
P99: 1ns
Max: 1ns
Min: 1ns`, gomock.Any()).AnyTimes()
				return t
			}(),

			bytesText: func() Text {
				t := NewMockText(ctrl)
				t.EXPECT().Write(`In:
  Total: 1
  Mean: 1
Out:
  Total: 1
  Mean: 1`, gomock.Any()).AnyTimes()
				return t
			}(),

			statusCodesText: func() Text {
				t := NewMockText(ctrl)
				t.EXPECT().Write(`"200": 2
`, gomock.Any()).AnyTimes()
				return t
			}(),

			errorsText: func() Text {
				t := NewMockText(ctrl)
				t.EXPECT().Write(`- error1
`, gomock.Any()).AnyTimes()
				return t
			}(),

			othersText: func() Text {
				t := NewMockText(ctrl)
				t.EXPECT().Write(`Duration: 1ns
Wait: 1ns
Requests: 1
Rate: 1.000000
Throughput: 1.000000
Success: 1.000000
Earliest: 2009-11-10T23:00:00Z
Latest: 2009-11-10T23:00:00Z
End: 2009-11-10T23:00:00Z`, gomock.Any()).AnyTimes()

				return t
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			d := &drawer{
				redrawInterval: DefaultRedrawInterval,
				widgets: &widgets{
					latenciesText:   tt.latenciesText,
					bytesText:       tt.bytesText,
					othersText:      tt.othersText,
					statusCodesText: tt.statusCodesText,
					errorsText:      tt.errorsText,
				},
				metrics: tt.metrics,
			}
			go d.redrawMetrics(ctx)
			time.Sleep(1 * time.Second)
			cancel()
		})
	}
}

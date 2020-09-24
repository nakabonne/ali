package gui

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/nakabonne/ali/attacker"
)

func TestRedrawChart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name         string
		results      []*attacker.Result
		latencyChart LineChart
	}{
		{
			name: "one result received",
			results: []*attacker.Result{
				{
					Latency: 1000000,
				},
			},
			latencyChart: func() LineChart {
				l := NewMockLineChart(ctrl)
				l.EXPECT().Series("latency", []float64{1.0}, gomock.Any())
				return l
			}(),
		},
		{
			name: "two results received",
			results: []*attacker.Result{
				{
					Latency: 1000000,
				},
				{
					Latency: 2000000,
				},
			},
			latencyChart: func() LineChart {
				l := NewMockLineChart(ctrl)
				l.EXPECT().Series("latency", []float64{1.0}, gomock.Any())
				l.EXPECT().Series("latency", []float64{1.0, 2.0}, gomock.Any())
				return l
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &drawer{
				widgets: &widgets{latencyChart: tt.latencyChart},
				chartCh: make(chan *attacker.Result),
				gaugeCh: make(chan bool),
			}
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case end := <-d.gaugeCh:
						if end {
							return
						}
					}
				}
			}()
			go d.redrawChart(ctx, len(tt.results))
			for _, res := range tt.results {
				d.chartCh <- res
			}
			d.chartCh <- &attacker.Result{End: true}
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
		size  int
		count int
		gauge Gauge
	}{
		{
			name: "draw once",
			size: 1,
			gauge: func() Gauge {
				g := NewMockGauge(ctrl)
				g.EXPECT().Percent(0)
				g.EXPECT().Percent(100)
				return g
			}(),
		},
		{
			name: "draw twice",
			size: 2,
			gauge: func() Gauge {
				g := NewMockGauge(ctrl)
				g.EXPECT().Percent(0)
				g.EXPECT().Percent(50)
				g.EXPECT().Percent(100)
				return g
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &drawer{
				widgets: &widgets{progressGauge: tt.gauge},
				gaugeCh: make(chan bool),
			}
			go d.redrawGauge(ctx, tt.size)
			for i := 0; i < tt.size; i++ {
				d.gaugeCh <- false
			}
			d.gaugeCh <- true
		})
	}
}

func TestRedrawMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name    string
		message string
		text    Text
	}{
		{
			name:    "write foo",
			message: "foo",
			text: func() Text {
				g := NewMockText(ctrl)
				g.EXPECT().Write("foo", gomock.Any())
				return g
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &drawer{
				widgets:   &widgets{messageText: tt.text},
				messageCh: make(chan string),
			}
			go d.redrawMessage(ctx)
			d.messageCh <- tt.message
		})
	}
}

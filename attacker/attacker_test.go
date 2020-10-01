package attacker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAttack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name         string
		target       string
		opts         Options
		want         *Metrics
		wantResCount int
	}{
		{
			name:   "no target given",
			target: "",
			want:   nil,
		},
		{
			name:   "no result given back",
			target: "http://host.xz",
			opts: Options{
				Attacker: &fakeAttacker{},
			},
			want: &Metrics{
				StatusCodes: make(map[string]int),
				Errors:      []string{},
			},
			wantResCount: 0,
		},
		{
			name:   "two result given back",
			target: "http://host.xz",
			opts: Options{
				Attacker: &fakeAttacker{
					results: []*vegeta.Result{
						{
							Code: 200,
						},
						{
							Code: 200,
						},
					},
				},
			},
			want: &Metrics{
				Requests:   2,
				Rate:       2,
				Throughput: 2,
				Success:    1,
				StatusCodes: map[string]int{
					"200": 2,
				},
				Errors: []string{},
			},
			wantResCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resCh := make(chan *Result, 100)
			metricsCh := make(chan *Metrics, 100)
			Attack(ctx, tt.target, resCh, metricsCh, tt.opts)
			assert.Equal(t, tt.wantResCount, len(resCh))
		})
	}
}

package attacker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	"go.uber.org/goleak"
)

type watcher struct {
	resCh chan *Result
	count int
}

func (w *watcher) countRes(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-w.resCh:
			w.count++
		}
	}
}

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
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		w := &watcher{
			resCh: make(chan *Result),
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go w.countRes(ctx, &wg)
		t.Run(tt.name, func(t *testing.T) {
			got := Attack(ctx, tt.target, w.resCh, tt.opts)
			cancel()
			wg.Wait()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantResCount, w.count)
		})
	}
}

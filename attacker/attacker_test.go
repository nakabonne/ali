package attacker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttack(t *testing.T) {
	tests := []struct {
		name   string
		target string
		resCh  chan *Result
		opts   Options
		want   *Metrics
	}{
		// TODO: Add test cases.
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt.opts.Attacker = &fakeAttacker{}
		t.Run(tt.name, func(t *testing.T) {
			got := Attack(ctx, tt.target, tt.resCh, tt.opts)
			assert.Equal(t, got, tt.want)
		})
	}
}

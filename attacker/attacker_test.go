package attacker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAttack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tests := []struct {
		name   string
		target string
		opts   Options
		want   *Metrics
	}{
		{
			name:   "no target given",
			target: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		tt.opts.Attacker = &fakeAttacker{}
		t.Run(tt.name, func(t *testing.T) {
			got := Attack(ctx, tt.target, make(chan *Result), tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

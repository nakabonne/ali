package gui

import (
	"context"
	"fmt"
	"testing"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/stretchr/testify/assert"
)

// TODO: Ensure to kill all goroutine when running unit tests.
/*func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}*/

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		r       runner
		wantErr bool
	}{
		{
			name: "successful running",
			r: func(context.Context, terminalapi.Terminal, *container.Container, ...termdash.Option) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "failed running",
			r: func(context.Context, terminalapi.Terminal, *container.Container, ...termdash.Option) error {
				return fmt.Errorf("error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.r)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

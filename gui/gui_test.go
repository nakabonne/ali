package gui

import (
	"context"
	"fmt"
	"testing"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/nakabonne/ali/attacker"
	"github.com/nakabonne/ali/storage"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		runner  runner
		wantErr bool
	}{
		{
			name: "successful running",
			runner: func(context.Context, terminalapi.Terminal, *container.Container, ...termdash.Option) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "failed running",
			runner: func(context.Context, terminalapi.Terminal, *container.Container, ...termdash.Option) error {
				return fmt.Errorf("error")
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(&termbox.Terminal{}, tt.runner, "", &storage.FakeStorage{}, &attacker.FakeAttacker{}, Options{})
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

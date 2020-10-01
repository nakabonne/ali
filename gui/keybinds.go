package gui

import (
	"context"
	"time"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

func keybinds(ctx context.Context, cancel context.CancelFunc, dr *drawer, targetURL string, opts attacker.Options) func(*terminalapi.Keyboard) {
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyCtrlC: // Quit
			cancel()
		case keyboard.KeyEnter: // Attack
			attack(ctx, dr, targetURL, opts)
		}
	}
}

func attack(ctx context.Context, d *drawer, target string, opts attacker.Options) {
	if d.chartDrawing {
		d.messageCh <- "Wait until the attack is over"
		return
	}
	requestNum := opts.Rate * int(opts.Duration/time.Second)

	// To pre-allocate, run redrawChart on a per-attack basis.
	go d.redrawChart(ctx, requestNum)
	go d.redrawGauge(ctx, requestNum)
	go func(ctx context.Context, d *drawer, t string, o attacker.Options) {
		attacker.Attack(ctx, t, d.chartCh, d.metricsCh, o) // this blocks until attack finishes
		d.chartCh <- &attacker.Result{End: true}
		d.messageCh <- "Attack completed"
	}(ctx, d, target, opts)
}

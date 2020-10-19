package gui

import (
	"context"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

func navigateCharts(chartFuncs []func()) func(bool) {
	position := 0
	numFuncs := len(chartFuncs)
	return func(backwards bool) {
		position++
		if backwards {
			position -= 2
			if position < 0 {
				position = numFuncs - 1
			}
		}
		chartFuncs[position%numFuncs]()
	}
}

func keybinds(ctx context.Context, cancel context.CancelFunc, c *container.Container, dr *drawer, targetURL string, opts attacker.Options) func(*terminalapi.Keyboard) {
	funcs := []func(){
		func() { c.Update(chartID, dr.gridOpts.latency...) },
		func() { c.Update(chartID, dr.gridOpts.percentiles...) },
	}
	navigateFunc := navigateCharts(funcs)
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyCtrlC: // Quit
			cancel()
		case keyboard.KeyEnter: // Attack
			attack(ctx, dr, targetURL, opts)
		case 'H', 'h': // backwards
			navigateFunc(true)
		case 'L', 'l': // forwards
			navigateFunc(false)
		}
	}
}

func attack(ctx context.Context, d *drawer, target string, opts attacker.Options) {
	if d.chartDrawing.Load() {
		return
	}
	requestNum := opts.Rate * int(opts.Duration/time.Second)
	d.doneCh = make(chan struct{})

	// To initialize, run redrawChart on a per-attack basis.
	go d.appendChartValues(ctx, requestNum)
	go d.redrawCharts(ctx)
	go func(ctx context.Context, d *drawer, t string, o attacker.Options) {
		attacker.Attack(ctx, t, d.chartCh, d.metricsCh, o) // this blocks until attack finishes
		close(d.doneCh)
	}(ctx, d, target, opts)
}

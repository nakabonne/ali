package gui

import (
	"context"

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

func keybinds(ctx context.Context, cancel context.CancelFunc, c *container.Container, dr *drawer, a attacker.Attacker) func(*terminalapi.Keyboard) {
	funcs := []func(){
		func() { c.Update(chartID, dr.gridOpts.latency...) },
		func() { c.Update(chartID, dr.gridOpts.percentiles...) },
	}
	navigateFunc := navigateCharts(funcs)
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyCtrlC, 'q': // Quit
			cancel()
		case keyboard.KeyEnter: // Attack
			attack(ctx, cancel, dr, a)
		case 'H', 'h': // backwards
			navigateFunc(true)
		case 'L', 'l': // forwards
			navigateFunc(false)
		}
	}
}

func attack(ctx context.Context, cancelParent context.CancelFunc, d *drawer, a attacker.Attacker) {
	if d.chartDrawing.Load() {
		return
	}
	child, cancelChild := context.WithCancel(ctx)

	// To initialize, run redrawChart on a per-attack basis.
	go d.redrawCharts(child)
	go d.redrawGauge(child, a.Duration())
	go func() {
		if err := a.Attack(child, d.metricsCh); err != nil {
			d.setExportErr(err)
			cancelChild()
			cancelParent()
			return
		}
		cancelChild()
	}()
}

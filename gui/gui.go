package gui

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

const rootID = "root"

func Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		return fmt.Errorf("failed to generate terminal interface: %w", err)
	}
	defer t.Close()

	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		return fmt.Errorf("failed to generate container: %w", err)
	}

	w, err := newWidgets()
	if err != nil {
		return fmt.Errorf("failed to generate widgets: %w", err)
	}
	gridOpts, err := gridLayout(w)
	if err != nil {
		return fmt.Errorf("failed to build grid layout: %w", err)
	}
	if err := c.Update(rootID, gridOpts...); err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	d := &drawer{
		widgets:  w,
		chartsCh: make(chan *attacker.Result),
		reportCh: make(chan string),
	}
	go d.redrawReport(ctx)

	k := keybinds(ctx, cancel, d)

	return termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(k), termdash.RedrawInterval(redrawInterval))
}

func keybinds(ctx context.Context, cancel context.CancelFunc, dr *drawer) func(*terminalapi.Keyboard) {
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyEsc, keyboard.KeyCtrlC:
			cancel()
		case keyboard.KeySpace:
			if dr.chartDrawing {
				return
			}
			// TODO: Enalble to poplulate from input
			var (
				rate        = 10
				duration, _ = time.ParseDuration("2s")
				requestNum  = rate * int(duration/time.Second)
				target      = "http://34.84.111.163:9898"
			)
			// To pre-allocate, run redrawChart on a per-attack basis.
			go dr.redrawChart(ctx, requestNum)
			go func(ctx context.Context, d *drawer, t string, r int, du time.Duration) {
				metrics := attacker.Attack(ctx, t, d.chartsCh, attacker.Options{Rate: r, Duration: du})
				d.reportCh <- metrics.String()
				d.chartsCh <- &attacker.Result{End: true}
			}(ctx, dr, target, rate, duration)
		}
	}
}

func gridLayout(w *widgets) ([]container.Option, error) {
	raw1 := grid.RowHeightPerc(50,
		grid.ColWidthPerc(99, grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency"))),
	)
	raw2 := grid.RowHeightPerc(45,
		grid.ColWidthPerc(50, grid.Widget(w.urlInput, container.Border(linestyle.Light), container.BorderTitle("Input"))),
		grid.ColWidthPerc(49, grid.Widget(w.reportText, container.Border(linestyle.Light), container.BorderTitle("Report"))),
	)
	raw3 := grid.RowHeightPerc(4,
		grid.ColWidthPerc(99, grid.Widget(w.navi, container.Border(linestyle.Light))),
	)

	builder := grid.New()
	builder.Add(
		raw1,
		raw2,
		raw3,
	)

	return builder.Build()
}

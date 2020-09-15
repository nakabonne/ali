package gui

import (
	"context"
	"fmt"
	"net/url"
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
		chartCh:  make(chan *attacker.Result),
		gaugeCh:  make(chan bool),
		reportCh: make(chan string),
	}
	go d.redrawReport(ctx)

	k := keybinds(ctx, cancel, d)

	return termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(k), termdash.RedrawInterval(redrawInterval))
}

func gridLayout(w *widgets) ([]container.Option, error) {
	raw1 := grid.RowHeightPerc(60,
		grid.ColWidthPerc(99, grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency (ms)"))),
	)
	raw2 := grid.RowHeightPerc(35,
		grid.ColWidthPerc(64,
			grid.RowHeightPerc(33, grid.Widget(w.urlInput, container.Border(linestyle.None))),
			grid.RowHeightPerc(33,
				grid.ColWidthPerc(33, grid.Widget(w.rateLimitInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(33, grid.Widget(w.durationInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(33, grid.Widget(w.timeoutInput, container.Border(linestyle.None))),
			),
			grid.RowHeightPerc(33,
				grid.ColWidthPerc(33, grid.Widget(w.methodInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(33, grid.Widget(w.headerInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(33, grid.Widget(w.bodyInput, container.Border(linestyle.None))),
			),
		),
		grid.ColWidthPerc(35, grid.Widget(w.reportText, container.Border(linestyle.Light), container.BorderTitle("Report"))),
	)
	raw3 := grid.RowHeightPerc(4,
		grid.ColWidthPerc(64, grid.Widget(w.progressGauge, container.Border(linestyle.None))),
		grid.ColWidthPerc(33, grid.Widget(w.navi, container.Border(linestyle.Light))),
	)

	builder := grid.New()
	builder.Add(
		raw1,
		raw2,
		raw3,
	)

	return builder.Build()
}

func keybinds(ctx context.Context, cancel context.CancelFunc, dr *drawer) func(*terminalapi.Keyboard) {
	return func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyCtrlC: // Quit
			cancel()
		case keyboard.KeyEnter: // Attack
			attack(ctx, dr)
		}
	}
}

func attack(ctx context.Context, d *drawer) {
	if d.chartDrawing {
		return
	}
	target := d.widgets.urlInput.Read()
	if _, err := url.ParseRequestURI(target); err != nil {
		d.reportCh <- fmt.Sprintf("Bad URL: %v", err)
		return
	}
	opts, err := makeOptions(d.widgets)
	if err != nil {
		d.reportCh <- err.Error()
		return
	}
	requestNum := opts.Rate * int(opts.Duration/time.Second)

	// To pre-allocate, run redrawChart on a per-attack basis.
	go d.redrawChart(ctx, requestNum)
	go d.redrawGauge(ctx, requestNum)
	go func(ctx context.Context, d *drawer, t string, o attacker.Options) {
		metrics := attacker.Attack(ctx, t, d.chartCh, o)
		d.reportCh <- metrics.String()
		d.chartCh <- &attacker.Result{End: true}
	}(ctx, d, target, opts)
}

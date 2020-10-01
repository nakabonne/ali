package gui

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

const (
	// How often termdash redraws the screen.
	redrawInterval = 250 * time.Millisecond
	rootID         = "root"
)

type runner func(ctx context.Context, t terminalapi.Terminal, c *container.Container, opts ...termdash.Option) error

func Run(targetURL string, opts *attacker.Options) error {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		return fmt.Errorf("failed to generate terminal interface: %w", err)
	}
	defer t.Close()
	return run(t, termdash.Run, targetURL, opts)
}

func run(t *termbox.Terminal, r runner, targetURL string, opts *attacker.Options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		widgets:   w,
		chartCh:   make(chan *attacker.Result),
		gaugeCh:   make(chan bool),
		metricsCh: make(chan *attacker.Metrics),
		messageCh: make(chan string),
	}
	go d.redrawMetrics(ctx)
	go d.redrawMessage(ctx)

	k := keybinds(ctx, cancel, d, targetURL, *opts)

	return r(ctx, t, c, termdash.KeyboardSubscriber(k), termdash.RedrawInterval(redrawInterval))
}

func gridLayout(w *widgets) ([]container.Option, error) {
	raw1 := grid.RowHeightPerc(65, grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency (ms)")))
	raw2 := grid.RowHeightPerc(30,
		grid.ColWidthPerc(15, grid.Widget(w.paramsText, container.Border(linestyle.Light), container.BorderTitle("Parameters"))),
		grid.ColWidthPerc(15, grid.Widget(w.latenciesText, container.Border(linestyle.Light), container.BorderTitle("Latencies"))),
		grid.ColWidthPerc(15, grid.Widget(w.bytesText, container.Border(linestyle.Light), container.BorderTitle("Bytes"))),
		grid.ColWidthPerc(50, grid.Widget(w.othersText, container.Border(linestyle.Light), container.BorderTitle("Others"))),
	)
	raw3 := grid.RowHeightPerc(4,
		grid.ColWidthPerc(50, grid.Widget(w.progressGauge, container.Border(linestyle.Light), container.BorderTitle("Progress"))),
		grid.ColWidthPerc(50, grid.Widget(w.navi, container.Border(linestyle.Light))),
	)

	builder := grid.New()
	builder.Add(
		raw1,
		raw2,
		raw3,
	)

	return builder.Build()
}

package gui

import (
	"context"
	"fmt"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
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
		widgets:   w,
		chartCh:   make(chan *attacker.Result),
		gaugeCh:   make(chan bool),
		metricsCh: make(chan *attacker.Metrics),
	}
	go d.redrawReport(ctx)

	k := keybinds(ctx, cancel, d)

	return termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(k), termdash.RedrawInterval(redrawInterval))
}

func gridLayout(w *widgets) ([]container.Option, error) {
	raw1 := grid.RowHeightPerc(65, grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency (ms)")))
	raw2 := grid.RowHeightPerc(30,
		grid.ColWidthPerc(50,
			grid.RowHeightPerc(34, grid.Widget(w.urlInput, container.Border(linestyle.Light), container.BorderTitle("Target URL"))),
			grid.RowHeightPerc(33,
				grid.ColWidthPerc(20, grid.Widget(w.rateLimitInput, container.Border(linestyle.Light), container.BorderTitle("Rate Limit"))),
				grid.ColWidthPerc(20, grid.Widget(w.durationInput, container.Border(linestyle.Light), container.BorderTitle("Duration"))),
				grid.ColWidthPerc(20, grid.Widget(w.timeoutInput, container.Border(linestyle.Light), container.BorderTitle("Timeout"))),
				grid.ColWidthPerc(20, grid.Widget(w.methodInput, container.Border(linestyle.Light), container.BorderTitle("Method"))),
				grid.ColWidthPerc(19, grid.Widget(w.headerInput, container.Border(linestyle.Light), container.BorderTitle("Header"))),
			),
			grid.RowHeightPerc(33, grid.Widget(w.bodyInput, container.Border(linestyle.Light), container.BorderTitle("Body"))),
		),
		grid.ColWidthPerc(50,
			grid.RowHeightPerc(85,
				grid.ColWidthPerc(25, grid.Widget(w.latenciesText, container.Border(linestyle.Light), container.BorderTitle("Latencies"))),
				grid.ColWidthPerc(25, grid.Widget(w.bytesText, container.Border(linestyle.Light), container.BorderTitle("Bytes"))),
				grid.ColWidthPerc(50, grid.Widget(w.othersText, container.Border(linestyle.Light), container.BorderTitle("Others"))),
			),
			grid.RowHeightPerc(15, grid.Widget(w.messageText, container.Border(linestyle.Light), container.BorderTitle("Message"))),
		),
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

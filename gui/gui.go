package gui

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"go.uber.org/atomic"

	"github.com/nakabonne/ali/attacker"
)

const (
	// How often termdash redraws the screen.
	redrawInterval = 250 * time.Millisecond
	rootID         = "root"
	chartID        = "chart"
)

type runner func(ctx context.Context, t terminalapi.Terminal, c *container.Container, opts ...termdash.Option) error

func Run(targetURL string, opts *attacker.Options) error {
	var (
		t   terminalapi.Terminal
		err error
	)
	if runtime.GOOS == "windows" {
		t, err = tcell.New()
	} else {
		t, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	}
	if err != nil {
		return fmt.Errorf("failed to generate terminal interface: %w", err)
	}
	defer t.Close()
	return run(t, termdash.Run, targetURL, opts)
}

func run(t terminalapi.Terminal, r runner, targetURL string, opts *attacker.Options) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		return fmt.Errorf("failed to generate container: %w", err)
	}

	w, err := newWidgets(targetURL, opts)
	if err != nil {
		return fmt.Errorf("failed to generate widgets: %w", err)
	}
	gridOpts, err := gridLayout(w)
	if err != nil {
		return fmt.Errorf("failed to build grid layout: %w", err)
	}
	if err := c.Update(rootID, gridOpts.base...); err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	d := &drawer{
		widgets:      w,
		gridOpts:     gridOpts,
		chartCh:      make(chan *attacker.Result, 10000),
		gaugeCh:      make(chan struct{}),
		metricsCh:    make(chan *attacker.Metrics),
		chartDrawing: atomic.NewBool(false),
		metrics:      &attacker.Metrics{},
	}
	go d.updateMetrics(ctx)
	go d.redrawMetrics(ctx)

	k := keybinds(ctx, cancel, c, d, targetURL, *opts)

	return r(ctx, t, c, termdash.KeyboardSubscriber(k), termdash.RedrawInterval(redrawInterval))
}

// newChartWithLegends creates a chart with legends at the bottom.
// TODO: use it for more charts than percentiles. Any chart that has multiple series would be able to use this func.
func newChartWithLegends(lineChart LineChart, opts []container.Option, texts ...Text) ([]container.Option, error) {
	textsInColumns := func() []grid.Element {
		els := make([]grid.Element, 0, len(texts))
		for _, text := range texts {
			els = append(els, grid.ColWidthPerc(3, grid.Widget(text)))
		}
		return els
	}

	lopts := lineChart.Options()
	el := grid.RowHeightPercWithOpts(70,
		opts,
		grid.RowHeightPerc(97, grid.ColWidthPerc(99, grid.Widget(lineChart))),
		grid.RowHeightPercWithOpts(3,
			[]container.Option{container.MarginLeftPercent(lopts.MinimumSize.X)},
			textsInColumns()...,
		),
	)

	g := grid.New()
	g.Add(el)
	return g.Build()
}

// gridOpts holds all options in our grid.
// It basically holds the container options (column, width, padding, etc) of our widgets.
type gridOpts struct {
	// base options
	base []container.Option

	// so we can replace containers
	latency     []container.Option
	percentiles []container.Option
}

func gridLayout(w *widgets) (*gridOpts, error) {
	raw1 := grid.RowHeightPercWithOpts(70,
		[]container.Option{container.ID(chartID)},
		grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency (ms)")),
	)
	raw2 := grid.RowHeightPerc(25,
		grid.ColWidthPerc(20, grid.Widget(w.paramsText, container.Border(linestyle.Light), container.BorderTitle("Parameters"))),
		grid.ColWidthPerc(20, grid.Widget(w.latenciesText, container.Border(linestyle.Light), container.BorderTitle("Latencies"))),
		grid.ColWidthPerc(20, grid.Widget(w.bytesText, container.Border(linestyle.Light), container.BorderTitle("Bytes"))),
		grid.ColWidthPerc(20,
			grid.RowHeightPerc(50, grid.Widget(w.statusCodesText, container.Border(linestyle.Light), container.BorderTitle("Status Codes"))),
			grid.RowHeightPerc(50, grid.Widget(w.errorsText, container.Border(linestyle.Light), container.BorderTitle("Errors"))),
		),
		grid.ColWidthPerc(20, grid.Widget(w.othersText, container.Border(linestyle.Light), container.BorderTitle("Others"))),
	)
	raw3 := grid.RowHeightPerc(4,
		grid.ColWidthPerc(60, grid.Widget(w.progressGauge, container.Border(linestyle.Light), container.BorderTitle("Progress"))),
		grid.ColWidthPerc(40, grid.Widget(w.navi, container.Border(linestyle.Light))),
	)

	builder := grid.New()
	builder.Add(
		raw1,
		raw2,
		raw3,
	)

	baseOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	latencyBuilder := grid.New()
	latencyBuilder.Add(raw1)
	latencyOpts, err := latencyBuilder.Build()
	if err != nil {
		return nil, err
	}

	percentilesOpts, err := newChartWithLegends(w.percentilesChart, []container.Option{
		container.Border(linestyle.Light),
		container.ID(chartID),
		container.BorderTitle("Percentiles (ms)"),
	}, w.p99Legend.text, w.p95Legend.text, w.p90Legend.text, w.p50Legend.text)
	if err != nil {
		return nil, err
	}

	return &gridOpts{
		latency:     latencyOpts,
		percentiles: percentilesOpts,
		base:        baseOpts,
	}, nil
}

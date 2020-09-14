package gui

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

func gridLayout(w *widgets) ([]container.Option, error) {
	raw1 := grid.RowHeightPerc(60,
		grid.ColWidthPerc(99, grid.Widget(w.latencyChart, container.Border(linestyle.Light), container.BorderTitle("Latency (ms)"))),
	)
	raw2 := grid.RowHeightPerc(36,
		grid.ColWidthPerc(30,
			grid.RowHeightPerc(33, grid.Widget(w.urlInput, container.Border(linestyle.None))),
			grid.RowHeightPerc(33,
				grid.ColWidthPerc(50, grid.Widget(w.rateLimitInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(49, grid.Widget(w.durationInput, container.Border(linestyle.None))),
			),
			grid.RowHeightPerc(32,
				grid.ColWidthPerc(50, grid.Widget(w.methodInput, container.Border(linestyle.None))),
				grid.ColWidthPerc(49, grid.Widget(w.bodyInput, container.Border(linestyle.None))),
			),
		),
		grid.ColWidthPerc(69, grid.Widget(w.reportText, container.Border(linestyle.Light), container.BorderTitle("Report"))),
	)
	raw3 := grid.RowHeightFixed(1,
		grid.ColWidthFixed(100, grid.Widget(w.navi, container.Border(linestyle.Light))),
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

func attack(ctx context.Context, dr *drawer) {
	if dr.chartDrawing {
		return
	}
	var (
		target   string
		rate     int
		duration time.Duration
		method   string
		err      error
	)
	target = dr.widgets.urlInput.Read()
	if _, err := url.ParseRequestURI(target); err != nil {
		dr.reportCh <- fmt.Sprintf("Bad URL: %v", err)
		return
	}
	if s := dr.widgets.rateLimitInput.Read(); s != "" {
		rate, err = strconv.Atoi(s)
		if err != nil {
			dr.reportCh <- fmt.Sprintf("Given rate limit %q isn't integer: %v", s, err)
			return
		}
	}
	if s := dr.widgets.durationInput.Read(); s != "" {
		duration, err = time.ParseDuration(s)
		if err != nil {
			dr.reportCh <- fmt.Sprintf("Unparseable duration %q: %v", s, err)
			return
		}
	}
	if method = dr.widgets.methodInput.Read(); method != "" {
		if !validateMethod(method) {
			dr.reportCh <- fmt.Sprintf("Given method %q isn't an HTTP request method", method)
			return
		}
	}
	requestNum := rate * int(duration/time.Second)
	// To pre-allocate, run redrawChart on a per-attack basis.
	go dr.redrawChart(ctx, requestNum)
	go func(ctx context.Context, d *drawer, t string, r int, du time.Duration) {
		metrics := attacker.Attack(ctx, t, d.chartsCh, attacker.Options{Rate: r, Duration: du, Method: method})
		d.reportCh <- metrics.String()
		d.chartsCh <- &attacker.Result{End: true}
	}(ctx, dr, target, rate, duration)
}

func validateMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

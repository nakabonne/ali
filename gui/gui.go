package gui

import (
	"context"
	"fmt"
	"time"

	"github.com/k0kubun/pp"
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

	w, err := newWidgets(ctx, c)
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

	keybinds := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyEsc, keyboard.KeyCtrlC:
			cancel()
		case keyboard.KeyEnter:
			resultCh := make(chan *attacker.Result)
			go func() {
				// TODO: Enalble to poplulate from input
				metrics := attacker.Attack(ctx, "http://34.84.111.163:9898", resultCh, attacker.Options{Rate: 50, Duration: 10 * time.Second})
				pp.Println(metrics)
			}()
			go redrawChart(ctx, w.plotChart, resultCh)
		}
	}

	return termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(keybinds), termdash.RedrawInterval(redrawInterval))
}

func gridLayout(w *widgets) ([]container.Option, error) {
	rows := []grid.Element{
		grid.RowHeightPerc(99,
			grid.Widget(w.plotChart,
				container.Border(linestyle.Light),
				container.BorderTitle("Plot"),
			),
		),
	}
	col := grid.ColWidthPerc(99, rows...)

	builder := grid.New()
	builder.Add(
		col,
	)

	return builder.Build()
}

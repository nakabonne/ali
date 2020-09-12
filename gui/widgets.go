package gui

import (
	"context"
	"time"

	"github.com/k0kubun/pp"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"

	"github.com/nakabonne/ali/attacker"
)

// redrawInterval is how often termdash redraws the screen.
const (
	redrawInterval = 250 * time.Millisecond
)

type widgets struct {
	urlInput  *textinput.TextInput
	plotChart *linechart.LineChart
	metrics   *text.Text
	navi      *text.Text
}

func newWidgets() (*widgets, error) {
	lc, err := newLineChart()
	if err != nil {
		return nil, err
	}
	rt, err := newRollText()
	if err != nil {
		return nil, err
	}
	wt, err := newWrapText("Ctrl-c: quit, Space: attack")
	if err != nil {
		return nil, err
	}
	ti, err := newTextInput()
	if err != nil {
		return nil, err
	}

	return &widgets{
		urlInput:  ti,
		plotChart: lc,
		metrics:   rt,
		navi:      wt,
	}, nil
}

func newLineChart() (*linechart.LineChart, error) {
	return linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
}

func newRollText() (*text.Text, error) {
	return text.New(text.RollContent())
}

func newWrapText(s string) (*text.Text, error) {
	t, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if err := t.Write(s); err != nil {
		return nil, err
	}
	return t, nil
}

func newTextInput() (*textinput.TextInput, error) {
	input, err := textinput.New(
		textinput.Label("Target URL: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(20),
		textinput.PlaceHolder("enter any text"),
		textinput.OnSubmit(func(text string) error {
			// TODO: Handle on submit action, for instance using channel.
			pp.Println("input text", text)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return input, err
}

func redrawChart(ctx context.Context, lineChart *linechart.LineChart, resultCh chan *attacker.Result) {
	values := []float64{}
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-resultCh:
			pp.Println("latency", res.Latency/time.Millisecond)
			values = append(values, float64(res.Latency/time.Millisecond))
			lineChart.Series("plot", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "num",
				}),
			)
		}
	}
}

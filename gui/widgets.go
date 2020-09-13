package gui

import (
	"time"

	"github.com/k0kubun/pp"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const (
	redrawInterval = 250 * time.Millisecond
)

type widgets struct {
	urlInput     *textinput.TextInput
	latencyChart *linechart.LineChart
	reportText   *text.Text
	navi         *text.Text
}

func newWidgets() (*widgets, error) {
	lc, err := newLineChart()
	if err != nil {
		return nil, err
	}
	rt, err := newRollText("Give the target URL and press Space, then the attack will be launched.")
	if err != nil {
		return nil, err
	}
	wt, err := newRollText("Ctrl-c: quit, Space: attack")
	if err != nil {
		return nil, err
	}
	ti, err := newTextInput()
	if err != nil {
		return nil, err
	}

	return &widgets{
		urlInput:     ti,
		latencyChart: lc,
		reportText:   rt,
		navi:         wt,
	}, nil
}

func newLineChart() (*linechart.LineChart, error) {
	return linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
}

func newRollText(s string) (*text.Text, error) {
	t, err := text.New(text.RollContent())
	if err != nil {
		return nil, err
	}
	if err := t.Write(s); err != nil {
		return nil, err
	}
	return t, nil
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
		textinput.PlaceHolder("enter target URL"),
		textinput.OnSubmit(func(text string) error {
			// TODO: Handle on submit action, for example, using channel.
			pp.Println("input text", text)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return input, err
}

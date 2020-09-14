package gui

import (
	"time"

	"github.com/k0kubun/pp"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const (
	redrawInterval = 250 * time.Millisecond
)

type widgets struct {
	attackButton   *button.Button
	urlInput       *textinput.TextInput
	rateLimitInput *textinput.TextInput
	durationInput  *textinput.TextInput
	latencyChart   *linechart.LineChart
	reportText     *text.Text
	navi           *text.Text
}

func newWidgets() (*widgets, error) {
	latencyChart, err := newLineChart()
	if err != nil {
		return nil, err
	}
	reportText, err := newRollText("Give the target URL and press Enter, then the attack will be launched.")
	if err != nil {
		return nil, err
	}
	navi, err := newRollText("Ctrl-c: quit, Enter: attack")
	if err != nil {
		return nil, err
	}
	urlInput, err := newTextInput("Target URL: ", "target URL")
	if err != nil {
		return nil, err
	}
	rateLimitInput, err := newTextInput("Rate limit: ", "number of requests per second (default 50)")
	if err != nil {
		return nil, err
	}
	attackButton, err := newButton("Attack", func() error {
		target := urlInput.Read()
		pp.Println(target)
		// TODO: Call Attack.
		return nil
	})
	return &widgets{
		attackButton:   attackButton,
		urlInput:       urlInput,
		rateLimitInput: rateLimitInput,
		latencyChart:   latencyChart,
		reportText:     reportText,
		navi:           navi,
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

func newButton(text string, onSubmit func() error) (*button.Button, error) {
	return button.New(text, onSubmit,
		button.GlobalKey(keyboard.KeyEnter),
		button.FillColor(cell.ColorNumber(196)),
	)
}

func newTextInput(label, placeHolder string) (*textinput.TextInput, error) {
	return textinput.New(
		textinput.Label(label, cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(99),
		textinput.PlaceHolder(placeHolder),
	)
}

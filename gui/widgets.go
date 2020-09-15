package gui

import (
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const (
	redrawInterval = 250 * time.Millisecond
)

type widgets struct {
	urlInput       *textinput.TextInput
	rateLimitInput *textinput.TextInput
	durationInput  *textinput.TextInput
	methodInput    *textinput.TextInput
	bodyInput      *textinput.TextInput
	headerInput    *textinput.TextInput
	timeoutInput   *textinput.TextInput
	latencyChart   *linechart.LineChart
	reportText     *text.Text
	progressGauge  *gauge.Gauge
	navi           *text.Text
}

func newWidgets() (*widgets, error) {
	latencyChart, err := newLineChart()
	if err != nil {
		return nil, err
	}
	reportText, err := newText("Give the target URL and press Enter, then the attack will be launched.")
	if err != nil {
		return nil, err
	}
	navi, err := newText("Ctrl-c: quit, Enter: attack")
	if err != nil {
		return nil, err
	}
	urlInput, err := newTextInput("Target URL: ", "", 60)
	if err != nil {
		return nil, err
	}
	rateLimitInput, err := newTextInput("Rate limit: ", "50", 30)
	if err != nil {
		return nil, err
	}
	durationInput, err := newTextInput("Duration: ", "10s", 30)
	if err != nil {
		return nil, err
	}
	methodInput, err := newTextInput("Method: ", "GET", 30)
	if err != nil {
		return nil, err
	}
	bodyInput, err := newTextInput("Body: ", "", 30)
	if err != nil {
		return nil, err
	}
	headerInput, err := newTextInput("Header: ", "", 30)
	if err != nil {
		return nil, err
	}
	timeoutInput, err := newTextInput("Timeout: ", "30s", 30)
	if err != nil {
		return nil, err
	}
	progressGauge, err := newGauge()
	if err != nil {
		return nil, err
	}
	return &widgets{
		urlInput:       urlInput,
		rateLimitInput: rateLimitInput,
		durationInput:  durationInput,
		methodInput:    methodInput,
		bodyInput:      bodyInput,
		headerInput:    headerInput,
		timeoutInput:   timeoutInput,
		latencyChart:   latencyChart,
		reportText:     reportText,
		progressGauge:  progressGauge,
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

func newText(s string) (*text.Text, error) {
	t, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if err := t.Write(s); err != nil {
		return nil, err
	}
	return t, nil
}

func newTextInput(label, placeHolder string, cells int) (*textinput.TextInput, error) {
	return textinput.New(
		textinput.Label(label, cell.FgColor(cell.ColorWhite)),
		textinput.MaxWidthCells(cells),
		textinput.PlaceHolder(placeHolder),
	)
}

func newGauge() (*gauge.Gauge, error) {
	return gauge.New(
		gauge.Height(1),
		gauge.Border(linestyle.Light),
		gauge.BorderTitle("Progress"),
	)
}

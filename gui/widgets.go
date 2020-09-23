package gui

import (
	"strconv"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"

	"github.com/nakabonne/ali/attacker"
)

type TextInput interface {
	widgetapi.Widget
	Read() string
}

type LineChart interface {
	widgetapi.Widget
	Series(label string, values []float64, opts ...linechart.SeriesOption) error
}

type Text interface {
	widgetapi.Widget
	Write(text string, wOpts ...text.WriteOption) error
}

type Gauge interface {
	widgetapi.Widget
	Percent(p int, opts ...gauge.Option) error
}

type widgets struct {
	urlInput       TextInput
	rateLimitInput TextInput
	durationInput  TextInput
	methodInput    TextInput
	bodyInput      TextInput
	headerInput    TextInput
	timeoutInput   TextInput

	latencyChart LineChart

	messageText   Text
	latenciesText Text
	bytesText     Text
	othersText    Text

	progressGauge Gauge
	navi          Text
}

func newWidgets() (*widgets, error) {
	latencyChart, err := newLineChart()
	if err != nil {
		return nil, err
	}
	messageText, err := newText("Give the target URL and press Enter")
	if err != nil {
		return nil, err
	}
	latenciesText, err := newText("")
	if err != nil {
		return nil, err
	}
	bytesText, err := newText("")
	if err != nil {
		return nil, err
	}
	othersText, err := newText("")
	if err != nil {
		return nil, err
	}
	navi, err := newText("Ctrl-C: quit, Enter: attack")
	if err != nil {
		return nil, err
	}
	urlInput, err := newTextInput("Target URL: ", "http://", 200)
	if err != nil {
		return nil, err
	}
	rateLimitInput, err := newTextInput("Rate limit: ", strconv.Itoa(attacker.DefaultRate), 50)
	if err != nil {
		return nil, err
	}
	durationInput, err := newTextInput("Duration: ", attacker.DefaultDuration.String(), 50)
	if err != nil {
		return nil, err
	}
	methodInput, err := newTextInput("Method: ", attacker.DefaultMethod, 50)
	if err != nil {
		return nil, err
	}
	bodyInput, err := newTextInput("Body: ", "", 200)
	if err != nil {
		return nil, err
	}
	headerInput, err := newTextInput("Header: ", "", 50)
	if err != nil {
		return nil, err
	}
	timeoutInput, err := newTextInput("Timeout: ", attacker.DefaultTimeout.String(), 50)
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
		messageText:    messageText,
		latenciesText:  latenciesText,
		bytesText:      bytesText,
		othersText:     othersText,
		progressGauge:  progressGauge,
		navi:           navi,
	}, nil
}

func newLineChart() (LineChart, error) {
	return linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
}

func newText(s string) (Text, error) {
	t, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if s != "" {
		if err := t.Write(s); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func newTextInput(label, placeHolder string, cells int) (TextInput, error) {
	return textinput.New(
		//textinput.Label(label, cell.FgColor(cell.ColorWhite)),
		//textinput.Border(linestyle.Double),
		//textinput.BorderColor(cell.ColorGreen),
		textinput.FillColor(cell.ColorBlue),
		textinput.MaxWidthCells(cells),
		textinput.PlaceHolder(placeHolder),
		textinput.PlaceHolderColor(cell.ColorDefault),
	)
}

func newGauge() (Gauge, error) {
	return gauge.New(
		//gauge.Height(1),
		gauge.Border(linestyle.None),
		//gauge.BorderTitle("Progress"),
	)
}

package gui

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/nakabonne/ali/attacker"
)

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
	latencyChart LineChart

	paramsText      Text
	latenciesText   Text
	bytesText       Text
	statusCodesText Text
	errorsText      Text
	othersText      Text

	progressGauge Gauge
	navi          Text
}

func newWidgets(targetURL string, opts *attacker.Options) (*widgets, error) {
	latencyChart, err := newLineChart()
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
	statusCodesText, err := newText("")
	if err != nil {
		return nil, err
	}
	errorsText, err := newText("")
	if err != nil {
		return nil, err
	}
	othersText, err := newText("")
	if err != nil {
		return nil, err
	}

	paramsText, err := newText(makeParamsText(targetURL, opts))
	if err != nil {
		return nil, err
	}

	navi, err := newText("Ctrl-C: quit, Enter: attack")
	if err != nil {
		return nil, err
	}
	progressGauge, err := newGauge()
	if err != nil {
		return nil, err
	}
	return &widgets{
		latencyChart:    latencyChart,
		paramsText:      paramsText,
		latenciesText:   latenciesText,
		bytesText:       bytesText,
		statusCodesText: statusCodesText,
		errorsText:      errorsText,
		othersText:      othersText,
		progressGauge:   progressGauge,
		navi:            navi,
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

func newGauge() (Gauge, error) {
	return gauge.New(
		//gauge.Height(1),
		gauge.Border(linestyle.None),
		//gauge.BorderTitle("Progress"),
	)
}

func makeParamsText(targetURL string, opts *attacker.Options) string {
	return fmt.Sprintf(`Target: %s
Rate: %d
Duration: %v
Method: %s
`, targetURL, opts.Rate, opts.Duration, opts.Method)
}

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

type chartLegend struct {
	text  Text
	cellOpts []cell.Option
}

type widgets struct {
	latencyChart LineChart

	paramsText      Text
	latenciesText   Text
	bytesText       Text
	statusCodesText Text
	errorsText      Text
	othersText      Text

	percentilesChart LineChart
	p99              chartLegend
	p95              chartLegend
	p90              chartLegend
	p50              chartLegend

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

	p99Color := cell.FgColor(cell.ColorNumber(104))
	p99Text, err := newText("p99", text.WriteCellOpts(p99Color))
	if err != nil {
		return nil, err
	}
	p95Color := cell.FgColor(cell.ColorNumber(89))
	p95Text, err := newText("p95", text.WriteCellOpts(p95Color))
	if err != nil {
		return nil, err
	}
	p90Color := cell.FgColor(cell.ColorNumber(88))
	p90Text, err := newText("p90", text.WriteCellOpts(p90Color))
	if err != nil {
		return nil, err
	}
	p50Color := cell.FgColor(cell.ColorNumber(87))
	p50Text, err := newText("p50", text.WriteCellOpts(p50Color))
	if err != nil {
		return nil, err
	}
	percentilesChart, err := newLineChart()
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
		latencyChart:     latencyChart,
		paramsText:       paramsText,
		latenciesText:    latenciesText,
		bytesText:        bytesText,
		statusCodesText:  statusCodesText,
		errorsText:       errorsText,
		othersText:       othersText,
		progressGauge:    progressGauge,
		percentilesChart: percentilesChart,
		p99:              chartLegend{p99Text, []cell.Option{p99Color}},
		p95:              chartLegend{p95Text, []cell.Option{p95Color}},
		p90:              chartLegend{p90Text, []cell.Option{p90Color}},
		p50:              chartLegend{p50Text, []cell.Option{p50Color}},
		navi:             navi,
	}, nil
}

func newLineChart() (LineChart, error) {
	return linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
}

func newText(s string, opts ...text.WriteOption) (Text, error) {
	t, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if s != "" {
		if err := t.Write(s, opts...); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func newGauge() (Gauge, error) {
	return gauge.New(
		gauge.Border(linestyle.None),
	)
}

func makeParamsText(targetURL string, opts *attacker.Options) string {
	return fmt.Sprintf(`Target: %s
Rate: %d
Duration: %v
Method: %s
`, targetURL, opts.Rate, opts.Duration, opts.Method)
}

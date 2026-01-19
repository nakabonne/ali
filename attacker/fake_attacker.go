package attacker

import (
	"context"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type FakeAttacker struct {
	rate     int
	duration time.Duration
	method   string
}

func (f *FakeAttacker) Attack(ctx context.Context, metricsCh chan *Metrics) error {
	return nil
}

func (f *FakeAttacker) Rate() int {
	return f.rate
}

func (f *FakeAttacker) Duration() time.Duration {
	return f.duration
}

func (f *FakeAttacker) Method() string {
	return f.method
}

type fakeBackedAttacker struct {
	results []*vegeta.Result
}

func (f *fakeBackedAttacker) Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result {
	resultCh := make(chan *vegeta.Result)
	go func() {
		defer close(resultCh)
		for _, r := range f.results {
			resultCh <- r
		}
	}()
	return resultCh
}

func (f *fakeBackedAttacker) Stop() {
}

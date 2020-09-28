package attacker

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type fakeAttacker struct {
	results []*vegeta.Result
}

func (f *fakeAttacker) Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result {
	resultCh := make(chan *vegeta.Result)
	go func() {
		defer close(resultCh)
		for _, r := range f.results {
			resultCh <- r
		}
	}()
	return resultCh
}

func (f *fakeAttacker) Stop() {
}

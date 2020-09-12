package attacker

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type fakeAttacker struct {
}

func (f *fakeAttacker) Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result {
	return nil
}

func (f *fakeAttacker) Stop() {
	return
}

package storage

import "time"

type FakeStorage struct {
	Values []float64
	err    error
}

func (f *FakeStorage) Insert(_ *Result) error {
	return f.err
}

func (f *FakeStorage) Select(_ string, _, _ time.Time) ([]float64, error) {
	return f.Values, f.err
}

package storage

import (
	"errors"
	"log"
	"time"

	"github.com/nakabonne/tstorage"
)

const (
	LatencyMetricName = "latency"
	P50MetricName     = "p50"
	P90MetricName     = "p90"
	P95MetricName     = "p95"
	P99MetricName     = "p99"
)

// Storage provides goroutine safe capabilities of insertion into and retrieval from the time-series storage.
// Backed by "nakabonne/tstorage"
type Storage interface {
	Writer
	Reader
}

type Writer interface {
	Insert(result *Result) error
}

type Reader interface {
	Select(metric string, start, end time.Time) ([]float64, error)
}

// Result contains the results of a single HTTP request.
type Result struct {
	Code      uint16
	Timestamp time.Time
	Latency   time.Duration
	P50       time.Duration
	P90       time.Duration
	P95       time.Duration
	P99       time.Duration
}

func NewStorage(partitionDuration time.Duration) (Storage, error) {
	s, err := tstorage.NewStorage(
		tstorage.WithLogger(log.Default()),
		tstorage.WithPartitionDuration(partitionDuration),
	)
	if err != nil {
		return nil, err
	}
	return &storage{backend: s}, nil
}

type storage struct {
	backend tstorage.Storage
}

// Insert writes the given result to the backend storage.
// The unit of value will be converted in milliseconds.
func (s *storage) Insert(result *Result) error {
	// Convert timestamp into unix time in nanoseconds.
	timestamp := result.Timestamp.UnixNano()
	// TODO: Think about how to handle code
	/*
		labels := []tstorage.Label{
			{
				Name:  codeLabelName,
				Value: strconv.Itoa(int(result.Code)),
			},
		}
	*/
	rows := []tstorage.Row{
		{
			Metric: LatencyMetricName,
			DataPoint: tstorage.DataPoint{
				Timestamp: timestamp,
				Value:     float64(result.Latency.Milliseconds()),
			},
		},
		{
			Metric: P50MetricName,
			DataPoint: tstorage.DataPoint{
				Timestamp: timestamp,
				Value:     float64(result.P50.Milliseconds()),
			},
		},
		{
			Metric: P90MetricName,
			DataPoint: tstorage.DataPoint{
				Timestamp: timestamp,
				Value:     float64(result.P90.Milliseconds()),
			},
		},
		{
			Metric: P95MetricName,
			DataPoint: tstorage.DataPoint{
				Timestamp: timestamp,
				Value:     float64(result.P95.Milliseconds()),
			},
		},
		{
			Metric: P99MetricName,
			DataPoint: tstorage.DataPoint{
				Timestamp: timestamp,
				Value:     float64(result.P99.Milliseconds()),
			},
		},
	}
	return s.backend.InsertRows(rows)
}

func (s *storage) Select(metric string, start, end time.Time) ([]float64, error) {
	// Convert timestamp into unix time in nanoseconds.
	points, err := s.backend.Select(metric, nil, start.UnixNano(), end.UnixNano())
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		return []float64{}, nil
	}
	if err != nil {
		return nil, err
	}
	values := make([]float64, len(points))
	for i := range points {
		values[i] = points[i].Value
	}
	return values, nil
}

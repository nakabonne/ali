package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nakabonne/ali/attacker"
	"github.com/nakabonne/ali/export"
	"github.com/nakabonne/ali/gui"
	"github.com/nakabonne/ali/storage"
)

func TestExportCLI_NoExportTo_NoFilesCreated(t *testing.T) {
	origRunGUI := runGUI
	origNewAttacker := newAttacker
	defer func() {
		runGUI = origRunGUI
		newAttacker = origNewAttacker
	}()

	runGUI = func(string, storage.Reader, attacker.Attacker, gui.Options) error {
		return nil
	}
	var gotExporter *export.FileExporter
	newAttacker = func(_ storage.Writer, _ string, opts *attacker.Options) (attacker.Attacker, error) {
		gotExporter = opts.Exporter
		if gotExporter != nil {
			return nil, fmt.Errorf("unexpected exporter %v", gotExporter)
		}
		return &attacker.FakeAttacker{}, nil
	}

	buf := &bytes.Buffer{}
	c := defaultCLI(buf)
	exitCode := c.run([]string{"https://example.com/"})
	require.Equal(t, 0, exitCode)
	require.Nil(t, gotExporter)
}

func TestExportCLI_CreateDirAndFiles(t *testing.T) {
	origRunGUI := runGUI
	origNewAttacker := newAttacker
	defer func() {
		runGUI = origRunGUI
		newAttacker = origNewAttacker
	}()

	resultsDir := filepath.Join(t.TempDir(), "results")
	runGUI = func(_ string, _ storage.Reader, a attacker.Attacker, _ gui.Options) error {
		metricsCh := make(chan *attacker.Metrics, 10)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.Attack(ctx, metricsCh)
	}
	newAttacker = func(_ storage.Writer, target string, opts *attacker.Options) (attacker.Attacker, error) {
		if opts.Exporter == nil {
			return nil, fmt.Errorf("exporter is required for this test")
		}
		meta := export.Meta{
			ID:        "00000000-0000-0000-0000-000000000000",
			TargetURL: target,
			Method:    opts.Method,
			Rate:      opts.Rate,
			Duration:  opts.Duration,
		}
		results := []export.Result{
			{
				Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, time.FixedZone("JST", 9*60*60)),
				LatencyNS:  123,
				StatusCode: 200,
			},
		}
		return &exportingAttacker{
			exporter: opts.Exporter,
			meta:     meta,
			results:  results,
			summary:  export.Summary{},
		}, nil
	}

	buf := &bytes.Buffer{}
	c := defaultCLI(buf)
	c.exportTo = resultsDir
	exitCode := c.run([]string{"https://example.com/"})
	require.Equal(t, 0, exitCode)
	require.FileExists(t, filepath.Join(resultsDir, "results.csv"))
	require.FileExists(t, filepath.Join(resultsDir, "summary-00000000-0000-0000-0000-000000000000.json"))
}

func TestExportCLI_ExistingDirFails(t *testing.T) {
	origRunGUI := runGUI
	origNewAttacker := newAttacker
	defer func() {
		runGUI = origRunGUI
		newAttacker = origNewAttacker
	}()

	runGUICalled := false
	newAttackerCalled := false
	runGUI = func(string, storage.Reader, attacker.Attacker, gui.Options) error {
		runGUICalled = true
		return nil
	}
	newAttacker = func(storage.Writer, string, *attacker.Options) (attacker.Attacker, error) {
		newAttackerCalled = true
		return &attacker.FakeAttacker{}, nil
	}

	resultsDir := filepath.Join(t.TempDir(), "results")
	require.NoError(t, os.MkdirAll(resultsDir, 0o755))
	sentinelPath := filepath.Join(resultsDir, "sentinel.txt")
	require.NoError(t, os.WriteFile(sentinelPath, []byte("keep"), 0o644))

	buf := &bytes.Buffer{}
	c := defaultCLI(buf)
	c.exportTo = resultsDir
	exitCode := c.run([]string{"https://example.com/"})
	require.Equal(t, 1, exitCode)
	require.False(t, runGUICalled)
	require.False(t, newAttackerCalled)

	content, err := os.ReadFile(sentinelPath)
	require.NoError(t, err)
	require.Equal(t, "keep", string(content))
}

type exportingAttacker struct {
	exporter *export.FileExporter
	meta     export.Meta
	results  []export.Result
	summary  export.Summary
}

func (e *exportingAttacker) Attack(ctx context.Context, metricsCh chan *attacker.Metrics) error {
	run, err := e.exporter.StartRun(e.meta)
	if err != nil {
		return err
	}
	for _, res := range e.results {
		if err := run.WriteResult(res); err != nil {
			return err
		}
	}
	return run.Close(e.summary)
}

func (e *exportingAttacker) Rate() int {
	return e.meta.Rate
}

func (e *exportingAttacker) Duration() time.Duration {
	return e.meta.Duration
}

func (e *exportingAttacker) Method() string {
	return e.meta.Method
}

func defaultCLI(buf *bytes.Buffer) *cli {
	return &cli{
		method:         "GET",
		localAddress:   "0.0.0.0",
		queryRange:     gui.DefaultQueryRange,
		redrawInterval: gui.DefaultRedrawInterval,
		stdout:         buf,
		stderr:         buf,
	}
}

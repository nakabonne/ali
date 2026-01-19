package export

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileExporter_Basic(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(dir)
	zone := time.FixedZone("JST", 9*60*60)

	run, err := exporter.StartRun(Meta{
		ID:        "00000000-0000-0000-0000-000000000000",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      50,
		Duration:  2 * time.Second,
	})
	require.NoError(t, err)

	results := []Result{
		{
			Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
			LatencyNS:  18234567,
			StatusCode: 200,
		},
		{
			Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 20*int(time.Millisecond), zone),
			LatencyNS:  44900123,
			StatusCode: 200,
		},
		{
			Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 41*int(time.Millisecond), zone),
			LatencyNS:  935489752,
			StatusCode: 500,
		},
	}
	for _, res := range results {
		require.NoError(t, run.WriteResult(res))
	}

	summary := Summary{
		Target: TargetSummary{
			URL:    "https://example.com/",
			Method: "GET",
		},
		Parameters: ParametersSummary{
			Rate:            50,
			DurationSeconds: 2,
		},
		Timing: TimingSummary{
			Earliest: time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
			Latest:   time.Date(2021, 3, 13, 15, 20, 45, 0, zone),
		},
		Requests: RequestsSummary{
			Count:        100,
			SuccessRatio: 0.98,
		},
		Throughput: 48.24,
		LatencyMS: LatencySummary{
			Total: 44000,
			Mean:  447.88,
			P50:   445.46,
			P90:   806.58,
			P95:   849.89,
			P99:   935.49,
			Max:   965.4,
			Min:   55.32,
		},
		Bytes: BytesSummary{
			In: BytesFlowSummary{
				Total: 2325200,
				Mean:  23252,
			},
			Out: BytesFlowSummary{
				Total: 0,
				Mean:  0,
			},
		},
		StatusCodes: StatusCodesSummary{
			"200": 98,
			"500": 2,
		},
	}
	require.NoError(t, run.Close(summary))

	wantResults := readGolden(t, filepath.Join("..", "testdata", "export", "basic", "results.csv"))
	gotResults := readFile(t, filepath.Join(dir, resultsFilename))
	require.Equal(t, string(wantResults), string(gotResults))

	wantSummary := readGolden(t, filepath.Join("..", "testdata", "export", "basic", "summary-00000000-0000-0000-0000-000000000000.json"))
	gotSummary := readFile(t, filepath.Join(dir, summaryFilename("00000000-0000-0000-0000-000000000000")))
	require.Equal(t, string(wantSummary), string(gotSummary))
}

func TestFileExporter_Quotes(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(dir)
	zone := time.FixedZone("JST", 9*60*60)

	run, err := exporter.StartRun(Meta{
		ID:        "11111111-1111-1111-1111-111111111111",
		TargetURL: "https://example.com/hello, \"world\"",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)

	require.NoError(t, run.WriteResult(Result{
		Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
		LatencyNS:  123,
		StatusCode: 200,
	}))
	require.NoError(t, run.Close(Summary{}))

	wantResults := readGolden(t, filepath.Join("..", "testdata", "export", "quotes", "results.csv"))
	gotResults := readFile(t, filepath.Join(dir, resultsFilename))
	require.Equal(t, string(wantResults), string(gotResults))
}

func TestFileExporter_EmptyResults(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(dir)

	run, err := exporter.StartRun(Meta{
		ID:        "33333333-3333-3333-3333-333333333333",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, run.Close(Summary{}))

	wantResults := readGolden(t, filepath.Join("..", "testdata", "export", "empty", "results.csv"))
	gotResults := readFile(t, filepath.Join(dir, resultsFilename))
	require.Equal(t, string(wantResults), string(gotResults))
}

func TestFileExporter_NaNInf(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(dir)
	zone := time.FixedZone("JST", 9*60*60)

	run, err := exporter.StartRun(Meta{
		ID:        "22222222-2222-2222-2222-222222222222",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)

	require.NoError(t, run.WriteResult(Result{
		Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
		LatencyNS:  math.NaN(),
		StatusCode: 200,
	}))
	require.NoError(t, run.Close(Summary{}))

	wantResults := readGolden(t, filepath.Join("..", "testdata", "export", "naninf", "results.csv"))
	gotResults := readFile(t, filepath.Join(dir, resultsFilename))
	require.Equal(t, string(wantResults), string(gotResults))
}

func TestFileExporter_AppendsRuns(t *testing.T) {
	dir := t.TempDir()
	exporter := NewFileExporter(dir)
	zone := time.FixedZone("JST", 9*60*60)

	run1, err := exporter.StartRun(Meta{
		ID:        "aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, run1.WriteResult(Result{
		Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
		LatencyNS:  1,
		StatusCode: 200,
	}))
	require.NoError(t, run1.Close(Summary{}))

	run2, err := exporter.StartRun(Meta{
		ID:        "bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, run2.WriteResult(Result{
		Timestamp:  time.Date(2021, 3, 13, 15, 20, 44, 0, zone),
		LatencyNS:  2,
		StatusCode: 200,
	}))
	require.NoError(t, run2.Close(Summary{}))

	records := readCSV(t, filepath.Join(dir, resultsFilename))
	require.Len(t, records, 3)
	require.Equal(t, resultsHeader, records[0])
	require.Equal(t, "aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa", records[1][0])
	require.Equal(t, "bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbbb", records[2][0])
}

func TestFileExporter_AtomicResultsWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod semantics are not reliable on windows")
	}

	dir := t.TempDir()
	original := []byte("id,timestamp,latency_ns,url,method,status_code\n")
	resultsPath := filepath.Join(dir, resultsFilename)
	require.NoError(t, os.WriteFile(resultsPath, original, 0o644))

	exporter := NewFileExporter(dir)
	zone := time.FixedZone("JST", 9*60*60)
	run, err := exporter.StartRun(Meta{
		ID:        "cccccccc-cccc-cccc-cccc-cccccccccccc",
		TargetURL: "https://example.com/",
		Method:    "GET",
		Rate:      1,
		Duration:  time.Second,
	})
	require.NoError(t, err)
	require.NoError(t, run.WriteResult(Result{
		Timestamp:  time.Date(2021, 3, 13, 15, 20, 43, 0, zone),
		LatencyNS:  3,
		StatusCode: 200,
	}))

	require.NoError(t, os.Chmod(dir, 0o555))
	err = run.Close(Summary{})
	require.Error(t, err)
	require.NoError(t, os.Chmod(dir, 0o755))

	got := readFile(t, resultsPath)
	require.Equal(t, string(original), string(got))
	_, err = os.Stat(filepath.Join(dir, summaryFilename("cccccccc-cccc-cccc-cccc-cccccccccccc")))
	require.True(t, os.IsNotExist(err))
}

func readGolden(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return content
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return content
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	return records
}

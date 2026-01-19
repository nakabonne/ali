package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	require.NoError(t, err)

	return records
}

func TestExportGoldenResultsCSVBasic(t *testing.T) {
	path := filepath.Join("testdata", "export", "basic", "results.csv")
	records := readCSV(t, path)

	require.GreaterOrEqual(t, len(records), 2)
	require.Equal(t, []string{"id", "timestamp", "latency_ns", "url", "method", "status_code"}, records[0])

	for i, row := range records[1:] {
		require.Len(t, row, 6, "row %d", i+1)
		require.Equal(t, "00000000-0000-0000-0000-000000000000", row[0])
		_, err := time.Parse(time.RFC3339, row[1])
		require.NoError(t, err)
		_, err = strconv.ParseInt(row[2], 10, 64)
		require.NoError(t, err)
		require.NotEmpty(t, row[3])
		require.NotEmpty(t, row[4])
		_, err = strconv.ParseUint(row[5], 10, 16)
		require.NoError(t, err)
	}
}

func TestExportGoldenResultsCSVQuotes(t *testing.T) {
	path := filepath.Join("testdata", "export", "quotes", "results.csv")
	records := readCSV(t, path)

	require.Len(t, records, 2)
	require.Equal(t, []string{"id", "timestamp", "latency_ns", "url", "method", "status_code"}, records[0])
	require.Equal(t, "https://example.com/hello, \"world\"", records[1][3])
}

func TestExportGoldenResultsCSVEmpty(t *testing.T) {
	path := filepath.Join("testdata", "export", "empty", "results.csv")
	records := readCSV(t, path)

	require.Len(t, records, 1)
	require.Equal(t, []string{"id", "timestamp", "latency_ns", "url", "method", "status_code"}, records[0])
}

func TestExportGoldenResultsCSVNaNInf(t *testing.T) {
	path := filepath.Join("testdata", "export", "naninf", "results.csv")
	records := readCSV(t, path)

	require.Len(t, records, 2)
	require.Equal(t, []string{"id", "timestamp", "latency_ns", "url", "method", "status_code"}, records[0])
	require.Equal(t, "", records[1][2])
}

func TestExportGoldenSummaryJSONSchema(t *testing.T) {
	path := filepath.Join("testdata", "export", "basic", "summary-00000000-0000-0000-0000-000000000000.json")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, json.Unmarshal(content, &doc))

	target := mustMap(t, doc["target"], "target")
	require.NotEmpty(t, mustString(t, target["url"], "target.url"))
	require.NotEmpty(t, mustString(t, target["method"], "target.method"))

	params := mustMap(t, doc["parameters"], "parameters")
	mustNumber(t, params["rate"], "parameters.rate")
	mustNumber(t, params["duration_seconds"], "parameters.duration_seconds")

	timing := mustMap(t, doc["timing"], "timing")
	_, err = time.Parse(time.RFC3339, mustString(t, timing["earliest"], "timing.earliest"))
	require.NoError(t, err)
	_, err = time.Parse(time.RFC3339, mustString(t, timing["latest"], "timing.latest"))
	require.NoError(t, err)

	requests := mustMap(t, doc["requests"], "requests")
	mustNumber(t, requests["count"], "requests.count")
	mustNumber(t, requests["success_ratio"], "requests.success_ratio")

	mustNumber(t, doc["throughput"], "throughput")

	latency := mustMap(t, doc["latency_ms"], "latency_ms")
	for _, key := range []string{"total", "mean", "p50", "p90", "p95", "p99", "max", "min"} {
		mustNumber(t, latency[key], "latency_ms."+key)
	}

	bytes := mustMap(t, doc["bytes"], "bytes")
	bytesIn := mustMap(t, bytes["in"], "bytes.in")
	mustNumber(t, bytesIn["total"], "bytes.in.total")
	mustNumber(t, bytesIn["mean"], "bytes.in.mean")

	bytesOut := mustMap(t, bytes["out"], "bytes.out")
	mustNumber(t, bytesOut["total"], "bytes.out.total")
	mustNumber(t, bytesOut["mean"], "bytes.out.mean")

	statusCodes := mustMap(t, doc["status_codes"], "status_codes")
	require.NotEmpty(t, statusCodes)
}

func mustMap(t *testing.T, value interface{}, name string) map[string]interface{} {
	t.Helper()
	m, ok := value.(map[string]interface{})
	require.True(t, ok, "%s must be an object", name)
	return m
}

func mustString(t *testing.T, value interface{}, name string) string {
	t.Helper()
	s, ok := value.(string)
	require.True(t, ok, "%s must be a string", name)
	return s
}

func mustNumber(t *testing.T, value interface{}, name string) float64 {
	t.Helper()
	n, ok := value.(float64)
	require.True(t, ok, "%s must be a number", name)
	return n
}

package export

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

const (
	resultsFilename = "results.csv"
)

var resultsHeader = []string{"id", "timestamp", "latency_ns", "url", "method", "status_code"}

type Meta struct {
	ID        string
	TargetURL string
	Method    string
	Rate      int
	Duration  time.Duration
}

type Result struct {
	Timestamp  time.Time
	LatencyNS  float64
	URL        string
	Method     string
	StatusCode uint16
}

type Summary struct {
	Target      TargetSummary      `json:"target"`
	Parameters  ParametersSummary  `json:"parameters"`
	Timing      TimingSummary      `json:"timing"`
	Requests    RequestsSummary    `json:"requests"`
	Throughput  float64            `json:"throughput"`
	LatencyMS   LatencySummary     `json:"latency_ms"`
	Bytes       BytesSummary       `json:"bytes"`
	StatusCodes StatusCodesSummary `json:"status_codes"`
}

type TargetSummary struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

type ParametersSummary struct {
	Rate            int     `json:"rate"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type TimingSummary struct {
	Earliest time.Time `json:"earliest"`
	Latest   time.Time `json:"latest"`
}

type RequestsSummary struct {
	Count        uint64  `json:"count"`
	SuccessRatio float64 `json:"success_ratio"`
}

type LatencySummary struct {
	Total float64 `json:"total"`
	Mean  float64 `json:"mean"`
	P50   float64 `json:"p50"`
	P90   float64 `json:"p90"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
}

type BytesSummary struct {
	In  BytesFlowSummary `json:"in"`
	Out BytesFlowSummary `json:"out"`
}

type BytesFlowSummary struct {
	Total uint64  `json:"total"`
	Mean  float64 `json:"mean"`
}

type StatusCodesSummary map[string]int

func (s StatusCodesSummary) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(s[key]))
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

type FileExporter struct {
	dir string
}

func NewFileExporter(dir string) *FileExporter {
	return &FileExporter{dir: dir}
}

type Run struct {
	meta Meta

	resultsPath string
	summaryPath string

	resultsFile *os.File
	resultsBuf  *bufio.Writer
	resultsCSV  *csv.Writer

	tempResultsPath string
	closed          bool
}

func (e *FileExporter) StartRun(meta Meta) (*Run, error) {
	if meta.ID == "" {
		return nil, errors.New("export run id is required")
	}
	if e.dir == "" {
		return nil, errors.New("export directory is required")
	}
	resultsPath := filepath.Join(e.dir, resultsFilename)
	summaryPath := filepath.Join(e.dir, summaryFilename(meta.ID))

	tmpFile, err := os.CreateTemp(e.dir, ".results.csv.")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp results file in %q: %w", e.dir, err)
	}
	tempResultsPath := tmpFile.Name()
	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tempResultsPath)
		return nil, fmt.Errorf("failed to chmod temp results file %q: %w", tempResultsPath, err)
	}

	var resultsExist bool
	info, err := os.Stat(resultsPath)
	if err == nil {
		if info.IsDir() {
			_ = tmpFile.Close()
			_ = os.Remove(tempResultsPath)
			return nil, fmt.Errorf("results path %q is a directory", resultsPath)
		}
		resultsExist = true
		src, err := os.Open(resultsPath)
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tempResultsPath)
			return nil, fmt.Errorf("failed to open results file %q: %w", resultsPath, err)
		}
		if _, err := io.Copy(tmpFile, src); err != nil {
			_ = src.Close()
			_ = tmpFile.Close()
			_ = os.Remove(tempResultsPath)
			return nil, fmt.Errorf("failed to copy results file %q: %w", resultsPath, err)
		}
		if err := src.Close(); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tempResultsPath)
			return nil, fmt.Errorf("failed to close results file %q: %w", resultsPath, err)
		}
	} else if !os.IsNotExist(err) {
		_ = tmpFile.Close()
		_ = os.Remove(tempResultsPath)
		return nil, fmt.Errorf("failed to stat results file %q: %w", resultsPath, err)
	}

	buf := bufio.NewWriter(tmpFile)
	writer := csv.NewWriter(buf)
	if !resultsExist {
		if err := writer.Write(resultsHeader); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tempResultsPath)
			return nil, fmt.Errorf("failed to write results header to %q: %w", resultsPath, err)
		}
	}

	return &Run{
		meta:            meta,
		resultsPath:     resultsPath,
		summaryPath:     summaryPath,
		resultsFile:     tmpFile,
		resultsBuf:      buf,
		resultsCSV:      writer,
		tempResultsPath: tempResultsPath,
	}, nil
}

func (r *Run) WriteResult(res Result) error {
	if r.closed {
		return errors.New("export run already closed")
	}
	url := res.URL
	if url == "" {
		url = r.meta.TargetURL
	}
	method := res.Method
	if method == "" {
		method = r.meta.Method
	}
	record := []string{
		r.meta.ID,
		res.Timestamp.Format(time.RFC3339Nano),
		formatLatencyNS(res.LatencyNS),
		url,
		method,
		strconv.FormatUint(uint64(res.StatusCode), 10),
	}
	if err := r.resultsCSV.Write(record); err != nil {
		_ = r.Abort()
		return fmt.Errorf("failed to write results to %q: %w", r.resultsPath, err)
	}
	return nil
}

func (r *Run) Close(summary Summary) error {
	if r.closed {
		return errors.New("export run already closed")
	}
	r.resultsCSV.Flush()
	if err := r.resultsCSV.Error(); err != nil {
		_ = r.Abort()
		return fmt.Errorf("failed to flush results to %q: %w", r.resultsPath, err)
	}
	if err := r.resultsBuf.Flush(); err != nil {
		_ = r.Abort()
		return fmt.Errorf("failed to flush results buffer to %q: %w", r.resultsPath, err)
	}
	if err := r.resultsFile.Sync(); err != nil {
		_ = r.Abort()
		return fmt.Errorf("failed to sync results file %q: %w", r.resultsPath, err)
	}
	if err := r.resultsFile.Close(); err != nil {
		_ = r.Abort()
		return fmt.Errorf("failed to close results file %q: %w", r.resultsPath, err)
	}
	if err := os.Rename(r.tempResultsPath, r.resultsPath); err != nil {
		_ = os.Remove(r.tempResultsPath)
		return fmt.Errorf("failed to replace results file %q: %w", r.resultsPath, err)
	}
	if err := writeSummary(r.summaryPath, summary); err != nil {
		return err
	}
	r.closed = true
	return nil
}

func (r *Run) Abort() error {
	if r.closed {
		return nil
	}
	_ = r.resultsFile.Close()
	if err := os.Remove(r.tempResultsPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	r.closed = true
	return nil
}

func writeSummary(path string, summary Summary) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".summary.")
	if err != nil {
		return fmt.Errorf("failed to create temp summary file in %q: %w", dir, err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to chmod temp summary file %q: %w", tmpPath, err)
	}

	enc := json.NewEncoder(tmpFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(summary); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to encode summary to %q: %w", path, err)
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync summary file %q: %w", path, err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close summary file %q: %w", path, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to replace summary file %q: %w", path, err)
	}
	return nil
}

func formatLatencyNS(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return ""
	}
	return strconv.FormatInt(int64(v), 10)
}

func summaryFilename(id string) string {
	return fmt.Sprintf("summary-%s.json", id)
}

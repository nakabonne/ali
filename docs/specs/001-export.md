# Export results

## Goal

Add an export feature that persists load test results for downstream processing:
- Export all data points to `results.csv` in a stable schema.
- Export a summary to `summary-<id>.json`.

## Deliverables

- Add a new CLI flag: `--export-to <dir>`.
- When `--export-to <dir>` is provided,
  - Create `<dir>`
  - Export all generated data points as CSV to `<dir>/results.csv` at the end of the run.
  - Write the processed data to `<dir>/summary-<id>.json`.
- Add unit tests

## Acceptance Criteria

### Core behavior

1. Running the tool **without** `--export-to` produces **no CSV file** and behaves as before.
2. Running with `--export-to results.csv` creates
   - `results.csv` with schema as defined in the [Schema](#schema) section. 
   - `summary-<id>.json` with schema as defined in the [Schema](#schema) section. 
3. If `--export-to <path>` points to an existing file:
   - The command fails with a non-zero exit code before rendering TUI
   - It does not modify the file.
4. Export failures (permission denied, invalid path, disk full, etc.) produce:
   - non-zero exit code,
   - a clear error message including the path.

### Compatibility / UX

1. Whenever a new run is trigger by hitting the `<Enter>` key, append new data points with a different id to `results.csv`.

### Testing

1. Unit tests verify CSV serialization with deterministic fixtures.
2. `make test` passes.
3. `go build` builds a binary without any problem.

## Context

- This tool is a TUI tool that allows us to generate HTTP requests.
- After rendering the TUI, you can hit the Enter key to start generating HTTP requests.
- Users want to import results into external systems (TSDB, spreadsheets, BI tools). CSV is broadly supported.

## Schema

### results.csv

Columns:

| Column           | Type    | Required | Description |
|------------------|---------|----------|-------------|
| `id`             | string  | yes      | Unique identifier for the run (UUID) |
| `timestamp`      | string  | yes      | RFC3339 |
| `latency_ns`     | int     | yes      | Metric name (e.g., `latency_ms_p50`, `rps`, `errors_total`) |
| `url`            | string  | yes      | The target URL. |
| `method`         | string  | yes      | The HTTP method. (e.g. GET, POST) |
| `status_code`    | int     | yes      | HTTP status code |

CSV Formatting Rules:

- UTF-8 encoding.
- Header row included.

### summary-<id>.json

```json
{
  "target": {
    "url": "string",
    "method": "string"
  },
  "parameters": {
    "rate": "number",
    "duration_seconds": "number"
  },
  "timing": {
    "earliest": "RFC3339 string",
    "latest": "RFC3339 string"
  },
  "requests": {
    "count": "integer",
    "success_ratio": "number"
  },
  "throughput": "number",
  "latency_ms": {
    "total": "number",
    "mean": "number",
    "p50": "number",
    "p90": "number",
    "p95": "number",
    "p99": "number",
    "max": "number",
    "min": "number"
  },
  "bytes": {
    "in": { "total": "integer", "mean": "number" },
    "out": { "total": "integer", "mean": "number" }
  },
  "status_codes": {
    "200": "number"
  }
}
```

## Example

### Usage

```bash
ali --export-to ./results/
```

### Output file

#### ./results/results.csv

```csv
id,timestamp,latency_ns,url,method,status_code
4c3f2a5c-1d5b-4c0b-9d7f-8f0c0a3c6a2e,2026-01-15T13:45:12+09:00,18234567,https://nakabonne.dev/,GET,200
4c3f2a5c-1d5b-4c0b-9d7f-8f0c0a3c6a2e,2026-01-15T13:45:12.020+09:00,44900123,https://nakabonne.dev/,GET,200
4c3f2a5c-1d5b-4c0b-9d7f-8f0c0a3c6a2e,2026-01-15T13:45:12.041+09:00,935489752,https://nakabonne.dev/,GET,500
4c3f2a5c-1d5b-4c0b-9d7f-8f0c0a3c6a2e,2026-01-15T13:45:12.063+09:00,55320758,https://nakabonne.dev/,GET,200
4c3f2a5c-1d5b-4c0b-9d7f-8f0c0a3c6a2e,2026-01-15T13:45:12.084+09:00,965397887,https://nakabonne.dev/,GET,200
```

#### ./results/summary-<id>.json

```json
{
  "target": { "url": "https://nakabonne.dev/", "method": "GET" },
  "parameters": { "rate": 50, "duration_seconds": 2},
  "timing": {
    "earliest": "2021-03-13T15:20:43+09:00",
    "latest": "2021-03-13T15:20:45+09:00"
  },
  "requests": { "count": 100, "success_ratio": 1.0 },
  "throughput": 48.24,
  "latency_ms": {
    "total": 44000,
    "mean": 447.88,
    "p50": 445.46,
    "p90": 806.58,
    "p95": 849.89,
    "p99": 935.49,
    "max": 965.40,
    "min": 55.32
  },
  "bytes": {
    "in": { "total": 2325200, "mean": 23252 },
    "out": { "total": 0, "mean": 0 }
  },
  "status_codes": { "200": 100 }
}
```

## Non-Functional Requirements

1. **Atomic file writes**:
   - For file exports, write to a temp file in the same directory, then rename/replace.
   - Do not leave partial/corrupted final files on failure.

2. **Streaming output**:
   - Write CSV incrementally with buffering; do not require holding all rows in memory.

3. **Deterministic output**:
   - Ensure stable row ordering for testability.
   - Ensure deterministic `id` in tests (inject clock / fixed ID in fixtures).

4. **Backward compatibility**:
   - No change to existing behavior unless `--export-to` is provided.
   - Do not break existing CLI flags, exit codes, or TUI flows.

5. **Clear, actionable errors**:
   - Return non-zero exit codes on any export failure.

6. **Test coverage**:
   - Unit tests for CSV serialization include quoting/escaping, empty results, and NaN/Inf handling.

## Edge Cases

1. **Empty results** (e.g., run aborted immediately):
   - CSV still includes header row.
   - Data rows may be zero.

2. **Special characters in fields**:
   - Scenario/request names containing commas, quotes, or newlines are properly quoted/escaped (RFC4180 style).

3. **NaN/Inf values**:
   - Define and implement one of:
     - omit the row, or
     - serialize `value` as empty.
   - Tests must cover the chosen behavior.

4. **Very large runs**:
   - Export should stream to the writer; avoid building the entire CSV in memory.

5. **Permission/path errors**:
   - Permission denied, invalid directory, non-existent parent directory, read-only filesystem.
   - Must fail with a non-zero exit code and an error message that includes the path.

6. **Existing file behavior (no-clobber default)**:
   - Check if the path specified by `--export-to` exist on the startup.
   - If the target exists:
     - fail without modifying the existing file,

7. **Concurrent runs writing to the same path**:
   - One should fail due to file existence.

8. **Cross-platform paths**:
   - Handle paths with spaces and (if supported) Windows-style paths/backslashes.

9. **Interrupted export** (process terminated mid-write):
   - Atomic-write strategy should ensure no partially written final file is left behind.

## Future extensions

Support these format:
- JSON
- Influx line protocol.

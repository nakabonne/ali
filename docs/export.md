# Exporting results

Use `--export-to` to persist load test results for downstream processing.

## Usage

```bash
ali --export-to ./results/
```

## What gets written

When `--export-to <dir>` is provided, ali creates the directory (if needed) and writes:

- `<dir>/results.csv` with all data points for the run.
- `<dir>/summary-<id>.json` with an aggregated summary for the run.

If you start a new run by pressing `<Enter>` in the TUI, ali appends new rows with a
new run `id` to `results.csv` and writes a new `summary-<id>.json`.

If `--export-to <dir>` points to an existing file, the command fails before rendering
the TUI and the file is left unchanged.

## CSV schema: `results.csv`

Columns:

| Column        | Type   | Description |
|---------------|--------|-------------|
| `id`          | string | Unique identifier for the run (UUID). |
| `timestamp`   | string | RFC3339 timestamp. |
| `latency_ns`  | int    | Request latency in nanoseconds. |
| `url`         | string | Target URL. |
| `method`      | string | HTTP method (e.g., GET, POST). |
| `status_code` | int    | HTTP status code. |

## JSON schema: `summary-<id>.json`

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

## Example output

`./results/results.csv`:

```csv
id,timestamp,latency_ns,url,method,status_code
f48ff413-c446-4021-8a28-f153ee2e1151,2026-01-19T13:44:38.779088333+09:00,199035250,https://example.com/,GET,200
f48ff413-c446-4021-8a28-f153ee2e1151,2026-01-19T13:44:39.779554166+09:00,10721500,https://example.com/,GET,200
f48ff413-c446-4021-8a28-f153ee2e1151,2026-01-19T13:44:40.779522791+09:00,11019792,https://example.com/,GET,200
```

`./results/summary-<id>.json`:

```json
{
  "target": {
    "url": "https://example.com/",
    "method": "GET"
  },
  "parameters": {
    "rate": 1,
    "duration_seconds": 3
  },
  "timing": {
    "earliest": "2026-01-19T13:44:38.779088333+09:00",
    "latest": "2026-01-19T13:44:40.779522791+09:00"
  },
  "requests": {
    "count": 3,
    "success_ratio": 1
  },
  "throughput": 1.4914582322715022,
  "latency_ms": {
    "total": 220.776542,
    "mean": 73.59218,
    "p50": 11.019792,
    "p90": 199.03525,
    "p95": 199.03525,
    "p99": 199.03525,
    "max": 199.03525,
    "min": 10.7215
  },
  "bytes": {
    "in": {
      "total": 70137,
      "mean": 23379
    },
    "out": {
      "total": 0,
      "mean": 0
    }
  },
  "status_codes": {
    "200": 3
  }
}
```

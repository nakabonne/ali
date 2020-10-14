package main

import (
	"bytes"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nakabonne/ali/attacker"
)

func TestValidateMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{
			name:   "wrong method",
			method: "WRONG",
			want:   false,
		},
		{
			name:   "right method",
			method: "GET",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		want    *cli
		wantErr bool
	}{
		{
			name: "with default options",
			want: &cli{
				rate:         50,
				duration:     time.Second * 10,
				timeout:      time.Second * 30,
				method:       "GET",
				headers:      []string{},
				maxBody:      -1,
				noKeepAlive:  false,
				workers:      10,
				maxWorkers:   math.MaxUint64,
				connections:  10000,
				stdout:       new(bytes.Buffer),
				stderr:       new(bytes.Buffer),
				noHTTP2:      false,
				localAddress: "0.0.0.0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			got, err := parseFlags(b, b)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		cli      *cli
		args     []string
		wantCode int
	}{
		{
			name:     "print version",
			cli:      &cli{version: true},
			args:     []string{},
			wantCode: 0,
		},
		{
			name:     "no target given",
			cli:      &cli{},
			args:     []string{},
			wantCode: 1,
		},
		{
			name:     "bad URL",
			cli:      &cli{},
			args:     []string{"bad-url"},
			wantCode: 1,
		},
		{
			name:     "failed to make options",
			cli:      &cli{method: "WRONG"},
			args:     []string{"bad-url"},
			wantCode: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			tt.cli.stdout = b
			tt.cli.stderr = b
			got := tt.cli.run(tt.args)
			assert.Equal(t, tt.wantCode, got)
		})
	}
}

func TestMakeOptions(t *testing.T) {
	tests := []struct {
		name    string
		cli     *cli
		want    *attacker.Options
		wantErr bool
	}{
		{
			name:    "wrong method",
			cli:     &cli{method: "WRONG"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "duration less than 0",
			cli: &cli{
				method:   "GET",
				duration: -1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing colon in given header",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"keyvalue"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing key in given header",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{":value"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing value in given header",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "both body and body file given",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:value"},
				body:     "body",
				bodyFile: "path/to",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "body given",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:value"},
				body:     `{"foo": 1}`,
			},
			want: &attacker.Options{
				Rate:     0,
				Duration: 1,
				Timeout:  0,
				Method:   "GET",
				Body:     []byte(`{"foo": 1}`),
				Header: http.Header{
					"key": []string{"value"},
				},
				Workers:    0,
				MaxWorkers: 0,
				MaxBody:    0,
				HTTP2:      true,
				KeepAlive:  true,
				Buckets:    []time.Duration{},
			},
			wantErr: false,
		},
		{
			name: "body file given",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:value"},
				body:     "",
				bodyFile: "testdata/body-1.json",
			},
			want: &attacker.Options{
				Rate:     0,
				Duration: 1,
				Timeout:  0,
				Method:   "GET",
				Body:     []byte(`{"foo": 1}`),
				Header: http.Header{
					"key": []string{"value"},
				},
				Workers:    0,
				MaxWorkers: 0,
				MaxBody:    0,
				HTTP2:      true,
				KeepAlive:  true,
				Buckets:    []time.Duration{},
			},
			wantErr: false,
		},
		{
			name: "wrong body file given",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:value"},
				bodyFile: "wrong",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "disable http2",
			cli: &cli{
				method:   "GET",
				duration: 1,
				headers:  []string{"key:value"},
				body:     "",
				bodyFile: "testdata/body-1.json",
				noHTTP2:  true,
			},
			want: &attacker.Options{
				Rate:     0,
				Duration: 1,
				Timeout:  0,
				Method:   "GET",
				Body:     []byte(`{"foo": 1}`),
				Header: http.Header{
					"key": []string{"value"},
				},
				Workers:    0,
				MaxWorkers: 0,
				MaxBody:    0,
				HTTP2:      false,
				KeepAlive:  true,
				Buckets:    []time.Duration{},
			},
			wantErr: false,
		},
		{
			name: "disable keepalive",
			cli: &cli{
				method:      "GET",
				duration:    1,
				headers:     []string{"key:value"},
				body:        "",
				bodyFile:    "testdata/body-1.json",
				noKeepAlive: true,
			},
			want: &attacker.Options{
				Rate:     0,
				Duration: 1,
				Timeout:  0,
				Method:   "GET",
				Body:     []byte(`{"foo": 1}`),
				Header: http.Header{
					"key": []string{"value"},
				},
				Workers:    0,
				MaxWorkers: 0,
				MaxBody:    0,
				HTTP2:      true,
				KeepAlive:  false,
				Buckets:    []time.Duration{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cli.makeOptions()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestSetDebug(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{
			name:  "in non-debug use",
			debug: false,
		},
		{
			name:  "in debug use",
			debug: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDebug(nil, tt.debug)
		})
	}
}

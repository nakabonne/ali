package main

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/nakabonne/ali/attacker"

	"github.com/stretchr/testify/assert"
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
			name: "missing colon in given header",
			cli: &cli{
				method:  "GET",
				headers: []string{"keyvalue"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing key in given header",
			cli: &cli{
				method:  "GET",
				headers: []string{":value"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing value in given header",
			cli: &cli{
				method:  "GET",
				headers: []string{"key:"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "both body and body file given",
			cli: &cli{
				method:   "GET",
				headers:  []string{"key:value"},
				body:     "body",
				bodyFile: "path/to",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "body file given",
			cli: &cli{
				method:   "GET",
				headers:  []string{"key:value"},
				body:     "",
				bodyFile: "testdata/body-1.json",
			},
			want: &attacker.Options{
				Rate:     0,
				Duration: 0,
				Timeout:  0,
				Method:   "GET",
				Body:     []byte(`{"foo": 1}`),
				Header: http.Header{
					"key": []string{"value"},
				},
				Workers:    0,
				MaxWorkers: 0,
				MaxBody:    0,
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

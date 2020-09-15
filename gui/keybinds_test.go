package gui

import (
	"testing"

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

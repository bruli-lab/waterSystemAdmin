package web

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bruli-lab/waterSystemAdmin/internal/api"
)

func TestParseSeconds(t *testing.T) {
	tests := []struct {
		value   string
		want    int
		wantErr bool
	}{
		{value: "1", want: 1},
		{value: "120", want: 120},
		{value: "0", wantErr: true},
		{value: "-5", wantErr: true},
		{value: "121", wantErr: true},
		{value: "abc", wantErr: true},
		{value: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseSeconds(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSeconds(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseSeconds(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestErrMessage(t *testing.T) {
	down := fmt.Errorf("%w: request GET /zones: connection refused", api.ErrUnavailable)
	if got := errMessage("Could not read the zones", down); got != "Could not read the zones: the irrigation API is not reachable." {
		t.Errorf("unexpected message for unavailable API: %q", got)
	}
	if got := errMessage("Could not read the zones", errors.New("status code 500")); got != "Could not read the zones: status code 500" {
		t.Errorf("unexpected message: %q", got)
	}
}

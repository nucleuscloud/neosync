package gotypeparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseStringAsNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"Valid int", "123", int64(123), false},
		{"Valid float", "123.45", float64(123.45), false},
		{"Invalid number", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStringAsNumber(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

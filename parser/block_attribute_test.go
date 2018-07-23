package parser

import (
	"testing"
)

func TestIsBlockAttribute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		data []byte
		ok   bool
	}{
		{
			data: []byte("{...}"),
			ok:   true,
		},
		{
			data: []byte("{... "),
		},
	}

	for _, test := range tests {
		got := isBlockAttribute(test.data)
		if got != test.ok {
			t.Errorf("got %t, want %t", got, test.ok)
		}
	}
}

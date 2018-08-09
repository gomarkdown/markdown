package parser

import (
	"strconv"
	"testing"
)

func TestIsCallout(t *testing.T) {
	tests := []struct {
		data   []byte
		number int
	}{
		// ok
		{[]byte("<<4>>"), 4},
		// fail
		{[]byte("<<5a5>>"), 0},
	}

	for i, test := range tests {
		bnum, read := IsCallout(test.data)
		num, _ := strconv.Atoi(string(bnum))

		if num != test.number {
			t.Errorf("test %d, want %d, got %d", i, test.number, num)
		}
		if num > 0 && read != len(test.data) {
			t.Errorf("test %d, want %d, got %d", i, len(test.data), read)
		}
	}
}

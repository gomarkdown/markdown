package parser

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestCrossReference(t *testing.T) {
	p := New()

	tests := []struct {
		data []byte
		r    *ast.CrossReference
		fail bool
	}{
		// ok
		{
			data: []byte("(#yes)"),
			r:    &ast.CrossReference{Destination: []byte("yes")},
		},
		// ok
		{
			data: []byte("(#y:es)"),
			r:    &ast.CrossReference{Destination: []byte("y:es")},
		},
		// fails
		{data: []byte("(#y es)"), r: nil, fail: true},
		{data: []byte("(#yes"), r: nil, fail: true},
	}

	for i, test := range tests {
		_, n := maybeShortRefOrIndex(p, test.data, 0)
		if test.fail && n != nil {
			t.Errorf("test %d, should have failed to parse %s", i, test.data)
		}
	}
}

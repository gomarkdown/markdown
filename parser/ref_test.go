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
		{
			data: []byte("(#y:es)"),
			r:    &ast.CrossReference{Destination: []byte("y:es")},
		},
		{
			data: []byte("(#id, random text)"),
			r:    &ast.CrossReference{Destination: []byte("id"), Suffix: []byte("random text")},
		},
		// fails
		{data: []byte("(#y es)"), r: nil, fail: true},
		{data: []byte("(#yes"), r: nil, fail: true},
		{data: []byte("(#y es, random text"), r: nil, fail: true},
	}

	for i, test := range tests {
		_, n := maybeShortRefOrIndex(p, test.data, 0)
		if test.fail && n != nil {
			t.Errorf("test %d, should have failed to parse %s", i, test.data)
		}
	}
}

func TestIndex(t *testing.T) {
	p := New()

	tests := []struct {
		data []byte
		i    *ast.Index
		fail bool
	}{
		// ok
		{
			data: []byte("(!yes)"),
			i:    &ast.Index{Item: []byte("yes")},
		},
		{
			data: []byte("(!y:es)"),
			i:    &ast.Index{Item: []byte("y:es")},
		},
		{
			data: []byte("(!yes, no)"),
			i:    &ast.Index{Item: []byte("yes"), Subitem: []byte("no")},
		},
		{
			data: []byte("(!  yes  , no  )"),
			i:    &ast.Index{Item: []byte("yes"), Subitem: []byte("no")},
		},
		// fails
		{data: []byte("(!yes"), fail: true},
		{data: []byte("(_yes"), fail: true},
	}

	for i, test := range tests {
		_, n := maybeShortRefOrIndex(p, test.data, 0)
		if test.fail && n != nil {
			t.Errorf("test %d, should have failed to parse %s", i, test.data)
			continue
		}
		if test.fail && n == nil {
			// ok
			continue
		}

		idx := n.(*ast.Index)

		if string(test.i.Item) != string(idx.Item) {
			t.Errorf("test %d, got item %s, wanted %s", i, idx.Item, test.i.Item)
		}
		if string(test.i.Subitem) != string(idx.Subitem) {
			t.Errorf("test %d, got item %s, wanted %s", i, idx.Subitem, test.i.Subitem)
		}
		if test.i.Primary != idx.Primary {
			t.Errorf("test %d, got item %t, wanted %t", i, idx.Primary, test.i.Primary)
		}
	}
}

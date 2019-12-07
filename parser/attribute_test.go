package parser

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestBlockAttribute(t *testing.T) {
	p := NewWithExtensions(CommonExtensions | Attributes)
	tests := []struct {
		data []byte
		attr *ast.Attribute
		left int
	}{
		// ok
		{
			data: []byte("{#myid}"),
			attr: &ast.Attribute{ID: []byte("myid")},
		},
		{
			data: []byte("{#myid #myid2}"),
			attr: &ast.Attribute{ID: []byte("myid2")},
		},
		{
			data: []byte("{.myclass}"),
			attr: &ast.Attribute{
				Classes: [][]byte{[]byte("myclass")},
			},
		},
		{
			data: []byte("{#myid .myclass}"),
			attr: &ast.Attribute{
				ID:      []byte("myid"),
				Classes: [][]byte{[]byte("myclass")},
			},
		},
		{
			data: []byte("{#myid .myclass .myclass2}"),
			attr: &ast.Attribute{
				ID:      []byte("myid"),
				Classes: [][]byte{[]byte("myclass")},
			},
		},
		{
			data: []byte(`{key="value"}`),
			attr: &ast.Attribute{
				Attrs: map[string][]byte{"key": []byte("value")},
			},
		},
		{
			data: []byte(`{key="value" #myid .myclass}`),
			attr: &ast.Attribute{
				ID:      []byte("myid"),
				Classes: [][]byte{[]byte("myclass")},
				Attrs:   map[string][]byte{"key": []byte("value")},
			},
		},
		{
			data: []byte(`{key="value" key2="value2" #myid #myid2 .myclass .myclass2}` + "\nmore"),
			attr: &ast.Attribute{
				ID:      []byte("myid2"),
				Classes: [][]byte{[]byte("myclass"), []byte("myclass2")},
				Attrs:   map[string][]byte{"key": []byte("value"), "key2": []byte("value2")},
			},
			left: 5,
		},
		// fails

		// missing quote
		{data: []byte(`{key=value" #myid .myclass}`)},
		// too many spaces (should allow this eventually)
		{data: []byte(`{key ="value"}`)},
		// not block attribute
		{data: []byte("hello")},
	}
	for i, test := range tests {
		p.attr = nil
		data := p.attribute(test.data)
		if x := len(data); test.left > 0 && x != test.left {
			t.Errorf("test %d, want %d of left data got %d", i, test.left, x)
		}

		if p.attr == nil && test.attr != nil {
			t.Errorf("test %d, got nil for attribute", i)
			continue
		}
		if p.attr == nil && test.attr == nil {
			// ok
			continue
		}

		if !bytes.Equal(test.attr.ID, p.attr.ID) {
			t.Errorf("test %d, got %q for ID, want %q", i, p.attr.ID, test.attr.ID)
		}
		for i, c := range test.attr.Classes {
			if !bytes.Equal(c, p.attr.Classes[i]) {
				t.Errorf("test %d, got %q for class, want %q", i, p.attr.Classes[i], c)
			}
		}
		if test.attr.Attrs != nil {
			if !reflect.DeepEqual(test.attr.Attrs, p.attr.Attrs) {
				t.Errorf("test %d, got %q for class, want %q", i, test.attr.Attrs, p.attr.Attrs)
			}
		}
	}
}

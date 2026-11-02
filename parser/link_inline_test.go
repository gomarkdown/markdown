package parser

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// <> around a destination are removed only as a pair, and an escaped > is
// part of the destination.
func TestInlineLinkAngleBrackets(t *testing.T) {
	for _, tt := range []struct{ input, dest string }{
		{"[x](<a>)", "a"},
		{"[x](<a b>)", "a b"},
		{"[x](/a>)", "/a>"},
		{`[x](/a\>)`, "/a>"},
		{"[x](<a)", "<a"},
		{`[x](<a\>)`, "<a>"},
	} {
		var dest string
		ast.WalkFunc(New().Parse([]byte(tt.input)), func(node ast.Node, entering bool) ast.WalkStatus {
			if link, ok := node.(*ast.Link); ok && entering {
				dest = string(link.Destination)
			}
			return ast.GoToNext
		})
		if dest != tt.dest {
			t.Errorf("%q: got destination %q, want %q", tt.input, dest, tt.dest)
		}
	}
}

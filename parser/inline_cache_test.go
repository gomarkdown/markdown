package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// TestInlineCachesResetPerBuffer reuses one Parser on a buffer rewritten in
// place. The memo tables key on the buffer's address, so without a reset
// the second parse would see the first buffer's code span.
func TestInlineCachesResetPerBuffer(t *testing.T) {
	p := New()
	buf := []byte("`a` b")

	hasCode := func(n ast.Node) bool {
		for _, c := range n.GetChildren() {
			if _, ok := c.(*ast.Code); ok {
				return true
			}
		}
		return false
	}

	first := &ast.Paragraph{}
	p.Inline(first, buf)
	if !hasCode(first) {
		t.Fatalf("first parse: want a code span")
	}

	copy(buf, "` a b")
	second := &ast.Paragraph{}
	p.Inline(second, buf)
	if hasCode(second) {
		t.Fatalf("second parse: stale code span from the first buffer")
	}
	var got bytes.Buffer
	for _, c := range second.Children {
		got.Write(c.AsLeaf().Literal)
	}
	if got.String() != "` a b" {
		t.Fatalf("second parse: got %q, want %q", got.String(), "` a b")
	}
}

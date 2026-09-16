package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// Adversarial inputs for the inline callbacks that fire once per byte of a
// run. Each used to rescan from the cursor on every firing.

func BenchmarkInlineBareURLs(b *testing.B) {
	input := []byte(strings.Repeat("http://x ", 10000) + "\n")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := NewWithExtensions(Autolink)
		p.Parse(input)
	}
}

func BenchmarkInlineSpaceRun(b *testing.B) {
	input := []byte("a" + strings.Repeat(" ", 50000) + "b\n")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := New()
		p.Parse(input)
	}
}

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

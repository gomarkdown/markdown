package parser

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestFencedCodeBlockFence(t *testing.T) {
	cases := []struct {
		name   string
		exts   Extensions
		flags  Flags
		input  string
		char   byte
		length int
		offset int
	}{
		{"backticks", CommonExtensions, 0, "```\ncode\n```\n", '`', 3, 0},
		{"long backticks", CommonExtensions, 0, "````md\n```\ncode\n```\n````\n", '`', 4, 0},
		{"tildes", CommonExtensions, 0, "~~~~~\ncode\n~~~~~\n", '~', 5, 0},
		{"offset", CommonExtensions, 0, "  ~~~\ncode\n  ~~~\n", '~', 3, 2},
		{"commonmark backticks", CommonExtensions, CommonMark, "````md\ncode\n````\n", '`', 4, 0},
		{"commonmark tildes", CommonExtensions, CommonMark, " ~~~~\ncode\n ~~~~\n", '~', 4, 1},
		{"commonmark offset 3", CommonExtensions, CommonMark, "   ```\ncode\n```\n", '`', 3, 3},
		{"inline code span", NoExtensions, 0, "```go\ncode\n```\n", '`', 3, 0},
		{"indented", CommonExtensions, 0, "    code\n", 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := NewWithExtensions(c.exts)
			p.Opts.Flags = c.flags
			doc := p.Parse([]byte(c.input))
			// find first code block
			var cb *ast.CodeBlock
			ast.WalkFunc(doc, func(n ast.Node, entering bool) ast.WalkStatus {
				if n, ok := n.(*ast.CodeBlock); ok && cb == nil {
					cb = n
				}
				return ast.GoToNext
			})
			if cb == nil {
				t.Fatal("no code block found")
			}
			if cb.FenceChar != c.char || cb.FenceLength != c.length || cb.FenceOffset != c.offset {
				t.Errorf("got char=%q length=%d offset=%d, want char=%q length=%d offset=%d",
					cb.FenceChar, cb.FenceLength, cb.FenceOffset, c.char, c.length, c.offset)
			}
		})
	}
}

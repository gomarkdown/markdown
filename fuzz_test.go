package markdown

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// FuzzParse parses and renders input with the historical extensions and with
// CommonMark. go test checks the seeds. go test -fuzz=FuzzParse searches further.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"",
		"# Intro\n\nParagraph with *em* and `code`.\n",
		"- a\n- b\n\n1. c\n",
		"[link](http://example.com \"title\")\n\n![alt](/img.png)\n",
		"```go\nfunc f() {}\n```\n",
		"| a | b |\n| --- | --- |\n| 1 | 2 |\n",
		"~~strike~~ and http://example.com/a_(b).\n",
		"> quote\n>\n> second\n",
		"foo  \nbar\n",
		"text[^1]\n\n[^1]: note\n",
		"{#id .cls}\n# Heading\n",
		"{{include.md}}\n",
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		fuzzParsed(t, input, parser.NewWithExtensions(parser.CommonExtensions), html.CommonFlags)
		fuzzParsed(t, input, parser.NewCommonMark(), html.UseXHTML)
	})
}

// fuzzParsed parses input, checks the tree, and renders it. A panic fails the test.
func fuzzParsed(t *testing.T, input []byte, p *parser.Parser, flags html.Flags) {
	t.Helper()
	original := bytes.Clone(input)
	doc := p.Parse(input)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}
		for i, child := range node.GetChildren() {
			if child.GetParent() != node {
				t.Fatal("incorrect AST parent")
			}
			if i > 0 && ast.GetPrevNode(child) != node.GetChildren()[i-1] {
				t.Fatal("incorrect AST previous sibling")
			}
			if i+1 < len(node.GetChildren()) && ast.GetNextNode(child) != node.GetChildren()[i+1] {
				t.Fatal("incorrect AST next sibling")
			}
		}
		return ast.GoToNext
	})
	Render(doc, html.NewRenderer(html.RendererOptions{Flags: flags}))
	if !bytes.Equal(input, original) {
		t.Fatal("parser modified the input")
	}
}

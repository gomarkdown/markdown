package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func astToString(node ast.Node) string {
	var buf bytes.Buffer
	ast.Print(&buf, node)
	return buf.String()
}

// Inside HTML, no Markdown is parsed.
func TestHtmlP(t *testing.T) {
	input := "<p>*not emph*</p>\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	got := astToString(doc)
	exp := "HTMLBlock '<p>*not emph*</p>'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

// Inside SVG, so the RECT element is passed through.
func TestSVG(t *testing.T) {
	input := "<svg><rect> *no emph* </rect></svg>\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	got := astToString(doc)
	exp := "HTMLBlock '<svg><rect> *no emph* </rect></svg>'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

/*
When we had code block:

```go main.go
package main
```

Originally it was:      CodeBlock.Info = "go"         (up to space)
For #320 changed it to: CodeBlock.Info = "go main.go" (up to newline)
*/
func TestCodeBlocIssue320(t *testing.T) {
	input := "```go main.go,readonly\npackage main\n```"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	got := astToString(doc)
	exp := "CodeBlock:go main.go,readonly 'package main\\n'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

// The RECT element on its own is nothing special, so Markdown is parsed.
func TestRect(t *testing.T) {
	input := "<rect> *emph* </rect>\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	got := astToString(doc)
	exp := "Paragraph\n  Text\n  HTMLSpan '<rect>'\n  Text\n  Emph\n    Text 'emph'\n  Text\n  HTMLSpan '</rect>'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

// https://github.com/gomarkdown/markdown/issues/326
func TestInfiniteLoopFix(t *testing.T) {
	input := "```\n: la"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	if doc == nil {
		t.Errorf("Expected non-nil AST")
	}
	//ast.Print(os.Stdout, doc)
}

// https://github.com/gomarkdown/markdown/issues/325: a fence closed inside a
// quote or aside must not pair with a fence after it.
func TestBug325(t *testing.T) {
	for _, prefix := range []string{">", "A>"} {
		input := prefix + " ~~~\n" + prefix + " // comment\n" + prefix + " ~~~\n\n~~~"
		p := NewWithExtensions(CommonExtensions | Mmark)
		got := astToString(p.Parse([]byte(input)))
		block := "BlockQuote"
		if prefix == "A>" {
			block = "Aside"
		}
		exp := block + "\n  CodeBlock: '// comment\\n'\nParagraph\n  Text '~~~'\n"
		if got != exp {
			t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
				input, exp, got)
		}
	}
}

// Lines of a fenced code block in a quote or aside may omit the prefix, but
// an unclosed fence is not code and must not keep the rest of the document in
// the block.
func TestPrefixedBlockFences(t *testing.T) {
	for _, prefix := range []string{">", "A>"} {
		block := "BlockQuote"
		if prefix == "A>" {
			block = "Aside"
		}
		for _, test := range []struct{ input, exp string }{
			{
				prefix + " ~~~\n" + prefix + " code\n\npara\n",
				block + "\n  Paragraph\n    Text '~~~\\ncode'\nParagraph\n  Text 'para'\n",
			},
			{
				prefix + " ~~~\ncode\n" + prefix + " ~~~\n\npara\n",
				block + "\n  CodeBlock: 'code\\n'\nParagraph\n  Text 'para'\n",
			},
			{
				prefix + " ~~~\n" + prefix + " code\n~~~\n\npara\n",
				block + "\n  CodeBlock: 'code\\n'\nParagraph\n  Text 'para'\n",
			},
		} {
			p := NewWithExtensions(CommonExtensions | Mmark)
			if got := astToString(p.Parse([]byte(test.input))); got != test.exp {
				t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n", test.input, test.exp, got)
			}
		}
	}
}

// A run of quotes with unclosed fences must not make each one scan the rest
// of the document.
func BenchmarkQuoteUnclosedFences(b *testing.B) {
	input := bytes.Repeat([]byte("> ~~~\n\n"), 20000)
	for i := 0; i < b.N; i++ {
		NewWithExtensions(CommonExtensions).Parse(input)
	}
}

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

// TODO: fix parsing for this test case without breaking other test cases
// https://github.com/gomarkdown/markdown/issues/325
// this can be fixed by removing p.fenceCodeBlock() call in Parser.quote()
// but would break other test cases
func TestBug325(t *testing.T) {
	input := `> ~~~
> // comment
> ~~~

~~~`
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	got := astToString(doc)
	// TODO: currently it sees the last `~~~` as a code block end
	// should see `> ~~~` as a code block end
	exp := "BlockQuote\n  CodeBlock: '> // comment\\n> ~~~\\n\\n'\n"
	//exp := "BlockQuote\n  CodeBlock: '// comment\\n'\nParagraph\n  Text '~~~'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

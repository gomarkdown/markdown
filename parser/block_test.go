package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// Inside HTML, no Markdown is parsed.
func TestHtmlP(t *testing.T) {
	input := "<p>*not emph*</p>\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
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
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "HTMLBlock '<svg><rect> *no emph* </rect></svg>'\n"
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
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "Paragraph\n  Text\n  HTMLSpan '<rect>'\n  Text\n  Emph\n    Text 'emph'\n  Text\n  HTMLSpan '</rect>'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

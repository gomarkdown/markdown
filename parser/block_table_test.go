package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestBug195(t *testing.T) {
	input := "| a | b |\n| - | - |\n|`foo|bar` | types |\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	// TODO: change expectations for https://github.com/gomarkdown/markdown/issues/195
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text\n        Code 'foo|bar'\n      TableCell\n        Text 'types'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

func TestBug198(t *testing.T) {
	// there's a space after end of table header, which used to break table parsing
	input := `| a | b|
| :--- | ---: |
| c | d |`
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text 'c'\n      TableCell\n        Text 'd'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

// https://github.com/gomarkdown/markdown/issues/274
func TestIssue274(t *testing.T) {
	input := "| a | b |\n| - | - |\n|	foo | bar |\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text '\\tfoo'\n      TableCell\n        Text 'bar'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

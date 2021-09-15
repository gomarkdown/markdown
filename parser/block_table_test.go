package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestBug195(t *testing.T) {
	p := NewWithExtensions(CommonExtensions)
	input := "| a | b |\n| - | - |\n|`foo|bar` | types |\n"
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	// TODO: change expectations for https://github.com/gomarkdown/markdown/issues/195
	exp := "Paragraph\n  Text '| a | b |\\n| - | - |\\n|'\n  Code 'foo|bar'\n  Text '| types |'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

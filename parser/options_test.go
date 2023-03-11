package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

type CustomNode struct {
	ast.Container
}

// we default to true so to test it, we need to test false
func (n *CustomNode) CanContain(ast.Node) bool {
	return false
}

type CustomNode2 struct {
	ast.Container
}

func blockTitleHook(data []byte) (ast.Node, []byte, int) {
	sep := []byte(`%%%`)
	// parse text between %%% and %%% and return it as a blockQuote.
	if !bytes.HasPrefix(data, sep) {
		return nil, data, 0
	}

	end := bytes.Index(data[3:], sep)
	if end < 0 {
		return nil, data, 0
	}
	end += 3
	node := &CustomNode{}
	return node, data[4:end], end + 3
}

func blockTitleHook2(data []byte) (ast.Node, []byte, int) {
	sep := []byte(`%%%`)
	// parse text between %%% and %%% and return it as a blockQuote.
	if !bytes.HasPrefix(data, sep) {
		return nil, data, 0
	}

	end := bytes.Index(data[3:], sep)
	if end < 0 {
		return nil, data, 0
	}
	end += 3
	node := &CustomNode2{}
	return node, data[4:end], end + 3
}

func astPrint(root ast.Node) string {
	buf := &bytes.Buffer{}
	ast.Print(buf, root)
	return buf.String()
}

func astPretty(s string) string {
	s = strings.Replace(s, " ", "_", -1)
	s = strings.Replace(s, "\t", "__", -1)
	s = strings.Replace(s, "\n", "+", -1)
	return s
}

func TestCustomNode(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{
			data: `
%%%
hallo
%%%
`,
			want: `CustomNode
Paragraph 'hallo'
`,
		},
	}

	p := New()
	p.Opts = Options{ParserHook: blockTitleHook}

	for _, test := range tests {
		p.Block([]byte(test.data))
		data := astPrint(p.Doc)
		got := astPretty(data)
		want := astPretty(test.want)

		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	}
}

func TestCustomNode2(t *testing.T) {
	tests := []struct {
		data string
		want string
	}{
		{
			data: `
%%%
hallo
%%%
`,
			want: `CustomNode2
  Paragraph 'hallo'
`,
		},
	}

	p := New()
	p.Opts = Options{ParserHook: blockTitleHook2}

	for _, test := range tests {
		p.Block([]byte(test.data))
		data := astPrint(p.Doc)
		got := astPretty(data)
		want := astPretty(test.want)

		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	}
}

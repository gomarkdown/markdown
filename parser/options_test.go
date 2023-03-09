package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

type CustomNode struct {
	ast.Container
}

func (n *CustomNode) CanContain(ast.Node) bool {
	return true
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

func TestOptions(t *testing.T) {
	tests := []struct {
		data []byte
		want []byte
	}{
		{
			data: []byte(`
%%%
hallo
%%%
`),
			want: []byte(`CustomNode
  Paragraph 'hallo'
`),
		},
	}

	p := New()
	p.Opts = Options{ParserHook: blockTitleHook}
	buf := &bytes.Buffer{}

	for _, test := range tests {
		p.Block(test.data)
		ast.Print(buf, p.Doc)
		data := buf.Bytes()
		data = bytes.Replace(data, []byte(" "), []byte("_"), -1)
		test.want = bytes.Replace(test.want, []byte(" "), []byte("_"), -1)

		if !bytes.Equal(data, test.want) {
			t.Errorf("want ast %s, got %s", test.want, data)
		}
	}
}

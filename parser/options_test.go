package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func blockTitleHook(data []byte) (ast.Node, []byte, int) {
	// parse text between %%% and %%% and return it as a blockQuote.
	i := 0
	if len(data) < 3 {
		return nil, data, 0
	}
	if data[i] != '%' && data[i+1] != '%' && data[i+2] != '%' {
		return nil, data, 0
	}

	i += 3
	// search for end.
	for i < len(data) {
		if data[i] == '%' && data[i+1] == '%' && data[i+2] == '%' {
			break
		}
		i++
	}
	node := &ast.BlockQuote{}
	return node, data[4:i], i + 3
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
			want: []byte(`BlockQuote
  Paragraph 'hallo'
`),
		},
	}

	p := New()
	p.Opts = Options{ParserHook: blockTitleHook}
	buf := &bytes.Buffer{}

	for _, test := range tests {
		p.block(test.data)
		ast.Print(buf, p.Doc)
		data := buf.Bytes()
		data = bytes.Replace(data, []byte(" "), []byte("_"), -1)
		test.want = bytes.Replace(test.want, []byte(" "), []byte("_"), -1)

		if bytes.Compare(data, test.want) != 0 {
			t.Errorf("want ast %s, got %s", test.want, data)
		}
	}
}

package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func blockTitleHook(p *Parser, data []byte) int {
	// parse text between %%% and %%% and return it as a blockQuote.
	i := 0
	if len(data) < 3 {
		return 0
	}
	if data[i] != '%' && data[i+1] != '%' && data[i+2] != '%' {
		return 0
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
	p.addBlock(node)
	p.block(data[4:i])
	p.finalize(node)

	return i + 3
}

func TestOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		data    []byte
		wantAst ast.Node
	}{
		{
			data: []byte(`
%%%
hallo
%%%
`),
		},
	}

	p := New()
	p.Opts = ParserOptions{ParserHook: blockTitleHook}

	for _, test := range tests {
		p.block(test.data)
		ast.Print(os.Stdout, p.Doc)
		fmt.Print("\n")
	}
}

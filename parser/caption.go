package parser

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

func (p *Parser) caption(startwith, data []byte) (ast.Node, int) {
	if !bytes.HasPrefix(data, startwith) {
		return nil, 0
	}
	j := len(startwith)
	data = data[j:]
	end := p.linesUntilEmpty(data)

	node := &ast.Caption{}
	p.inline(node, data[:end])

	return node, end + j
}

// linesUntilEmpty scans lines up to the first empty line.
func (p *Parser) linesUntilEmpty(data []byte) int {
	line, i := 0, 0

	for line < len(data) {
		i++

		// find the end of this line
		for i < len(data) && data[i-1] != '\n' {
			i++
		}

		if p.isEmpty(data[line:i]) == 0 {
			line = i
			continue
		}

		break
	}
	return i
}

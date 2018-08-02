package parser

import "github.com/gomarkdown/markdown/ast"

// parse '(#r)', where r does not contain spaces and is an existing label. Or.
// (!item) (!item, subitem), for an index, (!!item) signals primary.
func maybeShortRefOrIndex(p *Parser, date []byte, offset int) (int, ast.Node) {
	if len(data[offset:]) < 4 {
		return 0
	}
	// short ref first
	data = data[offset:]
	i := 1
	if data[i] != '#' {
		return 0
	}
	i++
	for i < len(data) && data[i] != ')' {
		if isAlnum(data[i]) {
			i++
			continue
		}
		if data[i] == '_' || data[i] == '-' || data[i] == ':' {
			i++
			continue
		}
		i = 0
		break
	}
	// not found, or not valid
	if i == 0 {
		return 0, nil
	}

	id := data[2:i]
	lr, ok := p.getRef(string(id))
	if !ok {
		return 0, nil
	}
}

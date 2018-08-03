package parser

import "github.com/gomarkdown/markdown/ast"

// parse '(#r)', where r does not contain spaces and is an existing label. Or.
// (!item) (!item, subitem), for an index, (!!item) signals primary.
func maybeShortRefOrIndex(p *Parser, data []byte, offset int) (int, ast.Node) {
	if len(data[offset:]) < 4 {
		return 0, nil
	}
	// short ref first
	data = data[offset:]
	i := 1
	if data[i] != '#' {
		return 0, nil
	}
	i++
Loop:
	for i < len(data) {
		c := data[i]
		switch {
		case c == ')':
			break Loop
		case !isAlnum(c):
			if c == '_' || c == '-' || c == ':' {
				i++
				continue
			}
			i = 0
			break Loop
		}
		i++
	}
	// end not found or no valid syntax
	if i == 0 || data[i-1] != ')' {
		return 0, nil
	}

	id := data[2:i]
	node := &ast.CrossReference{}
	node.Destination = id

	return i, node
}

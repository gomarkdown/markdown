package parser

import "github.com/gomarkdown/markdown/ast"

func isBlockAttribute(data []byte) bool {
	i := 0

	// skip up to three spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// first character must be brace
	if data[i] != '{' {
		return false
	}

	for i < len(data) && data[i] != '\n' {
		i++
	}
	return data[i-1] == '}' // closing brace at end of line
}

// blockAttribute parses a block attribute, see ast.BlockAttribute for the syntax.
func (p *Parser) blockAttribute(data []byte) []byte {
	// called after isBlockAttribute, so we now we start and end with {
	/*
		if p.isHRule(data) {
			p.addBlock(&ast.HorizontalRule{})
			i := skipUntilChar(data, 0, '\n')
			data = data[i:]
			continue
		}
	*/
	// TODO: this should be a parent for all subsequent children.
	b := &ast.BlockAttribute{Attrs: make(map[string]string)}
	b.ID = "miek"
	p.addBlock(b)

	i := skipUntilChar(data, 0, '\n')
	return data[i:]
}

/*
func (p *parser) isBlockAttribute(data []byte) int {
	esc := false
	quote := false
	ialB := 0
	ial := NewBlockAttr()
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ':
			if quote {
				continue
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				ialB = i
				continue
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				} else {
					// this is illegal in an IAL, discard the posibility
					return 0
				}
			}
			ialB = i
		case '"':
			if esc {
				esc = !esc
				continue
			}
			quote = !quote
		case '\\':
			esc = !esc
		case '}':
			if esc {
				esc = !esc
				continue
			}
			chunk := data[ialB+1 : i]
			if len(chunk) == 0 {
				return i + 1
			}
			switch {
			case chunk[0] == '.':
				ial.class[string(chunk[1:])] = true
			case chunk[0] == '#':
				ial.id = string(chunk[1:])
			default:
				k, v := parseKeyValue(chunk)
				if k != "" {
					ial.attr[k] = v
				} else {
					// this is illegal in an IAL, discard the posibility
					return 0
				}
			}
			p.ial = p.ial.add(ial)
			return i + 1
		default:
			esc = false
		}
	}
	return 0
}
*/

package parser

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

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

	b := &ast.BlockAttribute{Attrs: make(map[string][]byte)}

	esc := false
	quote := false
	trail := 0
	i := 0
	for i = 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '}': // space seperated.
			if esc {
				esc = false
				continue
			}
			if quote {
				continue
			}
			chunk := data[trail:i]
			if len(chunk) == 0 {
				trail = i
				continue
			}
			switch {
			case chunk[0] == '.':
				b.Classes = append(b.Classes, chunk[1:])
			case chunk[0] == '#':
				b.ID = chunk[1:]
			default:
				k, v := keyValue(chunk)
				if k != nil && v != nil {
					b.Attrs[string(k)] = v
					continue
				}
				// illegal, not an attribute return.
				println("needs test")
				return data
			}
			trail = i + 1
		case '"':
			if esc {
				esc = false
				continue
			}
			quote = !quote
		case '\\':
			esc = !esc
		default:
			esc = false
		}
	}
	i = skipUntilChar(data[i:], 0, '\n')

	p.addBlock(b)

	println("RET", string(data[i:]), "RET")
	return data[i:]
}

// key="value" quotes are mandatory.
func keyValue(data []byte) ([]byte, []byte) {
	chunk := bytes.SplitN(data, []byte{'='}, 2)
	if len(chunk) != 2 {
		return nil, nil
	}
	key := chunk[0]
	value := chunk[1]

	if len(value) < 3 || len(key) == 0 {
		return nil, nil
	}
	if value[0] != '"' || value[len(value)-1] != '"' {
		return key, nil
	}

	// Strip the quotes
	value = bytes.Replace(value, []byte{'"'}, nil, -1)
	return key, value
}

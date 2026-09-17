package parser

import (
	"github.com/gomarkdown/markdown/ast"
)

// Parsing block-level elements.

const (
	captionTable  = "Table: "
	captionFigure = "Figure: "
	captionQuote  = "Quote: "
)

func (p *Parser) Block(data []byte) {
	// this is called recursively: enforce a maximum depth
	if p.nesting >= p.maxNesting {
		return
	}
	p.nesting++

	// parse out one block-level construct at a time
	for len(data) > 0 {
		// attributes that can be specified before a block element:
		//
		// {#id .class1 .class2 key="value"}
		//
		// kramdown also allows an IAL on the line after a block:
		//
		// ## foo
		// {: data-line="1"}
		if p.extensions&Attributes != 0 {
			if n := p.applyAfterBlockAttribute(data); n > 0 {
				data = data[n:]
				continue
			}
			data = p.attribute(data)
		}

		if p.extensions&Includes != 0 {
			f := p.readInclude
			path, address, consumed := isInclude(data)
			if consumed == 0 {
				path, address, consumed = isCodeInclude(data)
				f = p.readCodeInclude
			}
			if consumed > 0 {
				included := f(p.includeStack.Last(), path, address)

				// Optional caption on the line after the include. Skip a following
				// newline when present; do not assume data[consumed+1] exists (EOF
				// after {{file}} used to panic with slice bounds out of range).
				if consumed < len(data) {
					rest := data[consumed:]
					captionOff := 0
					if rest[0] == '\n' {
						captionOff = 1
					}
					if captionOff < len(rest) {
						for _, caption := range []string{captionFigure, captionTable, captionQuote} {
							if _, _, capcon := parseCaption(rest[captionOff:], []byte(caption)); capcon > 0 {
								included = append(included, rest[captionOff:captionOff+capcon]...)
								consumed += captionOff + capcon
								break // there can only be 1 caption.
							}
						}
					}
				}
				p.includeStack.Push(path)
				p.Block(included)
				p.includeStack.Pop()
				data = data[consumed:]
				continue
			}
		}

		// user supplied parser function
		if p.Opts.ParserHook != nil {
			node, blockdata, consumed := p.Opts.ParserHook(data)
			if consumed > 0 {
				data = data[consumed:]

				if node != nil {
					p.AddBlock(node)
					if blockdata != nil {
						p.Block(blockdata)
						p.Finalize(node)
					}
				}
				continue
			}
		}

		if len(data) == 0 {
			continue
		}
		data = data[p.parseBlock(data):]
	}

	p.nesting--
}

func (p *Parser) AddBlock(n ast.Node) ast.Node {
	p.closeUnmatchedBlocks()

	if p.attr != nil {
		if c := n.AsContainer(); c != nil {
			c.Attribute = p.attr
		}
		if l := n.AsLeaf(); l != nil {
			l.Attribute = p.attr
		}
		p.attr = nil
	}
	return p.addChild(n)
}

func (p *Parser) blockMath(data []byte) int {
	if len(data) <= 4 || data[0] != '$' || data[1] != '$' || data[2] == '$' {
		return 0
	}

	// find next $$
	var end int
	for end = 2; end+1 < len(data) && (data[end] != '$' || data[end+1] != '$'); end++ {
	}

	// $$ not match
	if end+1 == len(data) {
		return 0
	}

	// render the display math
	mathBlock := &ast.MathBlock{}
	mathBlock.Literal = data[2:end]
	p.AddBlock(mathBlock)

	return end + 2
}

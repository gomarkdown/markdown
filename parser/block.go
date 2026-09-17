package parser

import (
	"bytes"
	"strconv"

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
			path, address, consumed := p.isInclude(data)
			if consumed == 0 {
				path, address, consumed = p.isCodeInclude(data)
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
							if _, _, capcon := p.caption(rest[captionOff:], []byte(caption)); capcon > 0 {
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

		// prefixed heading:
		//
		// # Heading 1
		// ## Heading 2
		// ...
		// ###### Heading 6
		if p.isPrefixHeading(data) {
			data = data[p.prefixHeading(data):]
			continue
		}

		// prefixed special heading:
		// (there are no levels.)
		//
		// .# Abstract
		if p.isPrefixSpecialHeading(data) {
			data = data[p.prefixSpecialHeading(data):]
			continue
		}

		// block of preformatted HTML:
		//
		// <div>
		//     ...
		// </div>

		if len(data) == 0 {
			continue
		}

		if data[0] == '<' {
			if i := p.html(data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		// title block
		//
		// % stuff
		// % more stuff
		// % even more stuff
		if p.extensions&Titleblock != 0 {
			if data[0] == '%' {
				if i := p.titleBlock(data, true); i > 0 {
					data = data[i:]
					continue
				}
			}
		}

		// blank lines.  note: returns the # of bytes to skip
		if i := IsEmpty(data); i > 0 {
			data = data[i:]
			continue
		}

		// indented code block:
		//
		//     func max(a, b int) int {
		//         if a > b {
		//             return a
		//         }
		//         return b
		//      }
		if p.codePrefix(data) > 0 {
			data = data[p.code(data):]
			continue
		}

		// fenced code block:
		//
		// ``` go
		// func fact(n int) int {
		//     if n <= 1 {
		//         return n
		//     }
		//     return n * fact(n-1)
		// }
		// ```
		if p.extensions&FencedCode != 0 {
			if i := p.fencedCodeBlock(data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		// horizontal rule:
		//
		// ------
		// or
		// ******
		// or
		// ______
		if isHRule(data) {
			i := skipUntilChar(data, 0, '\n')
			hr := ast.HorizontalRule{}
			hr.Literal = bytes.Trim(data[:i], " \n")
			p.AddBlock(&hr)
			data = data[i:]
			continue
		}

		// block quote:
		//
		// > A big quote I found somewhere
		// > on the web
		if p.quotePrefix(data) > 0 {
			data = data[p.quote(data):]
			continue
		}

		// aside:
		//
		// A> The proof is too large to fit
		// A> in the margin.
		if p.extensions&Mmark != 0 {
			if p.asidePrefix(data) > 0 {
				data = data[p.aside(data):]
				continue
			}
		}

		// figure block:
		//
		// !---
		// ![Alt Text](img.jpg "This is an image")
		// ![Alt Text](img2.jpg "This is a second image")
		// !---
		if p.extensions&Mmark != 0 {
			if i := p.figureBlock(data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		if p.extensions&Tables != 0 {
			if i := p.table(data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// an itemized/unordered list:
		//
		// * Item 1
		// * Item 2
		//
		// also works with + or -
		if p.uliPrefix(data) > 0 {
			data = data[p.list(data, 0, 0, '.'):]
			continue
		}

		// a numbered/ordered list:
		//
		// 1. Item 1
		// 2. Item 2
		if i := p.oliPrefix(data); i > 0 {
			start := 0
			delim := byte('.')
			if i > 2 {
				if p.extensions&OrderedListStart != 0 {
					s := string(data[:i-2])
					start, _ = strconv.Atoi(s)
					if start == 1 {
						start = 0
					}
				}
				delim = data[i-2]
			}
			data = data[p.list(data, ast.ListTypeOrdered, start, delim):]
			continue
		}

		// definition lists:
		//
		// Term 1
		// :   Definition a
		// :   Definition b
		//
		// Term 2
		// :   Definition c
		if p.extensions&DefinitionLists != 0 {
			if p.dliPrefix(data) > 0 {
				data = data[p.list(data, ast.ListTypeDefinition, 0, '.'):]
				continue
			}
		}

		if p.extensions&MathJax != 0 {
			if i := p.blockMath(data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// document matters:
		//
		// {frontmatter}/{mainmatter}/{backmatter}
		if p.extensions&Mmark != 0 {
			if i := p.documentMatter(data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// anything else must look like a normal paragraph
		// note: this finds underlined headings, too
		idx := p.paragraph(data)
		data = data[idx:]
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

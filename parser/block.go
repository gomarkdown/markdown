package parser

import (
	"bytes"
	"strconv"
	"unicode"

	"github.com/gomarkdown/markdown/ast"
)

// Parsing block-level elements.

const (
	captionTable  = "Table: "
	captionFigure = "Figure: "
	captionQuote  = "Quote: "
)

// sanitizeHeadingID returns a sanitized anchor name for the given text.
// Taken from https://github.com/shurcooL/sanitized_anchor_name/blob/master/main.go#L14:1
func sanitizeHeadingID(text string) string {
	var anchorName []rune
	var futureDash = false
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			if futureDash && len(anchorName) > 0 {
				anchorName = append(anchorName, '-')
			}
			futureDash = false
			anchorName = append(anchorName, unicode.ToLower(r))
		default:
			futureDash = true
		}
	}
	if len(anchorName) == 0 {
		return "empty"
	}
	return string(anchorName)
}

// Parse Block-level data.
// Note: this function and many that it calls assume that
// the input buffer ends with a newline.
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

func (p *Parser) isPrefixHeading(data []byte) bool {
	if len(data) > 0 && data[0] != '#' {
		return false
	}

	if p.extensions&SpaceHeadings != 0 {
		level := skipCharN(data, 0, '#', 6)
		if level == len(data) || data[level] != ' ' {
			return false
		}
	}
	return true
}

// parseHeadingContent extracts the text range and optional {#id} for a heading
// whose content starts at i and whose line ends just before the newline at end.
// It returns the heading id ("" if none), the index where the heading text ends
// (after trimming a trailing {#id} and any closing '#' markers and spaces), and
// the number of bytes the whole heading line occupies.
func (p *Parser) parseHeadingContent(data []byte, i, end int) (id string, contentEnd, skip int) {
	skip = end
	if p.extensions&HeadingIDs != 0 {
		j, k := 0, 0
		// find start/end of heading id
		for j = i; j < end-1 && (data[j] != '{' || data[j+1] != '#'); j++ {
		}
		for k = j + 1; k < end && data[k] != '}'; k++ {
		}
		// extract heading id iff found
		if j < end && k < end {
			id = string(data[j+2 : k])
			end = j
			skip = k + 1
			end = backChar(data, end, ' ')
		}
	}
	// strip trailing closing '#' markers and surrounding spaces
	for end > 0 && data[end-1] == '#' {
		if isBackslashEscaped(data, end-1) {
			break
		}
		end--
	}
	end = backChar(data, end, ' ')
	return id, end, skip
}

// setHeadingID assigns block's id, auto-generating one from text when id is
// empty and AutoHeadingIDs is enabled (recording it for later uniquification).
func (p *Parser) setHeadingID(block *ast.Heading, id string, text []byte) {
	block.HeadingID = id
	if id == "" && p.extensions&AutoHeadingIDs != 0 {
		block.HeadingID = sanitizeHeadingID(string(text))
		p.allHeadingsWithAutoID = append(p.allHeadingsWithAutoID, block)
	}
}

func (p *Parser) prefixHeading(data []byte) int {
	level := skipCharN(data, 0, '#', 6)
	i := skipChar(data, level, ' ')
	end := skipUntilChar(data, i, '\n')
	id, end, skip := p.parseHeadingContent(data, i, end)
	if end > i {
		block := &ast.Heading{
			Level: level,
		}
		p.setHeadingID(block, id, data[i:end])
		block.Content = data[i:end]
		p.AddBlock(block)
	}
	return skip
}

func (p *Parser) isPrefixSpecialHeading(data []byte) bool {
	if p.extensions|Mmark == 0 {
		return false
	}
	if len(data) < 4 {
		return false
	}
	if data[0] != '.' {
		return false
	}
	if data[1] != '#' {
		return false
	}
	if data[2] == '#' { // we don't support level, so nack this.
		return false
	}

	if p.extensions&SpaceHeadings != 0 {
		if data[2] != ' ' {
			return false
		}
	}
	return true
}

func (p *Parser) prefixSpecialHeading(data []byte) int {
	i := skipChar(data, 2, ' ') // ".#" skipped
	end := skipUntilChar(data, i, '\n')
	id, end, skip := p.parseHeadingContent(data, i, end)
	if end > i {
		block := &ast.Heading{
			IsSpecial: true,
			Level:     1, // always level 1.
		}
		p.setHeadingID(block, id, data[i:end])
		block.Literal = data[i:end]
		block.Content = data[i:end]
		p.AddBlock(block)
	}
	return skip
}

func (p *Parser) isUnderlinedHeading(data []byte) int {
	// test of level 1 heading
	if data[0] == '=' {
		i := skipChar(data, 1, '=')
		i = skipChar(data, i, ' ')
		if i < len(data) && data[i] == '\n' {
			return 1
		}
		return 0
	}

	// test of level 2 heading
	if data[0] == '-' {
		i := skipChar(data, 1, '-')
		i = skipChar(data, i, ' ')
		if i < len(data) && data[i] == '\n' {
			return 2
		}
		return 0
	}

	return 0
}

func (p *Parser) titleBlock(data []byte, doRender bool) int {
	if data[0] != '%' {
		return 0
	}
	splitData := bytes.Split(data, []byte("\n"))
	var i int
	for idx, b := range splitData {
		if !bytes.HasPrefix(b, []byte("%")) {
			i = idx // - 1
			break
		}
	}

	data = bytes.Join(splitData[0:i], []byte("\n"))
	consumed := len(data)
	data = bytes.TrimPrefix(data, []byte("% "))
	data = bytes.Replace(data, []byte("\n% "), []byte("\n"), -1)
	block := &ast.Heading{
		Level:        1,
		IsTitleblock: true,
	}
	block.Content = data
	p.AddBlock(block)

	return consumed
}

func isHRule(data []byte) bool {
	i := 0

	// skip up to three spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// look at the hrule char
	if data[i] != '*' && data[i] != '-' && data[i] != '_' {
		return false
	}
	c := data[i]

	// the whole line must be the char or whitespace
	n := 0
	for i < len(data) && data[i] != '\n' {
		switch {
		case data[i] == c:
			n++
		case data[i] != ' ':
			return false
		}
		i++
	}

	return n >= 3
}

// isFenceLine checks if there's a fence line (e.g., ``` or ``` go) at the beginning of data,
// and returns the end index if so, or 0 otherwise. It also returns the marker found.
// If syntax is not nil, it gets set to the syntax specified in the fence line.
// returns blockquote prefix length
func (p *Parser) quotePrefix(data []byte) int {
	i := 0
	n := len(data)
	for i < 3 && i < n && data[i] == ' ' {
		i++
	}
	if i < n && data[i] == '>' {
		if i+1 < n && data[i+1] == ' ' {
			return i + 2
		}
		return i + 1
	}
	return 0
}

// blockquote ends with at least one blank line
// followed by something without a blockquote prefix
func (p *Parser) terminateBlockquote(data []byte, beg, end int) bool {
	if IsEmpty(data[beg:]) <= 0 {
		return false
	}
	if end >= len(data) {
		return true
	}
	return p.quotePrefix(data[end:]) == 0 && IsEmpty(data[end:]) == 0
}

// parse a blockquote fragment
func (p *Parser) quote(data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	fenceMarker := ""
	for beg < len(data) {
		end = beg
		for end < len(data) && data[end] != '\n' {
			end++
		}
		end = skipCharN(data, end, '\n', 1)
		contentBeg := beg
		if pre := p.quotePrefix(data[beg:]); pre > 0 {
			// skip the prefix
			contentBeg += pre
		} else if fenceMarker != "" {
			// Lines inside a quoted fenced code block may omit the quote
			// prefix. Keep them in the quote until the fence closes.
		} else if p.terminateBlockquote(data, beg, end) {
			break
		}
		// this line is part of the blockquote
		raw.Write(data[contentBeg:end])
		if p.extensions&FencedCode != 0 {
			if _, marker := isFenceLine(data[contentBeg:end], nil, fenceMarker); marker != "" {
				if fenceMarker == "" {
					fenceMarker = marker
				} else {
					fenceMarker = ""
				}
			}
		}
		beg = end
	}

	if p.extensions&Mmark == 0 {
		block := p.AddBlock(&ast.BlockQuote{})
		p.Block(raw.Bytes())
		p.Finalize(block)
		return end
	}

	if captionContent, id, consumed := p.caption(data[end:], []byte(captionQuote)); consumed > 0 {
		figure := &ast.CaptionFigure{}
		caption := &ast.Caption{}
		figure.HeadingID = id
		p.Inline(caption, captionContent)

		p.AddBlock(figure) // this discard any attributes
		block := &ast.BlockQuote{}
		block.AsContainer().Attribute = figure.AsContainer().Attribute
		p.addChild(block)
		p.Block(raw.Bytes())
		p.Finalize(block)

		p.addChild(caption)
		p.Finalize(figure)

		end += consumed

		return end
	}

	block := p.AddBlock(&ast.BlockQuote{})
	p.Block(raw.Bytes())
	p.Finalize(block)

	return end
}

// returns prefix length for block code
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

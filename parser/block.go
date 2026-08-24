package parser

import (
	"bytes"
	"html"
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
		// attributes that can be specific before a block element:
		//
		// {#id .class1 .class2 key="value"}
		if p.extensions&Attributes != 0 {
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
func isFenceLine(data []byte, syntax *string, oldmarker string) (end int, marker string) {
	i, size := 0, 0

	n := len(data)
	// skip up to three spaces
	for i < n && i < 3 && data[i] == ' ' {
		i++
	}

	// check for the marker characters: ~ or `
	if i >= n {
		return 0, ""
	}
	if data[i] != '~' && data[i] != '`' {
		return 0, ""
	}

	c := data[i]

	// the whole line must be the same char or whitespace
	for i < n && data[i] == c {
		size++
		i++
	}

	// the marker char must occur at least 3 times
	if size < 3 {
		return 0, ""
	}
	marker = string(data[i-size : i])

	// if this is the end marker, it must match the beginning marker
	if oldmarker != "" && marker != oldmarker {
		return 0, ""
	}

	// if just read the beginning marker, read the syntax
	if oldmarker == "" {
		i = skipChar(data, i, ' ')
		if i >= n {
			if i == n {
				return i, marker
			}
			return 0, ""
		}

		syntaxStart, syntaxLen := syntaxRange(data, &i)
		if syntaxStart == 0 && syntaxLen == 0 {
			return 0, ""
		}

		// caller wants the syntax
		if syntax != nil {
			*syntax = string(data[syntaxStart : syntaxStart+syntaxLen])
		}
	}

	i = skipChar(data, i, ' ')
	if i >= n || data[i] != '\n' {
		if i == n {
			return i, marker
		}
		return 0, ""
	}
	return i + 1, marker // Take newline into account.
}

func syntaxRange(data []byte, iout *int) (int, int) {
	n := len(data)
	syn := 0
	i := *iout
	syntaxStart := i
	if data[i] == '{' {
		i++
		syntaxStart++

		for i < n && data[i] != '}' && data[i] != '\n' {
			syn++
			i++
		}

		if i >= n || data[i] != '}' {
			return 0, 0
		}

		// strip all whitespace at the beginning and the end
		// of the {} block
		for syn > 0 && IsSpace(data[syntaxStart]) {
			syntaxStart++
			syn--
		}

		for syn > 0 && IsSpace(data[syntaxStart+syn-1]) {
			syn--
		}

		i++
	} else {
		for i < n && data[i] != '\n' {
			syn++
			i++
		}
	}

	*iout = i
	return syntaxStart, syn
}

// fencedCodeBlock returns the end index if data contains a fenced code block at the beginning,
// or 0 otherwise. It writes to out if doRender is true, otherwise it has no side effects.
// If doRender is true, a final newline is mandatory to recognize the fenced code block.
func (p *Parser) fencedCodeBlock(data []byte, doRender bool) int {
	var syntax string
	beg, marker := isFenceLine(data, &syntax, "")
	if beg == 0 || beg >= len(data) {
		return 0
	}

	var work bytes.Buffer
	work.WriteString(syntax)
	work.WriteByte('\n')

	for {
		// check for the end of the code block
		fenceEnd, _ := isFenceLine(data[beg:], nil, marker)
		if fenceEnd != 0 {
			beg += fenceEnd
			break
		}

		// copy the current line
		end := skipUntilChar(data, beg, '\n') + 1

		// did we reach the end of the buffer without a closing marker?
		if end >= len(data) {
			return 0
		}

		// verbatim copy to the working buffer
		work.Write(data[beg:end])
		beg = end
	}

	if !doRender {
		return beg
	}
	codeBlock := &ast.CodeBlock{
		IsFenced: true,
	}
	codeBlock.Content = work.Bytes() // TODO: get rid of temp buffer

	if p.extensions&Mmark == 0 {
		p.AddBlock(codeBlock)
		finalizeCodeBlock(codeBlock)
		return beg
	}

	// Check for caption and if found make it a figure.
	if captionContent, id, consumed := p.caption(data[beg:], []byte(captionFigure)); consumed > 0 {
		figure := &ast.CaptionFigure{}
		caption := &ast.Caption{}
		figure.HeadingID = id
		p.Inline(caption, captionContent)

		p.AddBlock(figure)
		codeBlock.AsLeaf().Attribute = figure.AsContainer().Attribute
		p.addChild(codeBlock)
		finalizeCodeBlock(codeBlock)
		p.addChild(caption)
		p.Finalize(figure)

		beg += consumed

		return beg
	}

	// Still here, normal block
	p.AddBlock(codeBlock)
	finalizeCodeBlock(codeBlock)

	return beg
}

func unescapeString(str []byte) []byte {
	var out []byte
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '\\':
			if i+1 < len(str) && isEscapable(str[i+1]) {
				if out == nil {
					out = make([]byte, 0, len(str))
					out = append(out, str[:i]...)
				}
				out = append(out, str[i+1])
				i++
				continue
			}
		case '&':
			entityEnd := findEntityEnd(str, i)
			if entityEnd > i {
				replacement := html.UnescapeString(string(str[i:entityEnd]))
				if replacement != string(str[i:entityEnd]) {
					if out == nil {
						out = make([]byte, 0, len(str))
						out = append(out, str[:i]...)
					}
					out = append(out, replacement...)
					i = entityEnd - 1
					continue
				}
			}
		}
		if out != nil {
			out = append(out, str[i])
		}
	}
	if out != nil {
		return out
	}
	return str
}

func isEscapable(c byte) bool {
	switch c {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '.', '/', ':',
		';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~', '-':
		return true
	default:
		return false
	}
}

func findEntityEnd(str []byte, start int) int {
	i := start + 1
	if i >= len(str) {
		return 0
	}
	if str[i] == '#' {
		i++
		if i >= len(str) {
			return 0
		}
		if str[i] == 'x' || str[i] == 'X' {
			i++
			digits := 0
			for i < len(str) && digits < 8 && isHexDigit(str[i]) {
				i++
				digits++
			}
			if digits == 0 || i >= len(str) || str[i] != ';' {
				return 0
			}
			return i + 1
		}
		digits := 0
		for i < len(str) && digits < 8 && str[i] >= '0' && str[i] <= '9' {
			i++
			digits++
		}
		if digits == 0 || i >= len(str) || str[i] != ';' {
			return 0
		}
		return i + 1
	}
	if !isAlpha(str[i]) {
		return 0
	}
	i++
	for i < len(str) && i-start <= 32 && isAlnum(str[i]) {
		i++
	}
	if i >= len(str) || str[i] != ';' {
		return 0
	}
	return i + 1
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isAlnum(c byte) bool {
	return isAlpha(c) || c >= '0' && c <= '9'
}

func isHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func finalizeCodeBlock(code *ast.CodeBlock) {
	c := code.Content
	if code.IsFenced {
		newlinePos := bytes.IndexByte(c, '\n')
		firstLine := c[:newlinePos]
		rest := c[newlinePos+1:]
		code.Info = unescapeString(bytes.Trim(firstLine, "\n"))
		code.Literal = rest
	} else {
		code.Literal = c
	}
	code.Content = nil
}

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
func (p *Parser) codePrefix(data []byte) int {
	n := len(data)
	if n >= 1 && data[0] == '\t' {
		return 1
	}
	if n >= 4 && data[3] == ' ' && data[2] == ' ' && data[1] == ' ' && data[0] == ' ' {
		return 4
	}
	return 0
}

func (p *Parser) code(data []byte) int {
	var work bytes.Buffer

	i := 0
	for i < len(data) {
		beg := i

		i = skipUntilChar(data, i, '\n')
		i = skipCharN(data, i, '\n', 1)

		blankline := IsEmpty(data[beg:i]) > 0
		if pre := p.codePrefix(data[beg:i]); pre > 0 {
			beg += pre
		} else if !blankline {
			// non-empty, non-prefixed line breaks the pre
			i = beg
			break
		}

		// verbatim copy to the working buffer
		if blankline {
			work.WriteByte('\n')
		} else {
			work.Write(data[beg:i])
		}
	}

	// trim all the \n off the end of work
	workbytes := work.Bytes()

	eol := backChar(workbytes, len(workbytes), '\n')

	if eol != len(workbytes) {
		work.Truncate(eol)
	}

	work.WriteByte('\n')

	codeBlock := &ast.CodeBlock{
		IsFenced: false,
	}
	// TODO: get rid of temp buffer
	codeBlock.Content = work.Bytes()
	p.AddBlock(codeBlock)
	finalizeCodeBlock(codeBlock)

	return i
}

// returns unordered list item prefix
func (p *Parser) renderParagraph(data []byte) {
	if len(data) == 0 {
		return
	}

	// trim leading spaces
	beg := skipChar(data, 0, ' ')

	end := len(data)
	// trim trailing newline
	if data[len(data)-1] == '\n' {
		end--
	}

	// trim trailing spaces
	for end > beg && data[end-1] == ' ' {
		end--
	}
	para := &ast.Paragraph{}
	para.Content = data[beg:end]
	p.AddBlock(para)
}

// blockMath handle block surround with $$
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

func (p *Parser) paragraph(data []byte) int {
	// prev: index of 1st char of previous line
	// line: index of 1st char of current line
	// i: index of cursor/end of current line
	var prev, line, i int
	tabSize := tabSizeDefault
	if p.extensions&TabSizeEight != 0 {
		tabSize = tabSizeDouble
	}
	// keep going until we find something to mark the end of the paragraph
	for i < len(data) {
		// mark the beginning of the current line
		prev = line
		current := data[i:]
		line = i

		// did we find a reference or a footnote? If so, end a paragraph
		// preceding it and report that we have consumed up to the end of that
		// reference:
		if refEnd := isReference(p, current, tabSize); refEnd > 0 {
			p.renderParagraph(data[:i])
			return i + refEnd
		}

		// did we find a blank line marking the end of the paragraph?
		if n := IsEmpty(current); n > 0 {
			// did this blank line followed by a definition list item?
			if p.extensions&DefinitionLists != 0 {
				if i < len(data)-1 && data[i+1] == ':' {
					listLen := p.list(data[prev:], ast.ListTypeDefinition, 0, '.')
					if listLen > 0 {
						return prev + listLen
					}
				}
			}

			p.renderParagraph(data[:i])
			return i + n
		}

		// an underline under some text marks a heading, so our paragraph ended on prev line
		if i > 0 {
			if level := p.isUnderlinedHeading(current); level > 0 {
				// render the paragraph
				p.renderParagraph(data[:prev])

				// ignore leading and trailing whitespace
				eol := i - 1
				for prev < eol && data[prev] == ' ' {
					prev++
				}
				for eol > prev && data[eol-1] == ' ' {
					eol--
				}

				block := &ast.Heading{
					Level: level,
				}
				p.setHeadingID(block, "", data[prev:eol])

				block.Content = data[prev:eol]
				p.AddBlock(block)

				// find the end of the underline
				return skipUntilChar(data, i, '\n')
			}
		}

		// if the next line starts a block of HTML, then the paragraph ends here
		if p.extensions&LaxHTMLBlocks != 0 {
			if data[i] == '<' && p.html(current, false) > 0 {
				// rewind to before the HTML block
				p.renderParagraph(data[:i])
				return i
			}
		}

		// if there's a prefixed heading or a horizontal rule after this, paragraph is over
		if p.isPrefixHeading(current) || p.isPrefixSpecialHeading(current) || isHRule(current) {
			p.renderParagraph(data[:i])
			return i
		}

		// if there's a block quote, paragraph is over
		if p.quotePrefix(current) > 0 {
			p.renderParagraph(data[:i])
			return i
		}

		// if there's a fenced code block, paragraph is over
		if p.extensions&FencedCode != 0 {
			if p.fencedCodeBlock(current, false) > 0 {
				p.renderParagraph(data[:i])
				return i
			}
		}

		// if there's a figure block, paragraph is over
		if p.extensions&Mmark != 0 {
			if p.figureBlock(current, false) > 0 {
				p.renderParagraph(data[:i])
				return i
			}
		}

		// if there's a table, paragraph is over
		if p.extensions&Tables != 0 {
			if j, _, _ := p.tableHeader(current, false); j > 0 {
				p.renderParagraph(data[:i])
				return i
			}
		}

		// if there's a definition list item, prev line is a definition term
		if p.extensions&DefinitionLists != 0 {
			if p.dliPrefix(current) != 0 {
				ret := p.list(data[prev:], ast.ListTypeDefinition, 0, '.')
				return ret + prev
			}
		}

		// if there's a list after this, paragraph is over
		if p.extensions&NoEmptyLineBeforeBlock != 0 {
			if p.uliPrefix(current) != 0 ||
				p.oliPrefix(current) != 0 ||
				p.quotePrefix(current) != 0 ||
				p.codePrefix(current) != 0 {
				p.renderParagraph(data[:i])
				return i
			}
		}

		// otherwise, scan to the beginning of the next line
		nl := bytes.IndexByte(data[i:], '\n')
		if nl >= 0 {
			i += nl + 1
		} else {
			i += len(data[i:])
		}
	}

	p.renderParagraph(data[:i])
	return i
}

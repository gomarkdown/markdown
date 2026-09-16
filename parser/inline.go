package parser

import (
	"bytes"
	"regexp"

	"github.com/gomarkdown/markdown/ast"
)

// Parsing of inline elements

var (
	urlRe    = `((https?|ftp):\/\/|\/)[-A-Za-z0-9+&@#\/%?=~_|!:,.;\(\)]+`
	anchorRe = regexp.MustCompile(`^(<a\shref="` + urlRe + `"(\stitle="[^"<>]+")?\s?>` + urlRe + `<\/a>)`)

	// TODO: improve this regexp to catch all possible entities:
	htmlEntityRe = regexp.MustCompile(`&[a-z]{2,5};`)
)

// Inline parses text within a block.
// Each function returns the number of consumed chars.
func (p *Parser) Inline(currBlock ast.Node, data []byte) {
	// handlers might call us recursively: enforce a maximum depth
	if p.nesting >= p.maxNesting || len(data) == 0 {
		return
	}
	if p.nesting == 0 {
		p.resetInlineCaches()
	}
	p.nesting++
	prev := p.brackets
	p.brackets = bracketTable{data: data}
	defer func() {
		p.brackets = prev
		p.nesting--
	}()
	beg, end := 0, 0

	n := len(data)
	for end < n {
		handler := p.inlineCallback[data[end]]
		if handler == nil {
			end++
			continue
		}
		consumed, node := handler(p, data, end)
		if consumed == 0 {
			// no action from the callback
			end++
			continue
		}
		// copy inactive chars into the output
		ast.AppendChild(currBlock, newTextNode(data[beg:end]))
		if node != nil {
			ast.AppendChild(currBlock, node)
		}
		beg = end + consumed
		end = beg
	}

	if beg < n {
		if data[end-1] == '\n' {
			end--
		}
		ast.AppendChild(currBlock, newTextNode(data[beg:end]))
	}
}

// resetInlineCaches drops the memo tables that inline callbacks keep for
// the buffer under the cursor. They key on the buffer's address and
// length, so a caller that reuses a Parser on a buffer it has rewritten in
// place must not see entries built from the old contents.
func (p *Parser) resetInlineCaches() {
	p.codeSpans.data = nil
	p.spaces.data = nil
	p.angles.data = nil
	p.nextGt.data = nil
	p.nextCommentEnd.data = nil
}

// runCache remembers the end of the run of one byte value that the inline
// cursor is inside. A callback registered for that byte fires once per
// byte of the run, and without the cache each firing rescanned the run.
type runCache struct {
	data       *byte
	n          int
	start, end int
}

// endOf returns the index just past the run of b that starts at
// data[offset], scanning only when offset is outside the remembered run.
func (c *runCache) endOf(data []byte, offset int, b byte) int {
	if c.data == &data[0] && c.n == len(data) && offset >= c.start && offset < c.end {
		return c.end
	}
	end := skipChar(data, offset, b)
	*c = runCache{data: &data[0], n: len(data), start: offset, end: end}
	return end
}

// single and double emphasis parsing
func emphasis(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]
	c := data[0]

	n := len(data)
	if n > 2 && data[1] != c {
		// whitespace cannot follow an opening emphasis;
		// strikethrough only takes two characters '~~'
		if IsSpace(data[1]) {
			return 0, nil
		}
		if p.extensions&SuperSubscript != 0 && c == '~' {
			// potential subscript, no spaces, except when escaped, helperEmphasis does
			// not check that for us, so walk the bytes and check.
			ret := skipUntilChar(data[1:], 0, c)
			if ret == 0 || ret >= len(data)-1 {
				// empty, or no closing '~'
				return 0, nil
			}
			ret++ // we started with data[1:] above.
			for i := 1; i < ret; i++ {
				if IsSpace(data[i]) && !isEscape(data, i) {
					return 0, nil
				}
			}
			sub := &ast.Subscript{}
			sub.Literal = data[1:ret]
			return ret + 1, sub
		}
		ret, node := helperEmphasis(p, data[1:], c)
		if ret == 0 {
			return 0, nil
		}

		return ret + 1, node
	}

	if n > 3 && data[1] == c && data[2] != c {
		if IsSpace(data[2]) {
			return 0, nil
		}
		ret, node := helperDoubleEmphasis(p, data[2:], c)
		if ret == 0 {
			return 0, nil
		}

		return ret + 2, node
	}

	if n > 4 && data[1] == c && data[2] == c && data[3] != c {
		if c == '~' || IsSpace(data[3]) {
			return 0, nil
		}
		ret, node := helperTripleEmphasis(p, data, 3, c)
		if ret == 0 {
			return 0, nil
		}

		return ret + 3, node
	}

	return 0, nil
}

// codeSpanEnd returns the index just past the code span that opens with the
// backtick run at data[0], or 0 when that run has no closing delimiter. The
// closing delimiter is the first later run of at least as many backticks.
// It is the reference form of the search that codeSpanClosers memoises.
func codeSpanEnd(data []byte) int {
	// count the number of backticks in the delimiter
	nb := skipChar(data, 0, '`')

	// find the next delimiter
	i, end := 0, 0
	for end = nb; end < len(data) && i < nb; end++ {
		if data[end] == '`' {
			i++
		} else {
			i = 0
		}
	}

	// no matching delimiter?
	if i < nb && end >= len(data) {
		return 0
	}
	return end
}

// codeSpanCache holds the closer table for one backtick run. When the run
// has no closing delimiter, Inline advances one byte and calls codeSpan on
// the same run minus its first backtick, and so on down to a single
// backtick. Without the cache every retry rescanned to the end of the
// input, which is quadratic in the run length.
type codeSpanCache struct {
	data     *byte // first byte of the slice the entries index into
	n        int
	runStart int // the run the entries describe is data[runStart:runEnd]
	runEnd   int
	closers  []int  // closers[m]: index just past the m-th backtick of the first run of at least m backticks after runEnd, or 0
	small    [8]int // backing store for closers when the run is short, which is nearly always
}

// codeSpanClosers returns the length of the backtick run at data[offset]
// and a table whose entry m is where a delimiter of m backticks closes,
// for every m up to that length. The table is rebuilt only when the run
// changes, so retries from inside the same run cost nothing, not even the
// count of the remaining backticks.
func (p *Parser) codeSpanClosers(data []byte, offset int) (int, []int) {
	c := &p.codeSpans
	if c.data == &data[0] && c.n == len(data) && offset >= c.runStart && offset < c.runEnd {
		return c.runEnd - offset, c.closers
	}

	runEnd := skipChar(data, offset, '`')
	k := runEnd - offset
	closers := c.closers
	if cap(closers) <= k {
		closers = c.small[:]
		if k >= len(c.small) {
			closers = make([]int, 0, k+1)
		}
	}
	closers = append(closers[:0], 0)
	run, filled := 0, 0
	for j := runEnd; j < len(data) && filled < k; j++ {
		if data[j] != '`' {
			run = 0
			continue
		}
		run++
		if run > filled {
			closers = append(closers, j+1)
			filled = run
		}
	}
	for len(closers) <= k {
		closers = append(closers, 0)
	}
	c.data, c.n = &data[0], len(data)
	c.runStart, c.runEnd = offset, runEnd
	c.closers = closers
	return k, closers
}

func codeSpan(p *Parser, data []byte, offset int) (int, ast.Node) {
	nb, closers := p.codeSpanClosers(data, offset)
	if nb == 0 {
		return 0, nil
	}
	end := closers[nb] - offset
	data = data[offset:]
	if end <= 0 {
		return 0, nil
	}
	hasLFBeforeDelimiter := bytes.IndexByte(data[nb:end], '\n') >= 0

	// If there are non-space chars after the ending delimiter and before a '\n',
	// flag that this is not a well formed fenced code block.
	hasCharsAfterDelimiter := false
	for j := end; j < len(data); j++ {
		if data[j] == '\n' {
			break
		}
		if !IsSpace(data[j]) {
			hasCharsAfterDelimiter = true
			break
		}
	}

	// trim outside whitespace
	fBegin := nb
	for fBegin < end && data[fBegin] == ' ' {
		fBegin++
	}

	fEnd := end - nb
	for fEnd > fBegin && data[fEnd-1] == ' ' {
		fEnd--
	}

	if fBegin == fEnd {
		return end, nil
	}

	// if delimiter has 3 backticks
	if nb == 3 {
		i := fBegin
		syntaxStart, syntaxLen := syntaxRange(data, &i)

		// If we found a '\n' before the end marker and there are only spaces
		// after the end marker, then this is a code block.
		if hasLFBeforeDelimiter && !hasCharsAfterDelimiter {
			codeblock := &ast.CodeBlock{
				IsFenced: true,
				Info:     data[syntaxStart : syntaxStart+syntaxLen],
			}
			codeblock.Literal = data[i:fEnd]
			return end, codeblock
		}
	}

	// render the code span
	code := &ast.Code{}
	code.Literal = data[fBegin:fEnd]
	return end, code
}

// newline preceded by two spaces becomes <br>
func maybeLineBreak(p *Parser, data []byte, offset int) (int, ast.Node) {
	origOffset := offset
	offset = p.spaces.endOf(data, offset, ' ')

	if offset < len(data) && data[offset] == '\n' {
		if offset-origOffset >= 2 {
			return offset - origOffset + 1, &ast.Hardbreak{}
		}
		return offset - origOffset, nil
	}
	return 0, nil
}

// newline without two spaces works when HardLineBreak is enabled
func lineBreak(p *Parser, data []byte, offset int) (int, ast.Node) {
	if p.extensions&HardLineBreak != 0 {
		return 1, &ast.Hardbreak{}
	}
	return 0, nil
}

func maybeImage(p *Parser, data []byte, offset int) (int, ast.Node) {
	if offset < len(data)-1 && data[offset+1] == '[' {
		return link(p, data, offset)
	}
	return 0, nil
}

func maybeInlineFootnoteOrSuper(p *Parser, data []byte, offset int) (int, ast.Node) {
	if offset < len(data)-1 && data[offset+1] == '[' {
		return link(p, data, offset)
	}

	if p.extensions&SuperSubscript != 0 {
		ret := skipUntilChar(data[offset:], 1, '^')
		if ret >= len(data)-offset {
			// no closing '^'
			return 0, nil
		}
		for i := offset; i < offset+ret; i++ {
			if IsSpace(data[i]) && !isEscape(data, i) {
				return 0, nil
			}
		}
		sup := &ast.Superscript{}
		sup.Literal = data[offset+1 : offset+ret]
		return ret + 1, sup
	}

	return 0, nil
}

// '[': parse a link or an image or a footnote or a citation
func helperFindEmphChar(data []byte, c byte) int {
	i := 0

	for i < len(data) {
		for i < len(data) && data[i] != c && data[i] != '`' && data[i] != '[' {
			i++
		}
		if i >= len(data) {
			return 0
		}
		// do not count escaped chars; an even run of backslashes escapes
		// itself, not the delimiter
		if isEscape(data, i) {
			i++
			continue
		}
		if data[i] == c {
			return i
		}

		if data[i] == '`' {
			// skip a code span
			tmpI := 0
			i++
			for i < len(data) && data[i] != '`' {
				if tmpI == 0 && data[i] == c {
					tmpI = i
				}
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			i++
		} else if data[i] == '[' {
			// skip a link
			tmpI := 0
			i++
			for i < len(data) && data[i] != ']' {
				if tmpI == 0 && data[i] == c {
					tmpI = i
				}
				i++
			}
			i++
			for i < len(data) && (data[i] == ' ' || data[i] == '\n') {
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			if data[i] != '[' && data[i] != '(' { // not a link
				if tmpI > 0 {
					return tmpI
				}
				continue
			}
			cc := data[i]
			i++
			for i < len(data) && data[i] != cc {
				if tmpI == 0 && data[i] == c {
					return i
				}
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			i++
		}
	}
	return 0
}

func helperEmphasis(p *Parser, data []byte, c byte) (int, ast.Node) {
	i := 0

	// skip two symbol if coming from emph3, as it detected a double emphasis case
	if len(data) > 1 && data[0] == c && data[1] == c {
		i = 2
	}

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		i += length
		if i >= len(data) {
			return 0, nil
		}

		if i+1 < len(data) && data[i+1] == c {
			i += 2
			continue
		}

		if data[i] == c && !IsSpace(data[i-1]) {
			if p.extensions&NoIntraEmphasis != 0 {
				rest := data[i+1:]
				if !(len(rest) == 0 || IsSpace(rest[0]) || IsPunctuation2(rest)) {
					if length == 0 {
						return 0, nil
					}
					continue
				}
			}

			emph := &ast.Emph{}
			p.Inline(emph, data[:i])
			return i + 1, emph
		}

		// We have to check this at the end, otherwise the scenario where we find repeated c's will get skipped
		if length == 0 {
			return 0, nil
		}
	}

	return 0, nil
}

func helperDoubleEmphasis(p *Parser, data []byte, c byte) (int, ast.Node) {
	i := 0

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		if length == 0 {
			return 0, nil
		}
		i += length

		if i+1 < len(data) && data[i] == c && data[i+1] == c && i > 0 && !IsSpace(data[i-1]) {
			// When the closing delimiter is *** (3+ chars) and there is an
			// unclosed single emphasis opener inside the content, include
			// one extra char in the content so that the inner emphasis can
			// pair with it. For example: **bold *ital*** should produce
			// <strong>bold <em>ital</em></strong>, not <strong>bold *ital</strong>*.
			// See https://github.com/gomarkdown/markdown/issues/279
			contentEnd := i
			if i+2 < len(data) && data[i+2] == c && c != '~' {
				if hasTrailingEmphOpener(data[:i], c) {
					contentEnd = i + 1
				}
			}

			var node ast.Node = &ast.Strong{}
			if c == '~' {
				node = &ast.Del{}
			}
			p.Inline(node, data[:contentEnd])
			return contentEnd + 2, node
		}
		i++
	}
	return 0, nil
}

// hasTrailingEmphOpener checks if the last occurrence of c in data is an
// unclosed opener. An opener is c preceded by whitespace or start of data,
// followed by non-whitespace. If the last c is a closer (preceded by
// non-whitespace), the emphasis pair is balanced and we should not shift
// the content boundary.
func hasTrailingEmphOpener(data []byte, c byte) bool {
	// find the last c in data
	last := -1
	for j := len(data) - 1; j >= 0; j-- {
		if data[j] == c {
			last = j
			break
		}
	}
	if last < 0 {
		return false
	}
	// opener: preceded by space/start, followed by non-space
	return (last == 0 || IsSpace(data[last-1])) &&
		last+1 < len(data) && !IsSpace(data[last+1])
}

func helperTripleEmphasis(p *Parser, data []byte, offset int, c byte) (int, ast.Node) {
	i := 0
	origData := data
	data = data[offset:]

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		if length == 0 {
			return 0, nil
		}
		i += length

		// skip whitespace preceded symbols
		if data[i] != c || IsSpace(data[i-1]) {
			continue
		}

		switch {
		case i+2 < len(data) && data[i+1] == c && data[i+2] == c:
			// triple symbol found
			strong := &ast.Strong{}
			em := &ast.Emph{}
			ast.AppendChild(strong, em)
			p.Inline(em, data[:i])
			return i + 3, strong
		case i+1 < len(data) && data[i+1] == c:
			// double symbol found, hand over to emph1
			length, node := helperEmphasis(p, origData[offset-2:], c)
			if length == 0 {
				return 0, nil
			}
			return length - 2, node
		default:
			// single symbol found, hand over to emph2
			length, node := helperDoubleEmphasis(p, origData[offset-1:], c)
			if length == 0 {
				return 0, nil
			}
			return length - 1, node
		}
	}
	return 0, nil
}

// math handle inline math wrapped with '$'
func math(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]

	// too short, or block math
	if len(data) <= 2 || data[1] == '$' {
		return 0, nil
	}

	// find next '$'
	var end int
	for end = 1; end < len(data) && data[end] != '$'; end++ {
	}

	// $ not match
	if end == len(data) {
		return 0, nil
	}

	// create inline math node
	math := &ast.Math{}
	math.Literal = data[1:end]
	return end + 1, math
}

func newTextNode(d []byte) *ast.Text {
	return &ast.Text{Leaf: ast.Leaf{Literal: d}}
}

package parser

import (
	"bytes"
	"regexp"
	"strconv"

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
			if ret == 0 {
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

func codeSpan(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]

	// count the number of backticks in the delimiter
	nb := skipChar(data, 0, '`')

	// find the next delimiter
	i, end := 0, 0
	hasLFBeforeDelimiter := false
	for end = nb; end < len(data) && i < nb; end++ {
		if data[end] == '\n' {
			hasLFBeforeDelimiter = true
		}
		if data[end] == '`' {
			i++
		} else {
			i = 0
		}
	}

	// no matching delimiter?
	if i < nb && end >= len(data) {
		return 0, nil
	}

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
	offset = skipChar(data, offset, ' ')

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
		if ret == 0 {
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
func (p *Parser) inlineHTMLComment(data []byte) int {
	if len(data) < 5 {
		return 0
	}
	if data[0] != '<' || data[1] != '!' || data[2] != '-' || data[3] != '-' {
		return 0
	}
	i := 5
	// scan for an end-of-comment marker, across lines if necessary
	for i < len(data) && !(data[i-2] == '-' && data[i-1] == '-' && data[i] == '>') {
		i++
	}
	// no end-of-comment marker
	if i >= len(data) {
		return 0
	}
	return i + 1
}

func stripMailto(link []byte) []byte {
	if bytes.HasPrefix(link, []byte("mailto://")) {
		return link[9:]
	} else if bytes.HasPrefix(link, []byte("mailto:")) {
		return link[7:]
	} else {
		return link
	}
}

// autolinkType specifies a kind of autolink that gets detected.
type autolinkType int

// These are the possible flag values for the autolink renderer.
const (
	notAutolink autolinkType = iota
	normalAutolink
	emailAutolink
)

// '<' when tags or autolinks are allowed
func leftAngle(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]

	if p.extensions&Mmark != 0 {
		id, consumed := IsCallout(data)
		if consumed > 0 {
			node := &ast.Callout{}
			node.ID = id
			return consumed, node
		}
	}

	altype, end := tagLength(data)
	if size := p.inlineHTMLComment(data); size > 0 {
		end = size
	}
	if end <= 2 {
		return end, nil
	}
	if altype == notAutolink {
		htmlTag := &ast.HTMLSpan{}
		htmlTag.Literal = data[:end]
		return end, htmlTag
	}

	var uLink bytes.Buffer
	unescapeText(&uLink, data[1:end+1-2])
	if uLink.Len() <= 0 {
		return end, nil
	}
	link := uLink.Bytes()
	node := &ast.Link{
		Destination: link,
	}
	if altype == emailAutolink {
		node.Destination = append([]byte("mailto:"), link...)
	}
	ast.AppendChild(node, newTextNode(stripMailto(link)))
	return end, node
}

// '\\' backslash escape
var EscapeChars = []byte("\\`*_{}[]()#+-.!:|&<>~^$")

func escape(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]

	if len(data) <= 1 {
		return 2, nil
	}

	if p.extensions&NonBlockingSpace != 0 && data[1] == ' ' {
		return 2, &ast.NonBlockingSpace{}
	}

	if p.extensions&BackslashLineBreak != 0 && data[1] == '\n' {
		return 2, &ast.Hardbreak{}
	}

	if bytes.IndexByte(EscapeChars, data[1]) < 0 {
		return 0, nil
	}

	return 2, newTextNode(data[1:2])
}

func unescapeText(ob *bytes.Buffer, src []byte) {
	i := 0
	for i < len(src) {
		org := i
		for i < len(src) && src[i] != '\\' {
			i++
		}

		if i > org {
			ob.Write(src[org:i])
		}

		if i+1 >= len(src) {
			break
		}

		ob.WriteByte(src[i+1])
		i += 2
	}
}

// '&' escaped when it doesn't belong to an entity
// valid entities are assumed to be anything matching &#?[A-Za-z0-9]+;
func entity(p *Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]

	end := skipCharN(data, 1, '#', 1)
	end = skipAlnum(data, end)

	if end < len(data) && data[end] == ';' {
		end++ // real entity
	} else {
		return 0, nil // lone '&'
	}

	ent := data[:end]
	// undo &amp; escaping or it will be converted to &amp;amp; by another
	// escaper in the renderer
	if bytes.Equal(ent, []byte("&amp;")) {
		return end, newTextNode([]byte{'&'})
	}
	if len(ent) < 4 {
		return end, newTextNode(ent)
	}

	// if ent consists solely out of numbers (hex or decimal) convert that unicode codepoint to actual rune
	codepoint := uint64(0)
	var err error
	if ent[2] == 'x' || ent[2] == 'X' { // hexadecimal
		codepoint, err = strconv.ParseUint(string(ent[3:len(ent)-1]), 16, 64)
	} else {
		codepoint, err = strconv.ParseUint(string(ent[2:len(ent)-1]), 10, 64)
	}
	if err == nil { // only if conversion was valid return here.
		r := rune(codepoint)
		// Replace invalid codepoints with U+FFFD per CommonMark spec section 6.2
		if r == 0 || (r >= 0xD800 && r <= 0xDFFF) || r > 0x10FFFF {
			r = '\uFFFD'
		}
		return end, newTextNode([]byte(string(r)))
	}

	return end, newTextNode(ent)
}

func linkEndsWithEntity(data []byte, linkEnd int) bool {
	entityRanges := htmlEntityRe.FindAllIndex(data[:linkEnd], -1)
	return entityRanges != nil && entityRanges[len(entityRanges)-1][1] == linkEnd
}

// hasPrefixCaseInsensitive is a custom implementation of
//
//	strings.HasPrefix(strings.ToLower(s), prefix)
//
// we rolled our own because ToLower pulls in a huge machinery of lowercasing
// anything from Unicode and that's very slow. Since this func will only be
// used on ASCII protocol prefixes, we can take shortcuts.
func hasPrefixCaseInsensitive(s, prefix []byte) bool {
	if len(s) < len(prefix) {
		return false
	}
	delta := byte('a' - 'A')
	for i, b := range prefix {
		if b != s[i] && b != s[i]+delta {
			return false
		}
	}
	return true
}

var protocolPrefixes = [][]byte{
	[]byte("http://"),
	[]byte("https://"),
	[]byte("ftp://"),
	[]byte("file://"),
	[]byte("mailto:"),
}

const shortestPrefix = 6 // len("ftp://"), the shortest of the above

func maybeAutoLink(p *Parser, data []byte, offset int) (int, ast.Node) {
	// quick check to rule out most false hits
	if p.InsideLink || len(data) < offset+shortestPrefix {
		return 0, nil
	}
	// Handler is registered on h/m/f. Almost all hits are words like "the"/"from".
	c1 := data[offset+1] | 0x20
	switch data[offset] | 0x20 {
	case 'h':
		if c1 != 't' { // http(s)
			return 0, nil
		}
	case 'm':
		if c1 != 'a' { // mailto
			return 0, nil
		}
	case 'f':
		if c1 != 't' && c1 != 'i' { // ftp, file
			return 0, nil
		}
	}
	for _, prefix := range protocolPrefixes {
		endOfHead := offset + 8 // 8 is the len() of the longest prefix
		if endOfHead > len(data) {
			endOfHead = len(data)
		}
		if hasPrefixCaseInsensitive(data[offset:endOfHead], prefix) {
			return autoLink(p, data, offset)
		}
	}
	return 0, nil
}

func autoLink(p *Parser, data []byte, offset int) (int, ast.Node) {
	// Now a more expensive check to see if we're not inside an anchor element
	anchorStart := offset
	offsetFromAnchor := 0
	for anchorStart > 0 && data[anchorStart] != '<' {
		anchorStart--
		offsetFromAnchor++
	}

	anchorStr := anchorRe.Find(data[anchorStart:])
	if anchorStr != nil {
		anchorClose := &ast.HTMLSpan{}
		anchorClose.Literal = anchorStr[offsetFromAnchor:]
		return len(anchorStr) - offsetFromAnchor, anchorClose
	}

	// scan backward for a word boundary
	rewind := 0
	for offset-rewind > 0 && rewind <= 7 && IsLetter(data[offset-rewind-1]) {
		rewind++
	}
	if rewind > 6 { // longest supported protocol is "mailto" which has 6 letters
		return 0, nil
	}

	origData := data
	data = data[offset-rewind:]

	isSafeURL := p.IsSafeURLOverride
	if isSafeURL == nil {
		isSafeURL = IsSafeURL
	}
	if !isSafeURL(data) {
		return 0, nil
	}

	linkEnd := 0
	for linkEnd < len(data) && !isEndOfLink(data[linkEnd]) {
		linkEnd++
	}

	// Skip punctuation at the end of the link
	if (data[linkEnd-1] == '.' || data[linkEnd-1] == ',') && data[linkEnd-2] != '\\' {
		linkEnd--
	}

	// But don't skip semicolon if it's a part of escaped entity:
	if data[linkEnd-1] == ';' && data[linkEnd-2] != '\\' && !linkEndsWithEntity(data, linkEnd) {
		linkEnd--
	}

	// See if the link finishes with a punctuation sign that can be closed.
	var copen byte
	switch data[linkEnd-1] {
	case '"':
		copen = '"'
	case '\'':
		copen = '\''
	case ')':
		copen = '('
	case ']':
		copen = '['
	case '}':
		copen = '{'
	default:
		copen = 0
	}

	if copen != 0 {
		bufEnd := offset - rewind + linkEnd - 2

		openDelim := 1

		/* Try to close the final punctuation sign in this same line;
		 * if we managed to close it outside of the URL, that means that it's
		 * not part of the URL. If it closes inside the URL, that means it
		 * is part of the URL.
		 *
		 * Examples:
		 *
		 *      foo http://www.pokemon.com/Pikachu_(Electric) bar
		 *              => http://www.pokemon.com/Pikachu_(Electric)
		 *
		 *      foo (http://www.pokemon.com/Pikachu_(Electric)) bar
		 *              => http://www.pokemon.com/Pikachu_(Electric)
		 *
		 *      foo http://www.pokemon.com/Pikachu_(Electric)) bar
		 *              => http://www.pokemon.com/Pikachu_(Electric))
		 *
		 *      (foo http://www.pokemon.com/Pikachu_(Electric)) bar
		 *              => foo http://www.pokemon.com/Pikachu_(Electric)
		 */

		for bufEnd >= 0 && origData[bufEnd] != '\n' && openDelim != 0 {
			if origData[bufEnd] == data[linkEnd-1] {
				openDelim++
			}

			if origData[bufEnd] == copen {
				openDelim--
			}

			bufEnd--
		}

		if openDelim == 0 {
			linkEnd--
		}
	}

	var uLink bytes.Buffer
	unescapeText(&uLink, data[:linkEnd])

	if uLink.Len() > 0 {
		node := &ast.Link{
			Destination: uLink.Bytes(),
		}
		ast.AppendChild(node, newTextNode(uLink.Bytes()))
		return linkEnd, node
	}

	return linkEnd, nil
}

func isEndOfLink(char byte) bool {
	return IsSpace(char) || char == '<'
}

// return the length of the given tag, or 0 is it's not valid
func tagLength(data []byte) (autolink autolinkType, end int) {
	var i, j int

	// a valid tag can't be shorter than 3 chars
	if len(data) < 3 {
		return notAutolink, 0
	}

	// begins with a '<' optionally followed by '/', followed by letter or number
	if data[0] != '<' {
		return notAutolink, 0
	}
	if data[1] == '/' {
		i = 2
	} else {
		i = 1
	}

	if !IsAlnum(data[i]) {
		return notAutolink, 0
	}

	// scheme test
	autolink = notAutolink

	// try to find the beginning of an URI
	for i < len(data) && (IsAlnum(data[i]) || data[i] == '.' || data[i] == '+' || data[i] == '-') {
		i++
	}

	if i > 1 && i < len(data) && data[i] == '@' {
		if j = isMailtoAutoLink(data[i:]); j != 0 {
			return emailAutolink, i + j
		}
	}

	if i > 2 && i < len(data) && data[i] == ':' {
		autolink = normalAutolink
		i++
	}

	// complete autolink test: no whitespace or ' or "
	switch {
	case i >= len(data):
		autolink = notAutolink
	case autolink != notAutolink:
		j = i

		for i < len(data) {
			if data[i] == '\\' {
				i += 2
			} else if data[i] == '>' || data[i] == '\'' || data[i] == '"' || IsSpace(data[i]) {
				break
			} else {
				i++
			}

		}

		if i >= len(data) {
			return autolink, 0
		}
		if i > j && data[i] == '>' {
			return autolink, i + 1
		}

		// one of the forbidden chars has been found
		autolink = notAutolink
	}
	j = bytes.IndexByte(data[i:], '>')
	if j < 0 {
		return autolink, 0
	}
	i += j
	return autolink, i + 1
}

// look for the address part of a mail autolink and '>'
// this is less strict than the original markdown e-mail address matching
func isMailtoAutoLink(data []byte) int {
	nb := 0

	// address is assumed to be: [-@._a-zA-Z0-9]+ with exactly one '@'
	for i, c := range data {
		if IsAlnum(c) {
			continue
		}

		switch c {
		case '@':
			nb++

		case '-', '.', '_':
			// no-op but not defult

		case '>':
			if nb == 1 {
				return i + 1
			}
			return 0
		default:
			return 0
		}
	}

	return 0
}

// look for the next emph char, skipping other constructs
func helperFindEmphChar(data []byte, c byte) int {
	i := 0

	for i < len(data) {
		for i < len(data) && data[i] != c && data[i] != '`' && data[i] != '[' {
			i++
		}
		if i >= len(data) {
			return 0
		}
		// do not count escaped chars
		if i != 0 && data[i-1] == '\\' {
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

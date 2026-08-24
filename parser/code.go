package parser

import (
	"bytes"
	"html"

	"github.com/gomarkdown/markdown/ast"
)

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
// blockMath handle block surround with $$

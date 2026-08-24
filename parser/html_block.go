package parser

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

var (
	// blockTags is a set of tags that are recognized as HTML block tags.
	// Any of these can be included in markdown text without special escaping.
	blockTags = map[string]struct{}{
		"blockquote": {},
		"del":        {},
		"dd":         {},
		"div":        {},
		"dl":         {},
		"dt":         {},
		"fieldset":   {},
		"form":       {},
		"h1":         {},
		"h2":         {},
		"h3":         {},
		"h4":         {},
		"h5":         {},
		"h6":         {},
		// TODO: technically block but breaks Inline HTML (Simple).text
		//"hr":         {},
		"iframe":   {},
		"ins":      {},
		"li":       {},
		"math":     {},
		"noscript": {},
		"ol":       {},
		"pre":      {},
		"p":        {},
		"script":   {},
		"style":    {},
		"table":    {},
		"ul":       {},

		// HTML5
		"address":    {},
		"article":    {},
		"aside":      {},
		"canvas":     {},
		"details":    {},
		"dialog":     {},
		"figcaption": {},
		"figure":     {},
		"footer":     {},
		"header":     {},
		"hgroup":     {},
		"main":       {},
		"nav":        {},
		"output":     {},
		"progress":   {},
		"section":    {},
		"svg":        {},
		"video":      {},
	}

	markdownHTMLBlockTags = map[string]struct{}{
		"details": {},
		"div":     {},
	}

	// Tags whose interiors are walked as nested HTML when MarkdownInHTML is set.
	htmlStructureTags = map[string]struct{}{
		"table": {},
		"thead": {},
		"tbody": {},
		"tfoot": {},
		"tr":    {},
	}

	// Extra block tags recognized only with MarkdownInHTML.
	markdownInHTMLTags = map[string]struct{}{
		"table": {},
		"thead": {},
		"tbody": {},
		"tfoot": {},
		"tr":    {},
		"td":    {},
		"th":    {},
	}
)

func (p *Parser) html(data []byte, doRender bool) int {
	var i, j int

	// identify the opening tag
	if data[0] != '<' {
		return 0
	}
	curtag, tagfound := p.htmlFindTag(data[1:])

	// handle special cases
	if !tagfound {
		// check for an HTML comment
		if size := p.htmlComment(data, doRender); size > 0 {
			return size
		}

		// check for an <hr> tag
		if size := p.htmlHr(data, doRender); size > 0 {
			return size
		}

		// no special case recognized
		return 0
	}

	if p.markdownHTMLTag(curtag) {
		if p.extensions&MarkdownInHTML != 0 && isHTMLStructureTag(curtag) {
			if size := p.htmlStructuredBlock(data, curtag, doRender); size > 0 {
				return size
			}
		}
		if size := p.htmlMarkdownBlock(data, curtag, doRender); size > 0 {
			return size
		}
	}

	// look for an unindented matching closing tag
	// followed by a blank line
	found := false
	/*
		closetag := []byte("\n</" + curtag + ">")
		j = len(curtag) + 1
		for !found {
			// scan for a closing tag at the beginning of a line
			if skip := bytes.Index(data[j:], closetag); skip >= 0 {
				j += skip + len(closetag)
			} else {
				break
			}

			// see if it is the only thing on the line
			if skip := IsEmpty(data[j:]); skip > 0 {
				// see if it is followed by a blank line/eof
				j += skip
				if j >= len(data) {
					found = true
					i = j
				} else {
					if skip := IsEmpty(data[j:]); skip > 0 {
						j += skip
						found = true
						i = j
					}
				}
			}
		}
	*/

	// if not found, try a second pass looking for indented match
	// but not if tag is "ins" or "del" (following original Markdown.pl)
	if !found && curtag != "ins" && curtag != "del" {
		i = 1
		for i < len(data) {
			i++
			for i < len(data) && !(data[i-1] == '<' && data[i] == '/') {
				i++
			}

			if i+2+len(curtag) >= len(data) {
				break
			}

			j = p.htmlFindEnd(curtag, data[i-1:])

			if j > 0 {
				i += j - 1
				found = true
				break
			}
		}
	}

	if !found {
		return 0
	}

	// the end of the block has been found
	if doRender {
		// trim newlines
		end := backChar(data, i, '\n')
		htmlBLock := &ast.HTMLBlock{Leaf: ast.Leaf{Content: data[:end]}}
		p.AddBlock(htmlBLock)
		finalizeHTMLBlock(htmlBLock)
	}

	return i
}

func (p *Parser) markdownHTMLTag(tag string) bool {
	if _, ok := markdownHTMLBlockTags[tag]; ok {
		return true
	}
	if p.extensions&MarkdownInHTML != 0 {
		_, ok := markdownInHTMLTags[tag]
		return ok
	}
	return false
}

func isHTMLStructureTag(tag string) bool {
	_, ok := htmlStructureTags[tag]
	return ok
}

func (p *Parser) htmlMarkdownBlock(data []byte, tag string, doRender bool) int {
	openEnd := bytes.IndexByte(data, '>')
	if openEnd < 0 {
		return 0
	}
	openEnd++

	loose := p.extensions&MarkdownInHTML != 0
	closeStart, consumed := p.findHTMLCloseTag(data, tag, openEnd, loose)
	if consumed == 0 {
		return 0
	}
	if !hasBlankLineAfter(data, openEnd) || !hasBlankLineBefore(data, closeStart, loose) {
		return 0
	}

	closeEnd := closeStart + len("</"+tag+">")
	if doRender {
		open := bytes.TrimRight(data[:openEnd], "\n")
		p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: open}})

		innerStart := openEnd
		if innerStart < len(data) && data[innerStart] == '\n' {
			innerStart++
		}
		inner := data[innerStart:closeStart]
		if len(inner) > 0 {
			p.Block(inner)
		}

		close := bytes.TrimRight(data[closeStart:closeEnd], "\n")
		p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: close}})
	}

	return consumed
}

func hasBlankLineAfter(data []byte, pos int) bool {
	first := IsEmpty(data[pos:])
	if first == 0 {
		return false
	}
	return IsEmpty(data[pos+first:]) > 0
}

func hasBlankLineBefore(data []byte, pos int, skipIndent bool) bool {
	if skipIndent {
		for pos > 0 && (data[pos-1] == ' ' || data[pos-1] == '\t') {
			pos--
		}
	}
	if pos < 2 || data[pos-1] != '\n' {
		return false
	}
	lineEnd := pos - 1
	lineStart := lineEnd
	for lineStart > 0 && data[lineStart-1] != '\n' {
		lineStart--
	}
	for _, b := range data[lineStart:lineEnd] {
		if b != ' ' && b != '\t' {
			return false
		}
	}
	return true
}

func (p *Parser) findHTMLCloseTag(data []byte, tag string, start int, loose bool) (closeStart int, consumed int) {
	for i := start; i < len(data); i++ {
		for i < len(data) && !(data[i-1] == '<' && data[i] == '/') {
			i++
		}
		if i+2+len(tag) > len(data) {
			return 0, 0
		}

		var j int
		if loose {
			j = htmlFindEndLoose(tag, data[i-1:])
		} else {
			j = p.htmlFindEnd(tag, data[i-1:])
		}
		if j > 0 {
			return i - 1, i + j - 1
		}
	}
	return 0, 0
}

func htmlFindEndLoose(tag string, data []byte) int {
	closetag := []byte("</" + tag + ">")
	if !bytes.HasPrefix(data, closetag) {
		return 0
	}
	i := len(closetag)
	for i < len(data) && data[i] != '\n' {
		if data[i] != ' ' && data[i] != '\t' {
			return 0
		}
		i++
	}
	if i < len(data) && data[i] == '\n' {
		i++
	}
	return i
}

func (p *Parser) htmlStructuredBlock(data []byte, tag string, doRender bool) int {
	openEnd := bytes.IndexByte(data, '>')
	if openEnd < 0 {
		return 0
	}
	openEnd++

	closeStart, consumed := p.findHTMLCloseTag(data, tag, openEnd, true)
	if consumed == 0 {
		return 0
	}

	closeEnd := closeStart + len("</"+tag+">")
	if doRender {
		open := bytes.TrimRight(data[:openEnd], "\n")
		p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: open}})
		p.parseHTMLInterior(data[openEnd:closeStart])
		close := bytes.TrimRight(data[closeStart:closeEnd], "\n")
		p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: close}})
	}
	return consumed
}

func (p *Parser) parseHTMLInterior(data []byte) {
	for len(data) > 0 {
		if data[0] == '\n' {
			data = data[1:]
			continue
		}
		i := 0
		for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
			i++
		}
		if i < len(data) && data[i] == '<' {
			if n := p.html(data[i:], true); n > 0 {
				data = data[i+n:]
				continue
			}
		}
		nl := bytes.IndexByte(data, '\n')
		if nl < 0 {
			if len(bytes.TrimSpace(data)) > 0 {
				p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: bytes.TrimRight(data, "\n")}})
			}
			return
		}
		chunk := data[:nl]
		if len(bytes.TrimSpace(chunk)) > 0 {
			p.AddBlock(&ast.HTMLBlock{Leaf: ast.Leaf{Literal: chunk}})
		}
		data = data[nl+1:]
	}
}

func finalizeHTMLBlock(block *ast.HTMLBlock) {
	block.Literal = block.Content
	block.Content = nil
}

// HTML comment, lax form
func (p *Parser) htmlComment(data []byte, doRender bool) int {
	i := p.inlineHTMLComment(data)
	// needs to end with a blank line
	if j := IsEmpty(data[i:]); j > 0 {
		size := i + j
		if doRender {
			// trim trailing newlines
			end := backChar(data, size, '\n')
			htmlBLock := &ast.HTMLBlock{Leaf: ast.Leaf{Content: data[:end]}}
			p.AddBlock(htmlBLock)
			finalizeHTMLBlock(htmlBLock)
		}
		return size
	}
	return 0
}

// HR, which is the only self-closing block tag considered
func (p *Parser) htmlHr(data []byte, doRender bool) int {
	if len(data) < 4 {
		return 0
	}
	if data[0] != '<' || (data[1] != 'h' && data[1] != 'H') || (data[2] != 'r' && data[2] != 'R') {
		return 0
	}
	if data[3] != ' ' && data[3] != '/' && data[3] != '>' {
		// not an <hr> tag after all; at least not a valid one
		return 0
	}
	i := 3
	for i < len(data) && data[i] != '>' && data[i] != '\n' {
		i++
	}
	if i < len(data) && data[i] == '>' {
		i++
		if j := IsEmpty(data[i:]); j > 0 {
			size := i + j
			if doRender {
				// trim newlines
				end := backChar(data, size, '\n')
				htmlBlock := &ast.HTMLBlock{Leaf: ast.Leaf{Content: data[:end]}}
				p.AddBlock(htmlBlock)
				finalizeHTMLBlock(htmlBlock)
			}
			return size
		}
	}
	return 0
}

func (p *Parser) htmlFindTag(data []byte) (string, bool) {
	i := skipAlnum(data, 0)
	key := string(data[:i])
	if _, ok := blockTags[key]; ok {
		return key, true
	}
	if p.extensions&MarkdownInHTML != 0 {
		if _, ok := markdownInHTMLTags[key]; ok {
			return key, true
		}
	}
	return "", false
}

func (p *Parser) htmlFindEnd(tag string, data []byte) int {
	// assume data[0] == '<' && data[1] == '/' already tested
	if tag == "hr" {
		return 2
	}
	// check if tag is a match
	closetag := []byte("</" + tag + ">")
	if !bytes.HasPrefix(data, closetag) {
		return 0
	}
	i := len(closetag)

	// check that the rest of the line is blank
	skip := 0
	if skip = IsEmpty(data[i:]); skip == 0 {
		return 0
	}
	i += skip
	skip = 0

	if i >= len(data) {
		return i
	}

	if p.extensions&LaxHTMLBlocks != 0 {
		return i
	}
	if skip = IsEmpty(data[i:]); skip == 0 {
		// following line must be blank
		return 0
	}

	return i + skip
}

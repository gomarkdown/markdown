package parser

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

// quotePrefix returns the blockquote prefix length.
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

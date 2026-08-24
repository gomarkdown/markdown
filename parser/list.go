package parser

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

func (p *Parser) uliPrefix(data []byte) int {
	// start with up to 3 spaces
	i := skipCharN(data, 0, ' ', 3)

	if i >= len(data)-1 {
		return 0
	}
	// need one of {'*', '+', '-'} followed by a space or a tab
	if (data[i] != '*' && data[i] != '+' && data[i] != '-') ||
		(data[i+1] != ' ' && data[i+1] != '\t') {
		return 0
	}
	return i + 2
}

// returns ordered list item prefix
func (p *Parser) oliPrefix(data []byte) int {
	// start with up to 3 spaces
	i := skipCharN(data, 0, ' ', 3)

	// count the digits
	start := i
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		i++
	}
	if start == i || i >= len(data)-1 {
		return 0
	}

	// we need >= 1 digits followed by a dot and a space or a tab
	if data[i] != '.' && data[i] != ')' || !(data[i+1] == ' ' || data[i+1] == '\t') {
		return 0
	}
	return i + 2
}

// returns definition list item prefix
func (p *Parser) dliPrefix(data []byte) int {
	if len(data) < 2 {
		return 0
	}
	if data[0] != ':' || (data[1] != ' ' && data[1] != '\t') {
		return 0
	}
	return 2
}

// parse ordered or unordered list block
func (p *Parser) list(data []byte, flags ast.ListType, start int, delim byte) int {
	i := 0
	flags |= ast.ListItemBeginningOfList
	list := &ast.List{
		ListFlags: flags,
		Tight:     true,
		Start:     start,
		Delimiter: delim,
	}
	block := p.AddBlock(list)

	for i < len(data) {
		skip := p.listItem(data[i:], &flags)
		if flags&ast.ListItemContainsBlock != 0 {
			list.Tight = false
		}
		i += skip
		if skip == 0 || flags&ast.ListItemEndOfList != 0 {
			break
		}
		flags &= ^ast.ListItemBeginningOfList
	}

	above := block.GetParent()
	finalizeList(list)
	p.tip = above
	return i
}

// Returns true if the list item is not the same type as its parent list
func (p *Parser) listTypeChanged(data []byte, flags *ast.ListType) bool {
	if p.dliPrefix(data) > 0 && *flags&ast.ListTypeDefinition == 0 {
		return true
	} else if p.oliPrefix(data) > 0 && *flags&ast.ListTypeOrdered == 0 {
		return true
	} else if p.uliPrefix(data) > 0 && (*flags&ast.ListTypeOrdered != 0 || *flags&ast.ListTypeDefinition != 0) {
		return true
	}
	return false
}

// Returns true if block ends with a blank line, descending if needed
// into lists and sublists.
func endsWithBlankLine(block ast.Node) bool {
	// TODO: figure this out. Always false now.
	for block != nil {
		//if block.lastLineBlank {
		//return true
		//}
		switch block.(type) {
		case *ast.List, *ast.ListItem:
			block = ast.GetLastChild(block)
		default:
			return false
		}
	}
	return false
}

func finalizeList(list *ast.List) {
	items := list.GetChildren()
	lastItemIdx := len(items) - 1
	for i, item := range items {
		isLastItem := i == lastItemIdx
		// check for non-final list item ending with blank line:
		if !isLastItem && endsWithBlankLine(item) {
			list.Tight = false
			break
		}
		// recurse into children of list item, to see if there are spaces
		// between any of them:
		subItems := item.GetChildren()
		lastSubItemIdx := len(subItems) - 1
		for j, subItem := range subItems {
			isLastSubItem := j == lastSubItemIdx
			if (!isLastItem || !isLastSubItem) && endsWithBlankLine(subItem) {
				list.Tight = false
				break
			}
		}
	}
}

// Parse a single list item.
// Assumes initial prefix is already removed if this is a sublist.
func (p *Parser) listItem(data []byte, flags *ast.ListType) int {
	isDefinitionList := *flags&ast.ListTypeDefinition != 0
	// keep track of the indentation of the first line
	itemIndent := 0
	if data[0] == '\t' {
		itemIndent += 4
	} else {
		for itemIndent < 3 && data[itemIndent] == ' ' {
			itemIndent++
		}
	}

	var (
		bulletChar byte = '*'
		delimiter  byte = '.'
	)
	i := p.uliPrefix(data)
	if i == 0 {
		i = p.oliPrefix(data)
		if i > 0 {
			delimiter = data[i-2]
		}
	} else {
		bulletChar = data[i-2]
	}
	if i == 0 {
		i = p.dliPrefix(data)
		// reset definition term flag
		if i > 0 {
			*flags &= ^ast.ListTypeTerm
		}
	}
	if i == 0 {
		// if in definition list, set term flag and continue
		if isDefinitionList {
			*flags |= ast.ListTypeTerm
		} else {
			return 0
		}
	}

	// skip leading whitespace on first line
	i = skipChar(data, i, ' ')

	// find the end of the line
	line := i
	for i > 0 && i < len(data) && data[i-1] != '\n' {
		i++
	}

	// get working buffer
	var raw bytes.Buffer

	// put the first line into the working buffer
	raw.Write(data[line:i])
	line = i

	// process the following lines
	containsBlankLine := false
	sublist := 0
	// track fenced code blocks inside list items so that lines within
	// the fence are gathered verbatim (not misinterpreted as list items)
	fenceMarker := ""

gatherlines:
	for line < len(data) {
		i++

		// find the end of this line
		for i < len(data) && data[i-1] != '\n' {
			i++
		}

		// if it is an empty line, guess that it is part of this item
		// and move on to the next line
		if IsEmpty(data[line:i]) > 0 {
			containsBlankLine = true
			line = i
			continue
		}

		// calculate the indentation
		indent := 0
		indentIndex := 0
		if data[line] == '\t' {
			indentIndex++
			indent += 4
		} else {
			for indent < 4 && line+indent < i && data[line+indent] == ' ' {
				indent++
				indentIndex++
			}
		}

		chunk := data[line+indentIndex : i]

		// track fenced code blocks inside list items;
		// only track fences that are indented (part of the list item content),
		// a fence at indent 0 ends the list (handled below)
		if !isDefinitionList && p.extensions&FencedCode != 0 {
			if fenceMarker != "" {
				if indent == 0 {
					// non-indented line while inside a fence means we
					// left the list item content -- abandon the fence
					fenceMarker = ""
				} else {
					// inside a fence: check for closing fence
					_, marker := isFenceLine(chunk, nil, fenceMarker)
					if marker != "" {
						fenceMarker = ""
					}
					// gather the line verbatim, skip structure detection
					if containsBlankLine {
						containsBlankLine = false
						raw.WriteByte('\n')
					}
					raw.Write(chunk)
					line = i
					continue
				}
			} else if indent > 0 {
				// not inside a fence: check for opening fence (indented only)
				_, marker := isFenceLine(chunk, nil, "")
				if marker != "" {
					fenceMarker = marker
				}
			}
		}

		// If there is a fence line (marking starting of a code block)
		// without indent do not process it as part of the list.
		//
		// does not apply for definition lists because it causes infinite
		// loop if text before defintion term is fenced code block start
		// marker but not part of actual fenced code block
		// for defnition lists we're called after parsing fence code blocks
		// so we kno this cannot be a fenced block
		// https://github.com/gomarkdown/markdown/issues/326
		if !isDefinitionList && p.extensions&FencedCode != 0 {
			fenceLineEnd, _ := isFenceLine(chunk, nil, "")
			if fenceLineEnd > 0 && indent == 0 {
				*flags |= ast.ListItemEndOfList
				break gatherlines
			}
		}

		// evaluate how this line fits in
		switch {
		// is this a nested list item?
		case (p.uliPrefix(chunk) > 0 && !isHRule(chunk)) || p.oliPrefix(chunk) > 0 || p.dliPrefix(chunk) > 0:

			// if indent is 4 or more spaces on unordered or ordered lists
			// we need to add leadingWhiteSpaces + 1 spaces in the beginning of the chunk
			if indentIndex >= 4 && p.dliPrefix(chunk) <= 0 {
				leadingWhiteSpaces := skipChar(chunk, 0, ' ')
				chunk = data[line+indentIndex-(leadingWhiteSpaces+1) : i]
			}

			// to be a nested list, it must be indented more
			// if not, it is either a different kind of list
			// or the next item in the same list
			if indent <= itemIndent {
				if p.listTypeChanged(chunk, flags) {
					*flags |= ast.ListItemEndOfList
				} else if containsBlankLine {
					*flags |= ast.ListItemContainsBlock
				}

				break gatherlines
			}

			if containsBlankLine {
				*flags |= ast.ListItemContainsBlock
			}

			// is this the first item in the nested list?
			if sublist == 0 {
				sublist = raw.Len()
				// in the case of dliPrefix we are too late and need to search back for the definition item, which
				// should be on the previous line, we then adjust sublist to start there.
				if p.dliPrefix(chunk) > 0 {
					sublist = backUntilChar(raw.Bytes(), raw.Len()-1, '\n')
				}
			}

			// is this a nested prefix heading?
		case p.isPrefixHeading(chunk), p.isPrefixSpecialHeading(chunk):
			// if the heading is not indented, it is not nested in the list
			// and thus ends the list
			if containsBlankLine && indent < 4 {
				*flags |= ast.ListItemEndOfList
				break gatherlines
			}
			*flags |= ast.ListItemContainsBlock

		case p.quotePrefix(chunk) > 0 && indent < 4:
			*flags |= ast.ListItemEndOfList
			break gatherlines

		// anything following an empty line is only part
		// of this item if it is indented 4 spaces
		// (regardless of the indentation of the beginning of the item)
		case containsBlankLine && indent < 4:
			if *flags&ast.ListTypeDefinition != 0 && i < len(data)-1 {
				// is the next item still a part of this list?
				next := skipUntilChar(data, i, '\n')
				for next < len(data)-1 && data[next] == '\n' {
					next++
				}
				if i < len(data)-1 && data[i] != ':' && next < len(data)-1 && data[next] != ':' {
					*flags |= ast.ListItemEndOfList
				}
			} else {
				*flags |= ast.ListItemEndOfList
			}
			break gatherlines

		// a blank line means this should be parsed as a block
		case containsBlankLine:
			raw.WriteByte('\n')
			*flags |= ast.ListItemContainsBlock
		}

		// if this line was preceded by one or more blanks,
		// re-introduce the blank into the buffer
		if containsBlankLine {
			containsBlankLine = false
			raw.WriteByte('\n')
		}

		// add the line into the working buffer without prefix
		raw.Write(chunk)

		line = i
	}

	rawBytes := raw.Bytes()

	listItem := &ast.ListItem{
		ListFlags:  *flags,
		Tight:      false,
		BulletChar: bulletChar,
		Delimiter:  delimiter,
	}
	p.AddBlock(listItem)

	// render the contents of the list item
	if *flags&ast.ListItemContainsBlock != 0 && *flags&ast.ListTypeTerm == 0 {
		// intermediate render of block item, except for definition term
		if sublist > 0 {
			p.Block(rawBytes[:sublist])
			p.Block(rawBytes[sublist:])
		} else {
			p.Block(rawBytes)
		}
	} else {
		// intermediate render of inline item
		para := &ast.Paragraph{}
		if sublist > 0 {
			para.Content = rawBytes[:sublist]
		} else {
			para.Content = rawBytes
		}
		p.addChild(para)
		if sublist > 0 {
			p.Block(rawBytes[sublist:])
		}
	}
	return line
}

// render a single paragraph that has already been parsed out

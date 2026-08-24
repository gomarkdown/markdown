package parser

import (
	"bytes"
	"strconv"

	"github.com/gomarkdown/markdown/ast"
)

type linkType int

const (
	linkNormal linkType = iota
	linkImg
	linkDeferredFootnote
	linkInlineFootnote
	linkCitation
)

func isReferenceStyleLink(data []byte, pos int, t linkType) bool {
	if t == linkDeferredFootnote {
		return false
	}
	return pos < len(data)-1 && data[pos] == '[' && data[pos+1] != '^'
}

func link(p *Parser, data []byte, offset int) (int, ast.Node) {
	// no links allowed inside regular links, footnote, and deferred footnotes
	if p.InsideLink && (offset > 0 && data[offset-1] == '[' || len(data)-1 > offset && data[offset+1] == '^') {
		return 0, nil
	}

	var t linkType
	switch {
	// special case: ![^text] == deferred footnote (that follows something with
	// an exclamation point)
	case p.extensions&Footnotes != 0 && len(data)-1 > offset && data[offset+1] == '^':
		t = linkDeferredFootnote
	// ![alt] == image
	case offset >= 0 && data[offset] == '!':
		t = linkImg
		offset++
	// [@citation], [@-citation], [@?citation], [@!citation]
	case p.extensions&Mmark != 0 && len(data)-1 > offset && data[offset+1] == '@':
		t = linkCitation
	// [text] == regular link
	// ^[text] == inline footnote
	// [^refId] == deferred footnote
	case p.extensions&Footnotes != 0:
		if offset >= 0 && data[offset] == '^' {
			t = linkInlineFootnote
			offset++
		} else if len(data)-1 > offset && data[offset+1] == '^' {
			t = linkDeferredFootnote
		}
	default:
		t = linkNormal
	}

	open := offset
	data = data[offset:]

	if t == linkCitation {
		return citation(p, data, 0)
	}

	var (
		i                               int
		noteID                          int
		title, link, linkID, altContent []byte
		textHasNl                       = false
	)

	// Match ']' from the Inline() buffer table so a run of unmatched '[' is
	// O(n) instead of O(n²) (GHSA-85vw-wvf9-r522).
	closeAt, nested, found := p.brackets.lookup(open)
	if !found {
		return 0, nil
	}
	txtE := closeAt - open
	if txtE < 1 || txtE >= len(data) {
		return 0, nil
	}
	textHasNl = bytes.IndexByte(data[1:txtE], '\n') >= 0
	i = txtE + 1
	var footnoteNode ast.Node

	// skip any amount of whitespace or newline
	// (this is much more lax than original markdown syntax)
	i = skipSpace(data, i)

	// inline style link
	switch {
	case i < len(data) && data[i] == '(':
		// skip initial whitespace
		i++

		i = skipSpace(data, i)

		linkB := i
		brace := 0

		var c byte
		// look for link end: ' " )
	findlinkend:
		for i < len(data) {
			c = data[i]
			switch {
			case c == '\\':
				i += 2

			case c == '(':
				brace++
				i++

			case c == ')':
				if brace <= 0 {
					break findlinkend
				}
				brace--
				i++

			case brace == 0 && (c == '\'' || c == '"') && i > linkB && IsSpace(data[i-1]):
				break findlinkend

			default:
				i++
			}
		}

		if i >= len(data) {
			return 0, nil
		}
		linkE := i

		// look for title end if present
		titleB, titleE := 0, 0
		if data[i] == '\'' || data[i] == '"' {
			i++
			titleB = i
			titleEndCharFound := false

		findtitleend:
			for i < len(data) {
				c = data[i]
				switch {
				case c == '\\':
					i++

				case c == data[titleB-1]: // matching title delimiter
					titleEndCharFound = true

				case titleEndCharFound && c == ')':
					break findtitleend
				}
				i++
			}

			if i >= len(data) {
				return 0, nil
			}

			// skip whitespace after title
			titleE = i - 1
			for titleE > titleB && IsSpace(data[titleE]) {
				titleE--
			}

			// check for closing quote presence
			if data[titleE] != '\'' && data[titleE] != '"' {
				titleB, titleE = 0, 0
				linkE = i
			}
		}

		// remove whitespace at the end of the link
		for linkE > linkB && IsSpace(data[linkE-1]) {
			linkE--
		}

		// remove optional angle brackets around the link
		if data[linkB] == '<' {
			linkB++
		}
		if data[linkE-1] == '>' {
			linkE--
		}

		// build escaped link and title
		if linkE > linkB {
			link = data[linkB:linkE]
		}

		if titleE > titleB {
			title = data[titleB:titleE]
		}

		i++

	// reference style link
	case isReferenceStyleLink(data, i, t):
		var id []byte
		altContentConsidered := false

		// look for the id
		i++
		linkB := i
		i = skipUntilChar(data, i, ']')

		if i >= len(data) {
			return 0, nil
		}
		linkE := i

		// find the reference
		if linkB == linkE {
			if textHasNl {
				var b bytes.Buffer

				for j := 1; j < txtE; j++ {
					switch {
					case data[j] != '\n':
						b.WriteByte(data[j])
					case data[j-1] != ' ':
						b.WriteByte(' ')
					}
				}

				id = b.Bytes()
			} else {
				id = data[1:txtE]
				altContentConsidered = true
			}
		} else {
			id = data[linkB:linkE]
		}

		// find the reference with matching id
		lr, ok := p.getRef(string(id))
		if !ok {
			return 0, nil
		}

		// keep link and title from reference
		linkID = id
		link = lr.link
		title = lr.title
		if altContentConsidered {
			altContent = lr.text
		}
		i++

	// shortcut reference style link or reference or inline footnote
	default:
		var id []byte

		// craft the id
		if textHasNl {
			var b bytes.Buffer

			for j := 1; j < txtE; j++ {
				switch {
				case data[j] != '\n':
					b.WriteByte(data[j])
				case data[j-1] != ' ':
					b.WriteByte(' ')
				}
			}

			id = b.Bytes()
		} else {
			if t == linkDeferredFootnote {
				id = data[2:txtE] // get rid of the ^
			} else {
				id = data[1:txtE]
			}
		}

		footnoteNode = &ast.ListItem{}
		if t == linkInlineFootnote {
			// create a new reference
			noteID = len(p.notes) + 1

			var fragment []byte
			if len(id) > 0 {
				if len(id) < 16 {
					fragment = make([]byte, len(id))
				} else {
					fragment = make([]byte, 16)
				}
				copy(fragment, slugify(id))
			} else {
				fragment = append([]byte("footnote-"), []byte(strconv.Itoa(noteID))...)
			}

			ref := &reference{
				noteID:   noteID,
				hasBlock: false,
				link:     fragment,
				title:    id,
				footnote: footnoteNode,
			}

			p.notes = append(p.notes, ref)
			p.refsRecord[string(ref.link)] = struct{}{}

			link = ref.link
			title = ref.title
		} else {
			// Nested [...] cannot be a stored reference label (labels end at
			// the first ']'). Skip the map lookup: it cannot match, and doing
			// it for every nested '[' is quadratic (GHSA-85vw-wvf9-r522).
			// Still consult ReferenceOverride, which sees the full inner text.
			if nested && p.ReferenceOverride == nil {
				return 0, nil
			}
			// find the reference with matching id
			lr, ok := p.getRef(string(id))
			if !ok {
				return 0, nil
			}

			if t == linkDeferredFootnote && !p.isFootnote(lr) {
				lr.noteID = len(p.notes) + 1
				lr.footnote = footnoteNode
				p.notes = append(p.notes, lr)
				p.refsRecord[string(lr.link)] = struct{}{}
			}

			// keep link and title from reference
			link = lr.link
			// if inline footnote, title == footnote contents
			title = lr.title
			noteID = lr.noteID
			if len(lr.text) > 0 {
				altContent = lr.text
			}
			if len(linkID) == 0 && noteID == 0 {
				linkID = id
			}
		}

		// rewind the whitespace
		i = txtE + 1
	}

	var uLink []byte
	if t == linkNormal || t == linkImg {
		if len(link) > 0 {
			var uLinkBuf bytes.Buffer
			unescapeText(&uLinkBuf, link)
			uLink = uLinkBuf.Bytes()
		}

		// links need something to click on and somewhere to go
		// [](http://bla) is legal in CommonMark, so allow txtE <=1 for linkNormal
		// [bla]() is also legal in CommonMark, so allow empty uLink
	}

	var inlineAttr *ast.Attribute
	if p.extensions&InlineAttributes != 0 && (t == linkNormal || t == linkImg) {
		if attr, n := parseAttributeList(data[i:], false); n > 0 {
			inlineAttr = attr
			i += n
		}
	}

	// call the relevant rendering function
	switch t {
	case linkNormal:
		link := &ast.Link{
			Destination: normalizeURI(uLink),
			Title:       title,
			DeferredID:  linkID,
		}
		applyAttribute(link, inlineAttr)
		if len(altContent) > 0 {
			ast.AppendChild(link, newTextNode(altContent))
		} else {
			// links cannot contain other links, so turn off link parsing
			// temporarily and recurse
			InsideLink := p.InsideLink
			p.InsideLink = true
			p.Inline(link, data[1:txtE])
			p.InsideLink = InsideLink
		}
		return i, link

	case linkImg:
		image := &ast.Image{
			Destination: uLink,
			Title:       title,
		}
		applyAttribute(image, inlineAttr)
		ast.AppendChild(image, newTextNode(data[1:txtE]))
		return i + 1, image

	case linkInlineFootnote, linkDeferredFootnote:
		link := &ast.Link{
			Destination: link,
			Title:       title,
			NoteID:      noteID,
			Footnote:    footnoteNode,
		}
		if t == linkDeferredFootnote {
			link.DeferredID = data[2:txtE]
		}
		if t == linkInlineFootnote {
			i++
		}
		return i, link

	default:
		return 0, nil
	}
}

func normalizeURI(s []byte) []byte {
	return s // TODO: implement
}

func slugify(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	out := make([]byte, 0, len(in))
	sym := false

	for _, ch := range in {
		if IsAlnum(ch) {
			sym = false
			out = append(out, ch)
		} else if sym {
			continue
		} else {
			out = append(out, '-')
			sym = true
		}
	}
	var a, b int
	var ch byte
	for a, ch = range out {
		if ch != '-' {
			break
		}
	}
	for b = len(out) - 1; b > 0; b-- {
		if out[b] != '-' {
			break
		}
	}
	return out[a : b+1]
}

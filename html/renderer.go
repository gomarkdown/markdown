package html

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func (r *Renderer) OutTag(w io.Writer, name string, attrs []string) {
	s := name
	if len(attrs) > 0 {
		s += " " + strings.Join(attrs, " ")
	}
	io.WriteString(w, s+">")
	r.lastOutputLen = 1
}

func FootnoteRef(prefix string, node *ast.Link) string {
	urlFrag := prefix + string(Slugify(node.Destination))
	nStr := strconv.Itoa(node.NoteID)
	anchor := `<a href="#fn:` + urlFrag + `">` + nStr + `</a>`
	return `<sup class="footnote-ref" id="fnref:` + urlFrag + `">` + anchor + `</sup>`
}

func FootnoteItem(prefix string, slug []byte) string {
	return `<li id="fn:` + prefix + string(slug) + `">`
}

func FootnoteReturnLink(prefix, returnLink string, slug []byte) string {
	return ` <a class="footnote-return" href="#fnref:` + prefix + string(slug) + `">` + returnLink + `</a>`
}

func ListItemOpenCR(listItem *ast.ListItem) bool {
	if ast.GetPrevNode(listItem) == nil {
		return false
	}
	ld := listItem.Parent.(*ast.List)
	return !ld.Tight && ld.ListFlags&ast.ListTypeDefinition == 0
}

func SkipParagraphTags(para *ast.Paragraph) bool {
	parent := para.Parent
	grandparent := parent.GetParent()
	if grandparent == nil || !IsList(grandparent) {
		return false
	}
	isParentTerm := IsListItemTerm(parent)
	grandparentListData := grandparent.(*ast.List)
	tightOrTerm := grandparentListData.Tight || isParentTerm
	return tightOrTerm
}

// Out is a helper to write data to writer
func (r *Renderer) Out(w io.Writer, d []byte) {
	r.lastOutputLen = len(d)
	if r.DisableTags > 0 {
		d = htmlTagRe.ReplaceAll(d, []byte{})
	}
	w.Write(d)
}

// Outs is a helper to write data to writer
func (r *Renderer) Outs(w io.Writer, s string) {
	r.lastOutputLen = len(s)
	if r.DisableTags > 0 {
		s = htmlTagRe.ReplaceAllString(s, "")
	}
	io.WriteString(w, s)
}

// CR writes a new line
func (r *Renderer) CR(w io.Writer) {
	if r.lastOutputLen > 0 {
		r.Outs(w, "\n")
	}
}

var (
	openHTags  = []string{"<h1", "<h2", "<h3", "<h4", "<h5"}
	closeHTags = []string{"</h1>", "</h2>", "</h3>", "</h4>", "</h5>"}
)

func HeadingOpenTagFromLevel(level int) string {
	if level < 1 || level > 5 {
		return "<h6"
	}
	return openHTags[level-1]
}

func HeadingCloseTagFromLevel(level int) string {
	if level < 1 || level > 5 {
		return "</h6>"
	}
	return closeHTags[level-1]
}

func (r *Renderer) OutHRTag(w io.Writer, attrs []string) {
	hr := TagWithAttributes("<hr", attrs)
	r.OutOneOf(w, r.Opts.Flags&UseXHTML == 0, hr, "<hr />")
}

// Text writes ast.Text node
func (r *Renderer) Text(w io.Writer, text *ast.Text) {
	if r.Opts.Flags&Smartypants != 0 {
		var tmp bytes.Buffer
		EscapeHTML(&tmp, text.Literal)
		r.sr.Process(w, tmp.Bytes())
	} else {
		_, parentIsLink := text.Parent.(*ast.Link)
		if parentIsLink {
			EscLink(w, text.Literal)
		} else {
			EscapeHTML(w, text.Literal)
		}
	}
}

// HardBreak writes ast.Hardbreak node
func (r *Renderer) HardBreak(w io.Writer, node *ast.Hardbreak) {
	r.OutOneOf(w, r.Opts.Flags&UseXHTML == 0, "<br>", "<br />")
	r.CR(w)
}

// NonBlockingSpace writes ast.NonBlockingSpace node
func (r *Renderer) NonBlockingSpace(w io.Writer, node *ast.NonBlockingSpace) {
	r.Outs(w, "&nbsp;")
}

// OutOneOf writes first or second depending on outFirst
func (r *Renderer) OutOneOf(w io.Writer, outFirst bool, first string, second string) {
	if outFirst {
		r.Outs(w, first)
	} else {
		r.Outs(w, second)
	}
}

// OutOneOfCr writes CR + first or second + CR depending on outFirst
func (r *Renderer) OutOneOfCr(w io.Writer, outFirst bool, first string, second string) {
	if outFirst {
		r.CR(w)
		r.Outs(w, first)
	} else {
		r.Outs(w, second)
		r.CR(w)
	}
}

// HTMLSpan writes ast.HTMLSpan node
func (r *Renderer) HTMLSpan(w io.Writer, span *ast.HTMLSpan) {
	if r.Opts.Flags&SkipHTML == 0 {
		r.Out(w, span.Literal)
	}
}

func (r *Renderer) linkEnter(w io.Writer, link *ast.Link) {
	attrs := link.AdditionalAttributes
	dest := link.Destination
	dest = AddAbsPrefix(dest, r.Opts.AbsolutePrefix)
	var hrefBuf bytes.Buffer
	hrefBuf.WriteString("href=\"")
	EscLink(&hrefBuf, dest)
	hrefBuf.WriteByte('"')
	attrs = append(attrs, hrefBuf.String())
	if link.NoteID != 0 {
		r.Outs(w, FootnoteRef(r.Opts.FootnoteAnchorPrefix, link))
		return
	}

	attrs = appendLinkAttrs(attrs, r.Opts.Flags, dest)
	if len(link.Title) > 0 {
		var titleBuff bytes.Buffer
		titleBuff.WriteString("title=\"")
		EscapeHTML(&titleBuff, link.Title)
		titleBuff.WriteByte('"')
		attrs = append(attrs, titleBuff.String())
	}
	attrs = append(attrs, BlockAttrs(link)...)
	attrs = coalesceClassAttrs(attrs)
	r.OutTag(w, "<a", attrs)
}

func (r *Renderer) linkExit(w io.Writer, link *ast.Link) {
	if link.NoteID == 0 {
		r.Outs(w, "</a>")
	}
}

// Link writes ast.Link node
func (r *Renderer) Link(w io.Writer, link *ast.Link, entering bool) {
	// mark it but don't link it if it is not a safe link: no smartypants
	if needSkipLink(r, link.Destination) {
		r.OutOneOf(w, entering, "<tt>", "</tt>")
		return
	}

	if entering {
		r.linkEnter(w, link)
	} else {
		r.linkExit(w, link)
	}
}

func (r *Renderer) imageEnter(w io.Writer, image *ast.Image) {
	r.DisableTags++
	if r.DisableTags > 1 {
		return
	}
	src := image.Destination
	src = AddAbsPrefixToImage(src, r.Opts.AbsolutePrefix)
	attrs := BlockAttrs(image)
	if r.Opts.Flags&LazyLoadImages != 0 {
		attrs = append(attrs, `loading="lazy"`)
	}

	s := TagWithAttributes("<img", attrs)
	s = s[:len(s)-1] // hackish: strip off ">" from end
	r.Outs(w, s+` src="`)
	EscLink(w, src)
	r.Outs(w, `" alt="`)
}

func (r *Renderer) imageExit(w io.Writer, image *ast.Image) {
	r.DisableTags--
	if r.DisableTags > 0 {
		return
	}
	if image.Title != nil {
		r.Outs(w, `" title="`)
		EscapeHTML(w, image.Title)
	}
	r.Outs(w, `" />`)
}

// Image writes ast.Image node
func (r *Renderer) Image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		r.imageEnter(w, node)
	} else {
		r.imageExit(w, node)
	}
}

// prevVisibleBlock skips siblings that emit no HTML, such as
// reference definitions, so inter-block spacing stays stable.
func prevVisibleBlock(n ast.Node) ast.Node {
	prev := ast.GetPrevNode(n)
	for prev != nil {
		if _, ok := prev.(*ast.ReferenceDefinition); ok {
			prev = ast.GetPrevNode(prev)
			continue
		}
		return prev
	}
	return nil
}

func (r *Renderer) paragraphEnter(w io.Writer, para *ast.Paragraph) {
	// TODO: untangle this clusterfuck about when the newlines need
	// to be added and when not.
	prev := prevVisibleBlock(para)
	if prev != nil {
		switch prev.(type) {
		case *ast.HTMLBlock, *ast.List, *ast.Paragraph, *ast.Heading, *ast.CaptionFigure, *ast.CodeBlock, *ast.BlockQuote, *ast.Aside, *ast.HorizontalRule:
			r.CR(w)
		}
	}

	if prev == nil {
		_, isParentBlockQuote := para.Parent.(*ast.BlockQuote)
		if isParentBlockQuote {
			r.CR(w)
		}
		_, isParentAside := para.Parent.(*ast.Aside)
		if isParentAside {
			r.CR(w)
		}
	}

	ptag := "<p"
	if r.Opts.ParagraphTag != "" {
		ptag = "<" + r.Opts.ParagraphTag
	}
	tag := TagWithAttributes(ptag, BlockAttrs(para))
	r.Outs(w, tag)
}

func (r *Renderer) paragraphExit(w io.Writer, para *ast.Paragraph) {
	ptag := "</p>"
	if r.Opts.ParagraphTag != "" {
		ptag = "</" + r.Opts.ParagraphTag + ">"
	}
	r.Outs(w, ptag)
	if !(IsListItem(para.Parent) && ast.GetNextNode(para) == nil) {
		r.CR(w)
	}
}

// Paragraph writes ast.Paragraph node
func (r *Renderer) Paragraph(w io.Writer, para *ast.Paragraph, entering bool) {
	if SkipParagraphTags(para) {
		return
	}
	if entering {
		r.paragraphEnter(w, para)
	} else {
		r.paragraphExit(w, para)
	}
}

// Code writes ast.Code node
func (r *Renderer) Code(w io.Writer, node *ast.Code) {
	r.Outs(w, "<code>")
	EscapeHTML(w, node.Literal)
	r.Outs(w, "</code>")
}

// HTMLBlock write ast.HTMLBlock node
func (r *Renderer) HTMLBlock(w io.Writer, node *ast.HTMLBlock) {
	if r.Opts.Flags&SkipHTML != 0 {
		return
	}
	r.CR(w)
	r.Out(w, node.Literal)
	r.CR(w)
}

func (r *Renderer) EnsureUniqueHeadingID(id string) string {
	for count, found := r.headingIDs[id]; found; count, found = r.headingIDs[id] {
		tmp := fmt.Sprintf("%s-%d", id, count+1)

		if _, tmpFound := r.headingIDs[tmp]; !tmpFound {
			r.headingIDs[id] = count + 1
			id = tmp
		} else {
			id = id + "-1"
		}
	}

	if _, found := r.headingIDs[id]; !found {
		r.headingIDs[id] = 0
	}

	return id
}

func (r *Renderer) MakeUniqueHeadingID(hdr *ast.Heading) string {
	if hdr.HeadingID == "" {
		return ""
	}
	id := r.EnsureUniqueHeadingID(hdr.HeadingID)
	if r.Opts.HeadingIDPrefix != "" {
		id = r.Opts.HeadingIDPrefix + id
	}
	if r.Opts.HeadingIDSuffix != "" {
		id = id + r.Opts.HeadingIDSuffix
	}
	hdr.HeadingID = id
	return id
}

func (r *Renderer) HeadingEnter(w io.Writer, hdr *ast.Heading) {
	var attrs []string
	var class string
	if hdr.IsTitleblock {
		class = "title"
	}
	if hdr.IsSpecial {
		if class != "" {
			class += " special"
		} else {
			class = "special"
		}
	}
	if class != "" {
		attrs = []string{`class="` + class + `"`}
	}

	if hdr.HeadingID != "" {
		id := r.MakeUniqueHeadingID(hdr)
		// Escape so untrusted {#id} values cannot break out of the attribute
		// (GHSA-gc99-qr5c-98ff).
		var idBuf bytes.Buffer
		idBuf.WriteString(`id="`)
		EscapeHTML(&idBuf, []byte(id))
		idBuf.WriteByte('"')
		attrs = append(attrs, idBuf.String())
	}
	attrs = append(attrs, BlockAttrs(hdr)...)
	attrs = coalesceClassAttrs(attrs)
	r.CR(w)
	r.OutTag(w, HeadingOpenTagFromLevel(hdr.Level), attrs)
}

func (r *Renderer) HeadingExit(w io.Writer, hdr *ast.Heading) {
	r.Outs(w, HeadingCloseTagFromLevel(hdr.Level))
	if !(IsListItem(hdr.Parent) && ast.GetNextNode(hdr) == nil) {
		r.CR(w)
	}
}

// Heading writes ast.Heading node
func (r *Renderer) Heading(w io.Writer, hdr *ast.Heading, entering bool) {
	if entering {
		r.HeadingEnter(w, hdr)
	} else {
		r.HeadingExit(w, hdr)
	}
}

// HorizontalRule writes ast.HorizontalRule node
func (r *Renderer) HorizontalRule(w io.Writer, node *ast.HorizontalRule) {
	r.CR(w)
	r.OutHRTag(w, BlockAttrs(node))
	r.CR(w)
}

func (r *Renderer) listEnter(w io.Writer, nodeData *ast.List) {
	// TODO: attrs don't seem to be set
	var attrs []string

	if nodeData.IsFootnotesList {
		r.Outs(w, "\n<div class=\"footnotes\">\n\n")
		if r.Opts.Flags&FootnoteNoHRTag == 0 {
			r.OutHRTag(w, nil)
			r.CR(w)
		}
	}
	r.CR(w)
	if IsListItem(nodeData.Parent) {
		grand := nodeData.Parent.GetParent()
		if IsListTight(grand) {
			r.CR(w)
		}
	}

	openTag := "<ul"
	if nodeData.ListFlags&ast.ListTypeOrdered != 0 {
		if nodeData.Start > 0 {
			attrs = append(attrs, fmt.Sprintf(`start="%d"`, nodeData.Start))
		}
		openTag = "<ol"
	}
	if nodeData.ListFlags&ast.ListTypeDefinition != 0 {
		openTag = "<dl"
	}
	attrs = append(attrs, BlockAttrs(nodeData)...)
	r.OutTag(w, openTag, attrs)
	r.CR(w)
}

func (r *Renderer) listExit(w io.Writer, list *ast.List) {
	closeTag := "</ul>"
	if list.ListFlags&ast.ListTypeOrdered != 0 {
		closeTag = "</ol>"
	}
	if list.ListFlags&ast.ListTypeDefinition != 0 {
		closeTag = "</dl>"
	}
	r.Outs(w, closeTag)

	//cr(w)
	//if node.parent.Type != Item {
	//	cr(w)
	//}
	parent := list.Parent
	switch parent.(type) {
	case *ast.ListItem:
		if ast.GetNextNode(list) != nil {
			r.CR(w)
		}
	case *ast.Document, *ast.BlockQuote, *ast.Aside:
		r.CR(w)
	}

	if list.IsFootnotesList {
		r.Outs(w, "\n</div>\n")
	}
}

// List writes ast.List node
func (r *Renderer) List(w io.Writer, list *ast.List, entering bool) {
	if entering {
		r.listEnter(w, list)
	} else {
		r.listExit(w, list)
	}
}

func (r *Renderer) listItemEnter(w io.Writer, listItem *ast.ListItem) {
	if ListItemOpenCR(listItem) {
		r.CR(w)
	}
	if listItem.RefLink != nil {
		slug := Slugify(listItem.RefLink)
		r.Outs(w, FootnoteItem(r.Opts.FootnoteAnchorPrefix, slug))
		return
	}

	openTag := "<li>"
	if listItem.ListFlags&ast.ListTypeDefinition != 0 {
		openTag = "<dd>"
	}
	if listItem.ListFlags&ast.ListTypeTerm != 0 {
		openTag = "<dt>"
	}
	r.Outs(w, openTag)
}

func (r *Renderer) listItemExit(w io.Writer, listItem *ast.ListItem) {
	if listItem.RefLink != nil && r.Opts.Flags&FootnoteReturnLinks != 0 {
		slug := Slugify(listItem.RefLink)
		prefix := r.Opts.FootnoteAnchorPrefix
		link := r.Opts.FootnoteReturnLinkContents
		s := FootnoteReturnLink(prefix, link, slug)
		r.Outs(w, s)
	}

	closeTag := "</li>"
	if listItem.ListFlags&ast.ListTypeDefinition != 0 {
		closeTag = "</dd>"
	}
	if listItem.ListFlags&ast.ListTypeTerm != 0 {
		closeTag = "</dt>"
	}
	r.Outs(w, closeTag)
	r.CR(w)
}

// ListItem writes ast.ListItem node
func (r *Renderer) ListItem(w io.Writer, listItem *ast.ListItem, entering bool) {
	if entering {
		r.listItemEnter(w, listItem)
	} else {
		r.listItemExit(w, listItem)
	}
}

// EscapeHTMLCallouts writes html-escaped d to w. It escapes &, <, > and " characters, *but*
// expands callouts <<N>> with the callout HTML, i.e. by calling r.callout() with a newly created
// ast.Callout node.
func (r *Renderer) EscapeHTMLCallouts(w io.Writer, d []byte) {
	ld := len(d)
Parse:
	for i := 0; i < ld; i++ {
		for _, comment := range r.Opts.Comments {
			// try every configured marker, not just the first
			if !bytes.HasPrefix(d[i:], comment) {
				continue
			}

			lc := len(comment)
			if i+lc < ld {
				if id, consumed := parser.IsCallout(d[i+lc:]); consumed > 0 {
					// We have seen a callout
					callout := &ast.Callout{ID: id}
					r.Callout(w, callout)
					i += consumed + lc - 1
					continue Parse
				}
			}
		}

		escSeq := Escaper[d[i]]
		if escSeq != nil {
			w.Write(escSeq)
		} else {
			w.Write([]byte{d[i]})
		}
	}
}

// CodeBlock writes ast.CodeBlock node
func (r *Renderer) CodeBlock(w io.Writer, codeBlock *ast.CodeBlock) {
	var attrs []string
	attrs = appendLanguageAttr(attrs, codeBlock.Info)
	attrs = append(attrs, BlockAttrs(codeBlock)...)
	attrs = coalesceClassAttrs(attrs)
	r.CR(w)

	r.Outs(w, "<pre>")
	code := TagWithAttributes("<code", attrs)
	r.Outs(w, code)
	if r.Opts.Comments != nil {
		r.EscapeHTMLCallouts(w, codeBlock.Literal)
	} else {
		EscapeHTML(w, codeBlock.Literal)
	}
	r.Outs(w, "</code>")
	r.Outs(w, "</pre>")
	if !IsListItem(codeBlock.Parent) {
		r.CR(w)
	}
}

// Caption writes ast.Caption node
func (r *Renderer) Caption(w io.Writer, caption *ast.Caption, entering bool) {
	if entering {
		r.Outs(w, "<figcaption>")
		return
	}
	r.Outs(w, "</figcaption>")
}

// CaptionFigure writes ast.CaptionFigure node
func (r *Renderer) CaptionFigure(w io.Writer, figure *ast.CaptionFigure, entering bool) {
	// TODO(miek): copy more generic ways of mmark over to here.
	fig := "<figure"
	if figure.HeadingID != "" {
		fig += ` id="` + escapeAttr(figure.HeadingID) + `">`
	} else {
		fig += ">"
	}
	r.OutOneOf(w, entering, fig, "\n</figure>\n")
}

// TableCell writes ast.TableCell node
func (r *Renderer) TableCell(w io.Writer, tableCell *ast.TableCell, entering bool) {
	if !entering {
		r.OutOneOf(w, tableCell.IsHeader, "</th>", "</td>")
		r.CR(w)
		return
	}

	// entering
	var attrs []string
	openTag := "<td"
	if tableCell.IsHeader {
		openTag = "<th"
	}
	align := tableCell.Align.String()
	if align != "" {
		attrs = append(attrs, fmt.Sprintf(`align="%s"`, align))
	}
	if colspan := tableCell.ColSpan; colspan > 0 {
		attrs = append(attrs, fmt.Sprintf(`colspan="%d"`, colspan))
	}
	if ast.GetPrevNode(tableCell) == nil {
		r.CR(w)
	}
	r.OutTag(w, openTag, attrs)
}

// TableBody writes ast.TableBody node
func (r *Renderer) TableBody(w io.Writer, node *ast.TableBody, entering bool) {
	if entering {
		r.CR(w)
		r.Outs(w, "<tbody>")
		// XXX: this is to adhere to a rather silly test. Should fix test.
		if ast.GetFirstChild(node) == nil {
			r.CR(w)
		}
	} else {
		r.Outs(w, "</tbody>")
		r.CR(w)
	}
}

// DocumentMatter writes ast.DocumentMatter
func (r *Renderer) DocumentMatter(w io.Writer, node *ast.DocumentMatter, entering bool) {
	if !entering {
		return
	}
	if r.documentMatter != ast.DocumentMatterNone {
		r.Outs(w, "</section>\n")
	}
	switch node.Matter {
	case ast.DocumentMatterFront:
		r.Outs(w, `<section data-matter="front">`)
	case ast.DocumentMatterMain:
		r.Outs(w, `<section data-matter="main">`)
	case ast.DocumentMatterBack:
		r.Outs(w, `<section data-matter="back">`)
	}
	r.documentMatter = node.Matter
}

// Citation writes ast.Citation node
func (r *Renderer) Citation(w io.Writer, node *ast.Citation) {
	for i, c := range node.Destination {
		attr := []string{`class="none"`}
		switch node.Type[i] {
		case ast.CitationTypeNormative:
			attr[0] = `class="normative"`
		case ast.CitationTypeInformative:
			attr[0] = `class="informative"`
		case ast.CitationTypeSuppressed:
			attr[0] = `class="suppressed"`
		}
		r.OutTag(w, "<cite", attr)
		// Escape the destination for both the href attribute and the visible
		// text so untrusted [@dest] cannot inject markup (GHSA-g6w5-3rhc-c379).
		dest := escapeAttr(string(c))
		r.Outs(w, fmt.Sprintf(`<a href="#%s">`+r.Opts.CitationFormatString+`</a>`, dest, dest))
		r.Outs(w, "</cite>")
	}
}

// Callout writes ast.Callout node
func (r *Renderer) Callout(w io.Writer, node *ast.Callout) {
	attr := []string{`class="callout"`}
	r.OutTag(w, "<span", attr)
	r.Out(w, node.ID)
	r.Outs(w, "</span>")
}

// Index writes ast.Index node
func (r *Renderer) Index(w io.Writer, node *ast.Index) {
	// there is no in-text representation.
	attr := []string{`class="index"`, fmt.Sprintf(`id="%s"`, node.ID)}
	r.OutTag(w, "<span", attr)
	r.Outs(w, "</span>")
}

// RenderNode renders a markdown node to HTML
func (r *Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if r.Opts.RenderNodeHook != nil {
		status, didHandle := r.Opts.RenderNodeHook(w, node, entering)
		if didHandle {
			return status
		}
	}
	switch node := node.(type) {
	case *ast.Text:
		r.Text(w, node)
	case *ast.Softbreak:
		r.CR(w)
		// TODO: make it configurable via out(renderer.softbreak)
	case *ast.Hardbreak:
		r.HardBreak(w, node)
	case *ast.NonBlockingSpace:
		r.NonBlockingSpace(w, node)
	case *ast.Emph:
		r.OutOneOf(w, entering, "<em>", "</em>")
	case *ast.Strong:
		r.OutOneOf(w, entering, "<strong>", "</strong>")
	case *ast.Del:
		r.OutOneOf(w, entering, "<del>", "</del>")
	case *ast.BlockQuote:
		tag := TagWithAttributes("<blockquote", BlockAttrs(node))
		r.OutOneOfCr(w, entering, tag, "</blockquote>")
	case *ast.Aside:
		tag := TagWithAttributes("<aside", BlockAttrs(node))
		r.OutOneOfCr(w, entering, tag, "</aside>")
	case *ast.Link:
		r.Link(w, node, entering)
	case *ast.CrossReference:
		link := &ast.Link{Destination: append([]byte("#"), node.Destination...)}
		r.Link(w, link, entering)
	case *ast.Citation:
		r.Citation(w, node)
	case *ast.Image:
		if r.Opts.Flags&SkipImages != 0 {
			return ast.SkipChildren
		}
		r.Image(w, node, entering)
	case *ast.Code:
		r.Code(w, node)
	case *ast.CodeBlock:
		r.CodeBlock(w, node)
	case *ast.Caption:
		r.Caption(w, node, entering)
	case *ast.CaptionFigure:
		r.CaptionFigure(w, node, entering)
	case *ast.Document:
		// do nothing
	case *ast.Paragraph:
		r.Paragraph(w, node, entering)
	case *ast.HTMLSpan:
		r.HTMLSpan(w, node)
	case *ast.HTMLBlock:
		r.HTMLBlock(w, node)
	case *ast.Heading:
		r.Heading(w, node, entering)
	case *ast.HorizontalRule:
		r.HorizontalRule(w, node)
	case *ast.List:
		r.List(w, node, entering)
	case *ast.ListItem:
		r.ListItem(w, node, entering)
	case *ast.Table:
		tag := TagWithAttributes("<table", BlockAttrs(node))
		r.OutOneOfCr(w, entering, tag, "</table>")
	case *ast.TableCell:
		r.TableCell(w, node, entering)
	case *ast.TableHeader:
		r.OutOneOfCr(w, entering, "<thead>", "</thead>")
	case *ast.TableBody:
		r.TableBody(w, node, entering)
	case *ast.TableRow:
		r.OutOneOfCr(w, entering, "<tr>", "</tr>")
	case *ast.TableFooter:
		r.OutOneOfCr(w, entering, "<tfoot>", "</tfoot>")
	case *ast.Math:
		r.OutOneOf(w, true, `<span class="math inline">\(`, `\)</span>`)
		EscapeHTML(w, node.Literal)
		r.OutOneOf(w, false, `<span class="math inline">\(`, `\)</span>`)
	case *ast.MathBlock:
		r.OutOneOf(w, entering, `<p><span class="math display">\[`, `\]</span></p>`)
		if entering {
			EscapeHTML(w, node.Literal)
		}
	case *ast.DocumentMatter:
		r.DocumentMatter(w, node, entering)
	case *ast.Callout:
		r.Callout(w, node)
	case *ast.Index:
		r.Index(w, node)
	case *ast.Subscript:
		r.OutOneOf(w, true, "<sub>", "</sub>")
		if entering {
			Escape(w, node.Literal)
		}
		r.OutOneOf(w, false, "<sub>", "</sub>")
	case *ast.Superscript:
		r.OutOneOf(w, true, "<sup>", "</sup>")
		if entering {
			Escape(w, node.Literal)
		}
		r.OutOneOf(w, false, "<sup>", "</sup>")
	case *ast.Footnotes:
		// nothing by default; just output the list.
	case *ast.ReferenceDefinition:
		// Definitions are resolved onto links at parse time; omit from HTML.
	default:
		panic(fmt.Sprintf("Unknown node %T", node))
	}
	return ast.GoToNext
}

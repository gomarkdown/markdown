// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.

// HTMLRenderer converts AST of parsed markdown document into HTML text

package markdown

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// HTMLFlags control optional behavior of HTML renderer.
type HTMLFlags int

// HTML renderer configuration options.
const (
	HTMLFlagsNone           HTMLFlags = 0
	SkipHTML                HTMLFlags = 1 << iota // Skip preformatted HTML blocks
	SkipImages                                    // Skip embedded images
	SkipLinks                                     // Skip all links
	Safelink                                      // Only link to trusted protocols
	NofollowLinks                                 // Only link with rel="nofollow"
	NoreferrerLinks                               // Only link with rel="noreferrer"
	HrefTargetBlank                               // Add a blank target
	CompletePage                                  // Generate a complete HTML page
	UseXHTML                                      // Generate XHTML output instead of HTML
	FootnoteReturnLinks                           // Generate a link at the end of a footnote to return to the source
	Smartypants                                   // Enable smart punctuation substitutions
	SmartypantsFractions                          // Enable smart fractions (with Smartypants)
	SmartypantsDashes                             // Enable smart dashes (with Smartypants)
	SmartypantsLatexDashes                        // Enable LaTeX-style dashes (with Smartypants)
	SmartypantsAngledQuotes                       // Enable angled double quotes (with Smartypants) for double quotes rendering
	SmartypantsQuotesNBSP                         // Enable « French guillemets » (with Smartypants)
	TOC                                           // Generate a table of contents

	CommonHTMLFlags HTMLFlags = Smartypants | SmartypantsFractions | SmartypantsDashes | SmartypantsLatexDashes
)

var (
	htmlTagRe = regexp.MustCompile("(?i)^" + htmlTag)
)

const (
	htmlTag = "(?:" + openTag + "|" + closeTag + "|" + htmlComment + "|" +
		processingInstruction + "|" + declaration + "|" + cdata + ")"
	closeTag              = "</" + tagName + "\\s*[>]"
	openTag               = "<" + tagName + attribute + "*" + "\\s*/?>"
	attribute             = "(?:" + "\\s+" + attributeName + attributeValueSpec + "?)"
	attributeValue        = "(?:" + unquotedValue + "|" + singleQuotedValue + "|" + doubleQuotedValue + ")"
	attributeValueSpec    = "(?:" + "\\s*=" + "\\s*" + attributeValue + ")"
	attributeName         = "[a-zA-Z_:][a-zA-Z0-9:._-]*"
	cdata                 = "<!\\[CDATA\\[[\\s\\S]*?\\]\\]>"
	declaration           = "<![A-Z]+" + "\\s+[^>]*>"
	doubleQuotedValue     = "\"[^\"]*\""
	htmlComment           = "<!---->|<!--(?:-?[^>-])(?:-?[^-])*-->"
	processingInstruction = "[<][?].*?[?][>]"
	singleQuotedValue     = "'[^']*'"
	tagName               = "[A-Za-z][A-Za-z0-9-]*"
	unquotedValue         = "[^\"'=<>`\\x00-\\x20]+"
)

// HTMLRendererParameters is a collection of supplementary parameters tweaking
// the behavior of various parts of HTML renderer.
type HTMLRendererParameters struct {
	// Prepend this text to each relative URL.
	AbsolutePrefix string
	// Add this text to each footnote anchor, to ensure uniqueness.
	FootnoteAnchorPrefix string
	// Show this text inside the <a> tag for a footnote return link, if the
	// HTML_FOOTNOTE_RETURN_LINKS flag is enabled. If blank, the string
	// <sup>[return]</sup> is used.
	FootnoteReturnLinkContents string
	// If set, add this text to the front of each Heading ID, to ensure
	// uniqueness.
	HeadingIDPrefix string
	// If set, add this text to the back of each Heading ID, to ensure uniqueness.
	HeadingIDSuffix string

	Title string // Document title (used if CompletePage is set)
	CSS   string // Optional CSS file URL (used if CompletePage is set)
	Icon  string // Optional icon file URL (used if CompletePage is set)

	Flags HTMLFlags // Flags allow customizing this renderer's behavior
}

// HTMLRenderer implements Renderer interface for HTML output.
//
// Do not create this directly, instead use the NewHTMLRenderer function.
type HTMLRenderer struct {
	params HTMLRendererParameters

	closeTag string // how to end singleton tags: either " />" or ">"

	// Track heading IDs to prevent ID collision in a single generation.
	headingIDs map[string]int

	lastOutputLen int
	disableTags   int

	sr *SPRenderer
}

// NewHTMLRenderer creates and configures an HTMLRenderer object, which
// satisfies the Renderer interface.
func NewHTMLRenderer(params HTMLRendererParameters) *HTMLRenderer {
	// configure the rendering engine
	closeTag := ">"
	if params.Flags&UseXHTML != 0 {
		closeTag = " />"
	}

	if params.FootnoteReturnLinkContents == "" {
		params.FootnoteReturnLinkContents = `<sup>[return]</sup>`
	}

	return &HTMLRenderer{
		params: params,

		closeTag:   closeTag,
		headingIDs: make(map[string]int),

		sr: NewSmartypantsRenderer(params.Flags),
	}
}

func isHTMLTag(tag []byte, tagname string) bool {
	found, _ := findHTMLTagPos(tag, tagname)
	return found
}

// Look for a character, but ignore it when it's in any kind of quotes, it
// might be JavaScript
func skipUntilCharIgnoreQuotes(html []byte, start int, char byte) int {
	inSingleQuote := false
	inDoubleQuote := false
	inGraveQuote := false
	i := start
	for i < len(html) {
		switch {
		case html[i] == char && !inSingleQuote && !inDoubleQuote && !inGraveQuote:
			return i
		case html[i] == '\'':
			inSingleQuote = !inSingleQuote
		case html[i] == '"':
			inDoubleQuote = !inDoubleQuote
		case html[i] == '`':
			inGraveQuote = !inGraveQuote
		}
		i++
	}
	return start
}

func findHTMLTagPos(tag []byte, tagname string) (bool, int) {
	i := 0
	if i < len(tag) && tag[0] != '<' {
		return false, -1
	}
	i++
	i = skipSpace(tag, i)

	if i < len(tag) && tag[i] == '/' {
		i++
	}

	i = skipSpace(tag, i)
	j := 0
	for ; i < len(tag); i, j = i+1, j+1 {
		if j >= len(tagname) {
			break
		}

		if strings.ToLower(string(tag[i]))[0] != tagname[j] {
			return false, -1
		}
	}

	if i == len(tag) {
		return false, -1
	}

	rightAngle := skipUntilCharIgnoreQuotes(tag, i, '>')
	if rightAngle >= i {
		return true, rightAngle
	}

	return false, -1
}

func isRelativeLink(link []byte) (yes bool) {
	// a tag begin with '#'
	if link[0] == '#' {
		return true
	}

	// link begin with '/' but not '//', the second maybe a protocol relative link
	if len(link) >= 2 && link[0] == '/' && link[1] != '/' {
		return true
	}

	// only the root '/'
	if len(link) == 1 && link[0] == '/' {
		return true
	}

	// current directory : begin with "./"
	if bytes.HasPrefix(link, []byte("./")) {
		return true
	}

	// parent directory : begin with "../"
	if bytes.HasPrefix(link, []byte("../")) {
		return true
	}

	return false
}

func (r *HTMLRenderer) ensureUniqueHeadingID(id string) string {
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

func (r *HTMLRenderer) addAbsPrefix(link []byte) []byte {
	if r.params.AbsolutePrefix != "" && isRelativeLink(link) && link[0] != '.' {
		newDest := r.params.AbsolutePrefix
		if link[0] != '/' {
			newDest += "/"
		}
		newDest += string(link)
		return []byte(newDest)
	}
	return link
}

func appendLinkAttrs(attrs []string, flags HTMLFlags, link []byte) []string {
	if isRelativeLink(link) {
		return attrs
	}
	var val []string
	if flags&NofollowLinks != 0 {
		val = append(val, "nofollow")
	}
	if flags&NoreferrerLinks != 0 {
		val = append(val, "noreferrer")
	}
	if flags&HrefTargetBlank != 0 {
		attrs = append(attrs, `target="_blank"`)
	}
	if len(val) == 0 {
		return attrs
	}
	attr := fmt.Sprintf("rel=%q", strings.Join(val, " "))
	return append(attrs, attr)
}

func isMailto(link []byte) bool {
	return bytes.HasPrefix(link, []byte("mailto:"))
}

func needSkipLink(flags HTMLFlags, dest []byte) bool {
	if flags&SkipLinks != 0 {
		return true
	}
	return flags&Safelink != 0 && !isSafeLink(dest) && !isMailto(dest)
}

func isSmartypantable(node *Node) bool {
	switch node.Parent.Data.(type) {
	case *LinkData, *CodeBlockData, *CodeData:
		return false
	}
	return true
}

func appendLanguageAttr(attrs []string, info []byte) []string {
	if len(info) == 0 {
		return attrs
	}
	endOfLang := bytes.IndexAny(info, "\t ")
	if endOfLang < 0 {
		endOfLang = len(info)
	}
	return append(attrs, fmt.Sprintf("class=\"language-%s\"", info[:endOfLang]))
}

func (r *HTMLRenderer) outTag(w io.Writer, name string, attrs []string) {
	io.WriteString(w, name)
	if len(attrs) > 0 {
		io.WriteString(w, " ")
		io.WriteString(w, strings.Join(attrs, " "))
	}
	io.WriteString(w, ">")
	r.lastOutputLen = 1
}

func footnoteRef(prefix string, node *LinkData) string {
	urlFrag := prefix + string(slugify(node.Destination))
	anchor := fmt.Sprintf(`<a rel="footnote" href="#fn:%s">%d</a>`, urlFrag, node.NoteID)
	return fmt.Sprintf(`<sup class="footnote-ref" id="fnref:%s">%s</sup>`, urlFrag, anchor)
}

func footnoteItem(prefix string, slug []byte) string {
	return fmt.Sprintf(`<li id="fn:%s%s">`, prefix, slug)
}

func footnoteReturnLink(prefix, returnLink string, slug []byte) string {
	const format = ` <a class="footnote-return" href="#fnref:%s%s">%s</a>`
	return fmt.Sprintf(format, prefix, slug, returnLink)
}

func itemOpenCR(node *Node) bool {
	if node.Prev == nil {
		return false
	}
	ld := node.Parent.Data.(*ListData)
	return !ld.Tight && ld.ListFlags&ListTypeDefinition == 0
}

func skipParagraphTags(node *Node) bool {
	parent := node.Parent
	grandparent := parent.Parent
	if grandparent == nil || !isListData(grandparent.Data) {
		return false
	}
	isParentTerm := isListItemTerm(parent)
	grandparentListData := grandparent.Data.(*ListData)
	tightOrTerm := grandparentListData.Tight || isParentTerm
	return tightOrTerm
}

func cellAlignment(align CellAlignFlags) string {
	switch align {
	case TableAlignmentLeft:
		return "left"
	case TableAlignmentRight:
		return "right"
	case TableAlignmentCenter:
		return "center"
	default:
		return ""
	}
}

func (r *HTMLRenderer) out(w io.Writer, d []byte) {
	r.lastOutputLen = len(d)
	if r.disableTags > 0 {
		d = htmlTagRe.ReplaceAll(d, []byte{})
	}
	w.Write(d)
}

func (r *HTMLRenderer) outs(w io.Writer, s string) {
	r.lastOutputLen = len(s)
	if r.disableTags > 0 {
		s = htmlTagRe.ReplaceAllString(s, "")
	}
	io.WriteString(w, s)
}

func (r *HTMLRenderer) cr(w io.Writer) {
	if r.lastOutputLen > 0 {
		r.outs(w, "\n")
	}
}

var (
	openHTags  = []string{"<h1", "<h2", "<h3", "<h4", "<h5"}
	closeHTags = []string{"</h1>", "</h2>", "</h3>", "</h4>", "</h5>"}
)

func headingOpenTagFromLevel(level int) string {
	if level < 1 || level > 5 {
		return "<h6"
	}
	return openHTags[level-1]
}

func headingCloseTagFromLevel(level int) string {
	if level < 1 || level > 5 {
		return "</h6>"
	}
	return closeHTags[level-1]
}

func (r *HTMLRenderer) outHRTag(w io.Writer) {
	r.outOneOf(w, r.params.Flags&UseXHTML == 0, "<hr>", "<hr />")
}

func (r *HTMLRenderer) text(w io.Writer, node *Node, nodeData *TextData) {
	if r.params.Flags&Smartypants != 0 {
		var tmp bytes.Buffer
		escapeHTML(&tmp, node.Literal)
		r.sr.Process(w, tmp.Bytes())
	} else {
		if isLinkData(node.Parent.Data) {
			escLink(w, node.Literal)
		} else {
			escapeHTML(w, node.Literal)
		}
	}
}

func (r *HTMLRenderer) hardBreak(w io.Writer, node *Node, nodeData *HardbreakData) {
	r.outOneOf(w, r.params.Flags&UseXHTML == 0, "<br>", "<br />")
	r.cr(w)
}

func (r *HTMLRenderer) outOneOf(w io.Writer, outFirst bool, first string, second string) {
	if outFirst {
		r.outs(w, first)
	} else {
		r.outs(w, second)
	}
}

func (r *HTMLRenderer) outOneOfCr(w io.Writer, outFirst bool, first string, second string) {
	if outFirst {
		r.cr(w)
		r.outs(w, first)
	} else {
		r.outs(w, second)
		r.cr(w)
	}
}

func (r *HTMLRenderer) span(w io.Writer, node *Node, nodeData *HTMLSpanData) {
	if r.params.Flags&SkipHTML == 0 {
		r.out(w, node.Literal)
	}
}

func (r *HTMLRenderer) linkEnter(w io.Writer, node *Node, nodeData *LinkData) {
	var attrs []string
	dest := nodeData.Destination
	dest = r.addAbsPrefix(dest)
	var hrefBuf bytes.Buffer
	hrefBuf.WriteString("href=\"")
	escLink(&hrefBuf, dest)
	hrefBuf.WriteByte('"')
	attrs = append(attrs, hrefBuf.String())
	if nodeData.NoteID != 0 {
		r.outs(w, footnoteRef(r.params.FootnoteAnchorPrefix, nodeData))
		return
	}

	attrs = appendLinkAttrs(attrs, r.params.Flags, dest)
	if len(nodeData.Title) > 0 {
		var titleBuff bytes.Buffer
		titleBuff.WriteString("title=\"")
		escapeHTML(&titleBuff, nodeData.Title)
		titleBuff.WriteByte('"')
		attrs = append(attrs, titleBuff.String())
	}
	r.outTag(w, "<a", attrs)
}

func (r *HTMLRenderer) linkExit(w io.Writer, node *Node, nodeData *LinkData) {
	if nodeData.NoteID == 0 {
		r.outs(w, "</a>")
	}
}

func (r *HTMLRenderer) link(w io.Writer, node *Node, nodeData *LinkData, entering bool) {
	// mark it but don't link it if it is not a safe link: no smartypants
	if needSkipLink(r.params.Flags, nodeData.Destination) {
		r.outOneOf(w, entering, "<tt>", "</tt>")
		return
	}

	if entering {
		r.linkEnter(w, node, nodeData)
	} else {
		r.linkExit(w, node, nodeData)
	}
}

func (r *HTMLRenderer) imageEnter(w io.Writer, node *Node, nodeData *ImageData) {
	dest := nodeData.Destination
	dest = r.addAbsPrefix(dest)
	if r.disableTags == 0 {
		//if options.safe && potentiallyUnsafe(dest) {
		//out(w, `<img src="" alt="`)
		//} else {
		r.outs(w, `<img src="`)
		escLink(w, dest)
		r.outs(w, `" alt="`)
		//}
	}
	r.disableTags++
}

func (r *HTMLRenderer) imageExit(w io.Writer, node *Node, nodeData *ImageData) {
	r.disableTags--
	if r.disableTags == 0 {
		if nodeData.Title != nil {
			r.outs(w, `" title="`)
			escapeHTML(w, nodeData.Title)
		}
		r.outs(w, `" />`)
	}
}

func (r *HTMLRenderer) paragraphEnter(w io.Writer, node *Node, nodeData *ParagraphData) {
	// TODO: untangle this clusterfuck about when the newlines need
	// to be added and when not.
	if node.Prev != nil {
		switch node.Prev.Data.(type) {
		case *HTMLBlockData, *ListData, *ParagraphData, *HeadingData, *CodeBlockData, *BlockQuoteData, *HorizontalRuleData:
			r.cr(w)
		}
	}
	if isBlockQuoteData(node.Parent.Data) && node.Prev == nil {
		r.cr(w)
	}
	r.outs(w, "<p>")
}

func (r *HTMLRenderer) paragraphExit(w io.Writer, node *Node, nodeData *ParagraphData) {
	r.outs(w, "</p>")
	if !(isListItemData(node.Parent.Data) && node.Next == nil) {
		r.cr(w)
	}
}

func (r *HTMLRenderer) paragraph(w io.Writer, node *Node, nodeData *ParagraphData, entering bool) {
	if skipParagraphTags(node) {
		return
	}
	if entering {
		r.paragraphEnter(w, node, nodeData)
	} else {
		r.paragraphExit(w, node, nodeData)
	}
}
func (r *HTMLRenderer) image(w io.Writer, node *Node, nodeData *ImageData, entering bool) {
	if entering {
		r.imageEnter(w, node, nodeData)
	} else {
		r.imageExit(w, node, nodeData)
	}
}

func (r *HTMLRenderer) code(w io.Writer, node *Node, nodeData *CodeData) {
	r.outs(w, "<code>")
	escapeHTML(w, node.Literal)
	r.outs(w, "</code>")
}

func (r *HTMLRenderer) htmlBlock(w io.Writer, node *Node, nodeData *HTMLBlockData) {
	if r.params.Flags&SkipHTML != 0 {
		return
	}
	r.cr(w)
	r.out(w, node.Literal)
	r.cr(w)
}

func (r *HTMLRenderer) headingEnter(w io.Writer, node *Node, nodeData *HeadingData) {
	var attrs []string
	if nodeData.IsTitleblock {
		attrs = append(attrs, `class="title"`)
	}
	if nodeData.HeadingID != "" {
		id := r.ensureUniqueHeadingID(nodeData.HeadingID)
		if r.params.HeadingIDPrefix != "" {
			id = r.params.HeadingIDPrefix + id
		}
		if r.params.HeadingIDSuffix != "" {
			id = id + r.params.HeadingIDSuffix
		}
		attrID := `id="` + id + `"`
		attrs = append(attrs, attrID)
	}
	r.cr(w)
	r.outTag(w, headingOpenTagFromLevel(nodeData.Level), attrs)
}

func (r *HTMLRenderer) headingExit(w io.Writer, node *Node, nodeData *HeadingData) {
	r.outs(w, headingCloseTagFromLevel(nodeData.Level))
	if !(isListItemData(node.Parent.Data) && node.Next == nil) {
		r.cr(w)
	}
}

func (r *HTMLRenderer) heading(w io.Writer, node *Node, nodeData *HeadingData, entering bool) {
	if entering {
		r.headingEnter(w, node, nodeData)
	} else {
		r.headingExit(w, node, nodeData)
	}
}

func (r *HTMLRenderer) horizontalRule(w io.Writer) {
	r.cr(w)
	r.outHRTag(w)
	r.cr(w)
}

func (r *HTMLRenderer) listEnter(w io.Writer, node *Node, nodeData *ListData) {
	// TODO: attrs don't seem to be set
	var attrs []string

	if nodeData.IsFootnotesList {
		r.outs(w, "\n<div class=\"footnotes\">\n\n")
		r.outHRTag(w)
		r.cr(w)
	}
	r.cr(w)
	if isListItemData(node.Parent.Data) {
		grand := node.Parent.Parent
		if isListTight(grand.Data) {
			r.cr(w)
		}
	}

	openTag := "<ul"
	if nodeData.ListFlags&ListTypeOrdered != 0 {
		openTag = "<ol"
	}
	if nodeData.ListFlags&ListTypeDefinition != 0 {
		openTag = "<dl"
	}
	r.outTag(w, openTag, attrs)
	r.cr(w)
}

func (r *HTMLRenderer) listExit(w io.Writer, node *Node, nodeData *ListData) {
	closeTag := "</ul>"
	if nodeData.ListFlags&ListTypeOrdered != 0 {
		closeTag = "</ol>"
	}
	if nodeData.ListFlags&ListTypeDefinition != 0 {
		closeTag = "</dl>"
	}
	r.outs(w, closeTag)

	//cr(w)
	//if node.parent.Type != Item {
	//	cr(w)
	//}
	if isListItemData(node.Parent.Data) && node.Next != nil {
		r.cr(w)
	}
	if isDocumentData(node.Parent.Data) || isBlockQuoteData(node.Parent.Data) {
		r.cr(w)
	}
	if nodeData.IsFootnotesList {
		r.outs(w, "\n</div>\n")
	}
}

func (r *HTMLRenderer) list(w io.Writer, node *Node, nodeData *ListData, entering bool) {
	if entering {
		r.listEnter(w, node, nodeData)
	} else {
		r.listExit(w, node, nodeData)
	}
}

func (r *HTMLRenderer) listItemEnter(w io.Writer, node *Node, nodeData *ListItemData) {
	if itemOpenCR(node) {
		r.cr(w)
	}
	if nodeData.RefLink != nil {
		slug := slugify(nodeData.RefLink)
		r.outs(w, footnoteItem(r.params.FootnoteAnchorPrefix, slug))
		return
	}

	openTag := "<li>"
	if nodeData.ListFlags&ListTypeDefinition != 0 {
		openTag = "<dd>"
	}
	if nodeData.ListFlags&ListTypeTerm != 0 {
		openTag = "<dt>"
	}
	r.outs(w, openTag)
}

func (r *HTMLRenderer) listItemExit(w io.Writer, node *Node, nodeData *ListItemData) {
	if nodeData.RefLink != nil && r.params.Flags&FootnoteReturnLinks != 0 {
		slug := slugify(nodeData.RefLink)
		prefix := r.params.FootnoteAnchorPrefix
		link := r.params.FootnoteReturnLinkContents
		s := footnoteReturnLink(prefix, link, slug)
		r.outs(w, s)
	}

	closeTag := "</li>"
	if nodeData.ListFlags&ListTypeDefinition != 0 {
		closeTag = "</dd>"
	}
	if nodeData.ListFlags&ListTypeTerm != 0 {
		closeTag = "</dt>"
	}
	r.outs(w, closeTag)
	r.cr(w)
}

func (r *HTMLRenderer) listItem(w io.Writer, node *Node, nodeData *ListItemData, entering bool) {
	if entering {
		r.listItemEnter(w, node, nodeData)
	} else {
		r.listItemExit(w, node, nodeData)
	}
}

func (r *HTMLRenderer) codeBlock(w io.Writer, node *Node, nodeData *CodeBlockData) {
	var attrs []string
	attrs = appendLanguageAttr(attrs, nodeData.Info)
	r.cr(w)
	r.outs(w, "<pre>")
	r.outTag(w, "<code", attrs)
	escapeHTML(w, node.Literal)
	r.outs(w, "</code>")
	r.outs(w, "</pre>")
	if !isListItemData(node.Parent.Data) {
		r.cr(w)
	}
}

func (r *HTMLRenderer) tableCell(w io.Writer, node *Node, nodeData *TableCellData, entering bool) {
	if !entering {
		r.outOneOf(w, nodeData.IsHeader, "</th>", "</td>")
		r.cr(w)
		return
	}

	// entering
	var attrs []string
	openTag := "<td"
	if nodeData.IsHeader {
		openTag = "<th"
	}
	align := cellAlignment(nodeData.Align)
	if align != "" {
		attrs = append(attrs, fmt.Sprintf(`align="%s"`, align))
	}
	if node.Prev == nil {
		r.cr(w)
	}
	r.outTag(w, openTag, attrs)
}

func (r *HTMLRenderer) tableBody(w io.Writer, node *Node, nodeData *TableBodyData, entering bool) {
	if entering {
		r.cr(w)
		r.outs(w, "<tbody>")
		// XXX: this is to adhere to a rather silly test. Should fix test.
		if node.FirstChild == nil {
			r.cr(w)
		}
	} else {
		r.outs(w, "</tbody>")
		r.cr(w)
	}
}

// RenderNode is a default renderer of a single node of a syntax tree. For
// block nodes it will be called twice: first time with entering=true, second
// time with entering=false, so that it could know when it's working on an open
// tag and when on close. It writes the result to w.
//
// The return value is a way to tell the calling walker to adjust its walk
// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
// can ask the walker to skip a subtree of this node by returning SkipChildren.
// The typical behavior is to return GoToNext, which asks for the usual
// traversal to the next node.
func (r *HTMLRenderer) RenderNode(w io.Writer, node *Node, entering bool) WalkStatus {
	switch nodeData := node.Data.(type) {
	case *TextData:
		r.text(w, node, nodeData)
	case *SoftbreakData:
		r.cr(w)
		// TODO: make it configurable via out(renderer.softbreak)
	case *HardbreakData:
		r.hardBreak(w, node, nodeData)
	case *EmphData:
		r.outOneOf(w, entering, "<em>", "</em>")
	case *StrongData:
		r.outOneOf(w, entering, "<strong>", "</strong>")
	case *DelData:
		r.outOneOf(w, entering, "<del>", "</del>")
	case *HTMLSpanData:
		r.span(w, node, nodeData)
	case *LinkData:
		r.link(w, node, nodeData, entering)
	case *ImageData:
		if r.params.Flags&SkipImages != 0 {
			return SkipChildren
		}
		r.image(w, node, nodeData, entering)
	case *CodeData:
		r.code(w, node, nodeData)
	case *DocumentData:
		// do nothing
	case *ParagraphData:
		r.paragraph(w, node, nodeData, entering)
	case *BlockQuoteData:
		r.outOneOfCr(w, entering, "<blockquote>", "</blockquote>")
	case *HTMLBlockData:
		r.htmlBlock(w, node, nodeData)
	case *HeadingData:
		r.heading(w, node, nodeData, entering)
	case *HorizontalRuleData:
		r.horizontalRule(w)
	case *ListData:
		r.list(w, node, nodeData, entering)
	case *ListItemData:
		r.listItem(w, node, nodeData, entering)
	case *CodeBlockData:
		r.codeBlock(w, node, nodeData)
	case *TableData:
		r.outOneOfCr(w, entering, "<table>", "</table>")
	case *TableCellData:
		r.tableCell(w, node, nodeData, entering)
	case *TableHeadData:
		r.outOneOfCr(w, entering, "<thead>", "</thead>")
	case *TableBodyData:
		r.tableBody(w, node, nodeData, entering)
	case *TableRowData:
		r.outOneOfCr(w, entering, "<tr>", "</tr>")
	default:
		//panic("Unknown node type " + node.Type.String())
		panic(fmt.Sprintf("Unknown node type %T", node.Data))
	}
	return GoToNext
}

// RenderHeader writes HTML document preamble and TOC if requested.
func (r *HTMLRenderer) RenderHeader(w io.Writer, ast *Node) {
	r.writeDocumentHeader(w)
	if r.params.Flags&TOC != 0 {
		r.writeTOC(w, ast)
	}
}

// RenderFooter writes HTML document footer.
func (r *HTMLRenderer) RenderFooter(w io.Writer, ast *Node) {
	if r.params.Flags&CompletePage == 0 {
		return
	}
	io.WriteString(w, "\n</body>\n</html>\n")
}

func (r *HTMLRenderer) writeDocumentHeader(w io.Writer) {
	if r.params.Flags&CompletePage == 0 {
		return
	}
	ending := ""
	if r.params.Flags&UseXHTML != 0 {
		io.WriteString(w, "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" ")
		io.WriteString(w, "\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\n")
		io.WriteString(w, "<html xmlns=\"http://www.w3.org/1999/xhtml\">\n")
		ending = " /"
	} else {
		io.WriteString(w, "<!DOCTYPE html>\n")
		io.WriteString(w, "<html>\n")
	}
	io.WriteString(w, "<head>\n")
	io.WriteString(w, "  <title>")
	if r.params.Flags&Smartypants != 0 {
		r.sr.Process(w, []byte(r.params.Title))
	} else {
		escapeHTML(w, []byte(r.params.Title))
	}
	io.WriteString(w, "</title>\n")
	io.WriteString(w, "  <meta name=\"GENERATOR\" content=\"Markdown Processor for Go v")
	io.WriteString(w, Version)
	io.WriteString(w, "\"")
	io.WriteString(w, ending)
	io.WriteString(w, ">\n")
	io.WriteString(w, "  <meta charset=\"utf-8\"")
	io.WriteString(w, ending)
	io.WriteString(w, ">\n")
	if r.params.CSS != "" {
		io.WriteString(w, "  <link rel=\"stylesheet\" type=\"text/css\" href=\"")
		escapeHTML(w, []byte(r.params.CSS))
		io.WriteString(w, "\"")
		io.WriteString(w, ending)
		io.WriteString(w, ">\n")
	}
	if r.params.Icon != "" {
		io.WriteString(w, "  <link rel=\"icon\" type=\"image/x-icon\" href=\"")
		escapeHTML(w, []byte(r.params.Icon))
		io.WriteString(w, "\"")
		io.WriteString(w, ending)
		io.WriteString(w, ">\n")
	}
	io.WriteString(w, "</head>\n")
	io.WriteString(w, "<body>\n\n")
}

func (r *HTMLRenderer) writeTOC(w io.Writer, ast *Node) {
	buf := bytes.Buffer{}

	inHeading := false
	tocLevel := 0
	headingCount := 0

	ast.WalkFunc(func(node *Node, entering bool) WalkStatus {
		if nodeData, ok := node.Data.(*HeadingData); ok && !nodeData.IsTitleblock {
			inHeading = entering
			if entering {
				nodeData.HeadingID = fmt.Sprintf("toc_%d", headingCount)
				if nodeData.Level == tocLevel {
					buf.WriteString("</li>\n\n<li>")
				} else if nodeData.Level < tocLevel {
					for nodeData.Level < tocLevel {
						tocLevel--
						buf.WriteString("</li>\n</ul>")
					}
					buf.WriteString("</li>\n\n<li>")
				} else {
					for nodeData.Level > tocLevel {
						tocLevel++
						buf.WriteString("\n<ul>\n<li>")
					}
				}

				fmt.Fprintf(&buf, `<a href="#toc_%d">`, headingCount)
				headingCount++
			} else {
				buf.WriteString("</a>")
			}
			return GoToNext
		}

		if inHeading {
			return r.RenderNode(&buf, node, entering)
		}

		return GoToNext
	})

	for ; tocLevel > 0; tocLevel-- {
		buf.WriteString("</li>\n</ul>")
	}

	if buf.Len() > 0 {
		io.WriteString(w, "<nav>\n")
		w.Write(buf.Bytes())
		io.WriteString(w, "\n\n</nav>\n")
	}
	r.lastOutputLen = buf.Len()
}

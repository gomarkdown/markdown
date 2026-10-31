package md

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// Renderer renders to markdown. Allows converting to a canonical
// form.
type Renderer struct {
	orderedListCounter map[int]int

	lastOutput        bool // the current paragraph has written something
	listDepth         int
	previousWasNumber bool
	inTableCell       bool

	C *RendererConfig

	linkcache map[string]bool // cache for link definitions to write in the footer, if renderLinksInFooter is set
}

const listIndentSize = 4

type RendererConfig struct {
	Flags Flags
}

type Flags int

const renderLinksInFooter Flags = 1 << iota

type RendererOpt func(c *RendererConfig)

// NewRenderer returns a Markdown renderer.
func NewRenderer(opts ...RendererOpt) *Renderer {
	c := &RendererConfig{}
	for _, opt := range opts {
		opt(c)
	}
	return &Renderer{
		orderedListCounter: map[int]int{},
		C:                  c,
	}
}

func WithRenderInFooter(renderInFooter bool) RendererOpt {
	return func(c *RendererConfig) {
		if renderInFooter {
			c.Flags |= renderLinksInFooter
		} else {
			c.Flags &^= renderLinksInFooter
		}
	}
}

func (r *Renderer) out(w io.Writer, d []byte) {
	r.lastOutput = r.lastOutput || len(d) > 0
	w.Write(d)
}

func (r *Renderer) outs(w io.Writer, s string) {
	r.lastOutput = r.lastOutput || len(s) > 0
	io.WriteString(w, s)
}

// render returns the output for node and its children.
func (r *Renderer) render(node ast.Node) []byte {
	var buf bytes.Buffer
	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		return r.RenderNode(&buf, n, entering)
	})
	return buf.Bytes()
}

// isBlockNode returns true if node is a block and not inline content.
func isBlockNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.Paragraph, *ast.List, *ast.ListItem, *ast.CodeBlock, *ast.Heading, *ast.HTMLBlock,
		*ast.BlockQuote, *ast.Aside, *ast.HorizontalRule, *ast.Table, *ast.MathBlock,
		*ast.CaptionFigure, *ast.Caption, *ast.DocumentMatter, *ast.Footnotes, *ast.ReferenceDefinition:
		return true
	}
	return false
}

// renderBlocks returns the output for children of node without newlines around
// them, separated by a blank line or, if tight, by a line break where possible.
func (r *Renderer) renderBlocks(node ast.Node, tight bool) []byte {
	var res []byte
	var prev ast.Node
	for _, child := range node.GetChildren() {
		block := bytes.Trim(r.render(child), "\n")
		if len(block) == 0 {
			continue
		}
		if prev != nil {
			_, prevIsPara := prev.(*ast.Paragraph)
			_, isPara := child.(*ast.Paragraph)
			switch {
			case !isBlockNode(prev) && !isBlockNode(child):
			case tight && !(prevIsPara && isPara):
				res = append(res, '\n')
			default:
				res = append(res, "\n\n"...)
			}
		}
		res = append(res, block...)
		prev = child
	}
	return res
}

// prefixLines adds prefix to all lines of text but the first and to no empty lines.
func prefixLines(text []byte, prefix string) []byte {
	lines := bytes.Split(text, []byte("\n"))
	for i, line := range lines {
		if i > 0 && len(line) > 0 {
			lines[i] = append([]byte(prefix), line...)
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

func (r *Renderer) list(w io.Writer, node *ast.List, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	r.listDepth++
	if node.ListFlags&ast.ListTypeOrdered != 0 {
		if r.orderedListCounter == nil {
			r.orderedListCounter = map[int]int{}
		}
		r.orderedListCounter[r.listDepth] = 1
		if node.Start > 0 {
			r.orderedListCounter[r.listDepth] = node.Start
		}
	}
	for i, item := range node.GetChildren() {
		// without a blank line a term would become a part of the previous definition
		listItem, _ := item.(*ast.ListItem)
		isTerm := listItem != nil && listItem.ListFlags&ast.ListTypeTerm != 0
		if i > 0 && (!node.Tight || isTerm) {
			r.outs(w, "\n")
		}
		r.out(w, r.render(item))
	}
	r.listDepth--
	if _, ok := node.Parent.(*ast.ListItem); !ok {
		r.outs(w, "\n")
	}
	return ast.SkipChildren
}

func (r *Renderer) listItem(w io.Writer, node *ast.ListItem, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	var marker string
	if node.ListFlags&ast.ListTypeTerm != 0 {
		marker = ""
	} else if node.ListFlags&ast.ListTypeDefinition != 0 {
		marker = ": "
	} else if node.ListFlags&ast.ListTypeOrdered != 0 {
		delimiter := byte('.')
		if node.Delimiter == ')' {
			delimiter = ')'
		}
		marker = fmt.Sprintf("%d%c ", r.orderedListCounter[r.listDepth], delimiter)
		r.orderedListCounter[r.listDepth]++
	} else {
		marker = string(node.BulletChar) + " "
	}
	// lines must be indented at least to the content after the marker. CommonMark
	// keeps any extra indent as content, while the legacy parser needs 4 spaces.
	indent := len(marker)
	if indent < listIndentSize && !isCommonMarkDoc(node) {
		indent = listIndentSize
	}
	list, _ := node.Parent.(*ast.List)
	content := r.renderBlocks(node, list != nil && list.Tight)
	r.outs(w, marker)
	r.out(w, prefixLines(content, strings.Repeat(" ", indent)))
	r.outs(w, "\n")
	return ast.SkipChildren
}

func (r *Renderer) para(w io.Writer, node *ast.Paragraph, entering bool) {
	if entering {
		r.lastOutput = false
		return
	}
	if r.lastOutput {
		br := "\n\n"

		// List items don't need the extra line-break.
		if _, ok := node.Parent.(*ast.ListItem); ok {
			br = "\n"
		}

		r.outs(w, br)
	}
}

// escape replaces instances of backslash with escaped backslash in text.
func escape(text []byte) []byte { return bytes.ReplaceAll(text, []byte(`\`), []byte(`\\`)) }

func isNumber(data []byte) bool {
	for _, b := range data {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func needsEscaping(text []byte, previousWasNumber bool) bool {
	switch string(text) {
	case `\`,
		"`",
		"*",
		"_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#",
		"+",
		"-",
		"$",
		"^":
		return true
	case "!":
		return false
	case ".":
		// Return true if number, because a period after a number must be escaped to not get parsed as an ordered list.
		return previousWasNumber
	case "<", ">":
		return true
	default:
		return false
	}
}

func cleanWithoutTrim(text []byte) []byte {
	clean := make([]byte, 0, len(text))
	var previous byte
	for _, q := range text {
		if q == '\n' || q == '\r' || q == '\t' {
			q = ' '
		}
		if q != ' ' || previous != ' ' {
			clean = append(clean, q)
			previous = q
		}
	}
	return clean
}

// isCommonMarkDoc returns true if node is in a document parsed as CommonMark.
func isCommonMarkDoc(node ast.Node) bool {
	for ; node != nil; node = node.GetParent() {
		if doc, ok := node.(*ast.Document); ok {
			return doc.CommonMark
		}
	}
	return false
}

// isRawText returns true if the literal of text is still in Markdown syntax.
// The legacy parser doesn't parse image descriptions.
func isRawText(text *ast.Text) bool {
	_, ok := text.Parent.(*ast.Image)
	return ok && !isCommonMarkDoc(text)
}

// isAtLineStart returns true if text starts a line of a block where a leading
// marker would start a new block.
func isAtLineStart(text *ast.Text) bool {
	prev := ast.GetPrevNode(text)
	for {
		sibling, ok := prev.(*ast.Text)
		if !ok || len(sibling.Literal) > 0 {
			break
		}
		prev = ast.GetPrevNode(sibling)
	}
	switch prev.(type) {
	case nil:
		switch text.Parent.(type) {
		case *ast.Paragraph, *ast.ListItem:
			return true
		}
		return false
	case *ast.Hardbreak, *ast.Softbreak:
		return true
	}
	return false
}

// blockMarkerIndex returns the index of the character to escape so that text at
// the start of a line doesn't start a block, or -1 if there's no such character.
func blockMarkerIndex(text []byte) int {
	i := 0
	for i < len(text) && text[i] == ' ' {
		i++
	}
	if i == len(text) {
		return -1
	}
	first := i
	switch c := text[first]; c {
	case '#', '>':
		return first
	case '+', '-':
		// a list item or a setext heading underline or a horizontal rule
		if first+1 == len(text) || text[first+1] == ' ' || len(bytes.Trim(text[first:], string(c)+" ")) == 0 {
			return first
		}
		return -1
	}
	for i < len(text) && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	if i > first && i < len(text) && (text[i] == '.' || text[i] == ')') && (i+1 == len(text) || text[i+1] == ' ') {
		return i
	}
	return -1
}

// nextNonEmptyNode returns the next sibling, skipping the empty text nodes that
// the legacy parser leaves around inline nodes.
func nextNonEmptyNode(node ast.Node) ast.Node {
	next := ast.GetNextNode(node)
	for {
		text, ok := next.(*ast.Text)
		if !ok || len(text.Literal) > 0 {
			return next
		}
		next = ast.GetNextNode(text)
	}
}

// startsWithEntity returns true if text, which starts with &, would be parsed
// as an entity.
func startsWithEntity(text []byte) bool {
	i := 1
	if i < len(text) && text[i] == '#' {
		i++
	}
	start := i
	for i < len(text) && text[i] < 0x80 && isWordChar(text[i]) {
		i++
	}
	return i > start && i < len(text) && text[i] == ';'
}

func isWordChar(c byte) bool {
	return c >= 0x80 || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// escapeText escapes the characters in the text of a node that would otherwise
// be parsed as Markdown syntax.
func escapeText(clean []byte, text *ast.Text) []byte {
	marker := -1
	if isAtLineStart(text) {
		marker = blockMarkerIndex(clean)
	}
	res := make([]byte, 0, len(clean))
	last := len(clean) - 1
	for i, c := range clean {
		needsEscape := i == marker
		switch c {
		case '\\', '`', '*', '[', ']':
			needsEscape = true
		case '_':
			// not emphasis inside of a word
			needsEscape = i == 0 || i == last || !isWordChar(clean[i-1]) || !isWordChar(clean[i+1])
		case '<':
			needsEscape = i == last || clean[i+1] != ' '
		case '~':
			needsEscape = true
		case '!':
			// would turn a link that follows into an image
			_, nextIsLink := nextNonEmptyNode(text).(*ast.Link)
			needsEscape = i == last && nextIsLink
		case '&':
			rest := clean[i:]
			if next, ok := nextNonEmptyNode(text).(*ast.Text); ok && i == last {
				// the legacy parser splits text at some characters
				rest = append([]byte{'&'}, next.Literal...)
			}
			needsEscape = startsWithEntity(rest)
		}
		if needsEscape {
			res = append(res, '\\')
		}
		res = append(res, c)
	}
	return res
}

func (r *Renderer) text(w io.Writer, text *ast.Text) {
	lit := text.Literal
	raw := isRawText(text)
	needsEscape := !raw && needsEscaping(lit, r.previousWasNumber)
	r.previousWasNumber = isNumber(lit)
	if r.listDepth > 0 && len(lit) == 1 && lit[0] == '\n' {
		return
	}
	clean := cleanWithoutTrim(lit)
	if len(clean) == 0 {
		return
	}
	if needsEscape {
		clean = append([]byte("\\"), clean...)
	} else if !raw {
		clean = escapeText(clean, text)
	}
	if r.inTableCell {
		clean = bytes.ReplaceAll(clean, []byte("|"), []byte(`\|`))
	}
	r.out(w, clean)
}

func (r *Renderer) htmlSpan(w io.Writer, node *ast.HTMLSpan) {
	r.out(w, node.Literal)
}

func (r *Renderer) htmlBlock(w io.Writer, node *ast.HTMLBlock) {
	r.outs(w, "\n")
	r.out(w, node.Literal)
	r.outs(w, "\n\n")
}

func (r *Renderer) codeBlock(w io.Writer, node *ast.CodeBlock) {
	r.outs(w, "\n")
	text := node.Literal
	info := strings.TrimLeft(strings.TrimSpace(string(node.Info)), ". ")
	// the info string of a backtick fence can't contain backticks
	fenceChar := byte('`')
	if strings.IndexByte(info, '`') >= 0 {
		fenceChar = '~'
	}
	fenceLen := fenceLength(text, fenceChar)
	if fenceLen < 3 {
		fenceLen = 3
	}
	fence := strings.Repeat(string(fenceChar), fenceLen)
	r.outs(w, fence+info+"\n")
	r.out(w, text)
	if len(text) == 0 || text[len(text)-1] != '\n' {
		r.outs(w, "\n")
	}
	r.outs(w, fence+"\n\n")
}

// fenceLength returns a fence length longer than any run of c in text.
func fenceLength(text []byte, c byte) int {
	fenceLen, run := 1, 0
	for _, ch := range text {
		if ch == c {
			run++
			if run >= fenceLen {
				fenceLen = run + 1
			}
		} else {
			run = 0
		}
	}
	return fenceLen
}

func (r *Renderer) code(w io.Writer, node *ast.Code) {
	literal := node.Literal
	// there's no syntax for an empty code span: `` parses as literal backticks
	if len(literal) == 0 {
		return
	}
	// the delimiter must be longer than any backtick run in the literal because
	// the legacy parser closes a code span at the first run at least as long
	fenceLen := fenceLength(literal, '`')
	// the legacy parser turns a multi-line code span with a delimiter of 3
	// backticks into a code block
	if fenceLen == 3 {
		fenceLen = 4
	}
	fence := strings.Repeat("`", fenceLen)
	// padding keeps backticks at the edges from merging with the delimiter.
	// It also protects spaces at the edges: CommonMark strips one from each end
	// if both ends are spaces, after turning line endings into spaces, unless
	// the content is all spaces.
	first, last := literal[0], literal[len(literal)-1]
	padding := ""
	if first == '`' || last == '`' ||
		((first == ' ' || first == '\n') && (last == ' ' || last == '\n') && len(bytes.Trim(literal, " \n")) > 0) {
		padding = " "
	}
	r.outs(w, fence+padding+string(literal)+padding+fence)
}

func (r *Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		r.outs(w, strings.Repeat("#", node.Level)+" ")
		r.out(w, node.Literal)
	} else {
		if node.HeadingID != "" {
			r.outs(w, " {#"+node.HeadingID+"}")
		}
		r.outs(w, "\n\n")
	}
}

// prefixedBlock renders a block whose lines all start with prefix.
func (r *Renderer) prefixedBlock(w io.Writer, node ast.Node, entering bool, prefix string) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	lines := bytes.Split(r.renderBlocks(node, false), []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			r.outs(w, strings.TrimRight(prefix, " "))
		} else {
			r.outs(w, prefix)
			r.out(w, line)
		}
		r.outs(w, "\n")
	}
	r.outs(w, "\n")
	return ast.SkipChildren
}

func (r *Renderer) tableRow(w io.Writer, row ast.Node) {
	r.outs(w, "|")
	for _, child := range row.GetChildren() {
		cell, ok := child.(*ast.TableCell)
		if !ok {
			continue
		}
		wasInCell := r.inTableCell
		r.inTableCell = true
		content := bytes.ReplaceAll(r.renderBlocks(cell, true), []byte("\n"), []byte(" "))
		r.inTableCell = wasInCell
		r.outs(w, " ")
		r.out(w, content)
		r.outs(w, " |")
		// each extra | makes the cell span one more column
		if cell.ColSpan > 1 {
			r.outs(w, strings.Repeat("|", cell.ColSpan-1))
		}
	}
	r.outs(w, "\n")
}

// tableColumnAligns returns the alignment of each column, using the first row.
func tableColumnAligns(node *ast.Table) []ast.CellAlignFlags {
	var aligns []ast.CellAlignFlags
	for _, section := range node.GetChildren() {
		for _, row := range section.GetChildren() {
			for _, child := range row.GetChildren() {
				cell, ok := child.(*ast.TableCell)
				if !ok {
					continue
				}
				aligns = append(aligns, cell.Align)
				for i := 1; i < cell.ColSpan; i++ {
					aligns = append(aligns, cell.Align)
				}
			}
			return aligns
		}
	}
	return aligns
}

func (r *Renderer) table(w io.Writer, node *ast.Table, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	aligns := tableColumnAligns(node)
	writeDelimiter := func() {
		r.outs(w, "|")
		for _, align := range aligns {
			switch align {
			case ast.TableAlignmentLeft:
				r.outs(w, " :--- |")
			case ast.TableAlignmentRight:
				r.outs(w, " ---: |")
			case ast.TableAlignmentCenter:
				r.outs(w, " :---: |")
			default:
				r.outs(w, " --- |")
			}
		}
		r.outs(w, "\n")
	}
	wroteDelimiter := false
	for _, section := range node.GetChildren() {
		_, isHeader := section.(*ast.TableHeader)
		if !isHeader && !wroteDelimiter {
			writeDelimiter()
			wroteDelimiter = true
		}
		if _, isFooter := section.(*ast.TableFooter); isFooter {
			// the parser allows no spaces in this line and needs at least one |
			r.outs(w, strings.Repeat("===|", len(aligns))+"\n")
		}
		for _, row := range section.GetChildren() {
			r.tableRow(w, row)
		}
		if isHeader {
			writeDelimiter()
			wroteDelimiter = true
		}
	}
	r.outs(w, "\n")
	return ast.SkipChildren
}

func (r *Renderer) captionFigure(w io.Writer, node *ast.CaptionFigure, entering bool) ast.WalkStatus {
	if !entering {
		return ast.GoToNext
	}
	children := node.GetChildren()
	var caption *ast.Caption
	if len(children) > 0 {
		if c, ok := children[len(children)-1].(*ast.Caption); ok {
			caption = c
			children = children[:len(children)-1]
		}
	}
	// a code block, a quote or a table can have a caption directly after it,
	// anything else has to be in a figure block
	prefix, separator := "", "\n"
	if len(children) == 1 {
		switch children[0].(type) {
		case *ast.CodeBlock:
			prefix = "Figure: "
		case *ast.BlockQuote:
			// a line directly after a quote is a part of the quote
			prefix, separator = "Quote: ", "\n\n"
		case *ast.Table:
			prefix = "Table: "
		}
	}
	if prefix != "" {
		r.out(w, bytes.Trim(r.render(children[0]), "\n"))
		r.outs(w, separator)
	} else {
		prefix = "Figure: "
		r.outs(w, "!---\n")
		for _, child := range children {
			r.out(w, bytes.Trim(r.render(child), "\n"))
			r.outs(w, "\n\n")
		}
		r.outs(w, "!---\n")
	}
	if caption != nil {
		r.outs(w, prefix)
		r.out(w, bytes.TrimRight(r.renderBlocks(caption, true), " "))
		if node.HeadingID != "" {
			r.outs(w, " {#"+node.HeadingID+"}")
		}
		r.outs(w, "\n")
	}
	r.outs(w, "\n")
	return ast.SkipChildren
}

func (r *Renderer) documentMatter(w io.Writer, node *ast.DocumentMatter, entering bool) {
	if !entering {
		return
	}
	switch node.Matter {
	case ast.DocumentMatterFront:
		r.outs(w, "{frontmatter}\n\n")
	case ast.DocumentMatterMain:
		r.outs(w, "{mainmatter}\n\n")
	case ast.DocumentMatterBack:
		r.outs(w, "{backmatter}\n\n")
	}
}

func (r *Renderer) citation(w io.Writer, node *ast.Citation) {
	r.outs(w, "[")
	for i, destination := range node.Destination {
		if i > 0 {
			r.outs(w, "; ")
		}
		r.outs(w, "@")
		if i < len(node.Type) {
			switch node.Type[i] {
			case ast.CitationTypeNormative:
				r.outs(w, "!")
			case ast.CitationTypeSuppressed:
				r.outs(w, "-")
			}
		}
		r.out(w, destination)
		if i < len(node.Suffix) && len(node.Suffix[i]) > 0 {
			r.outs(w, ", "+string(node.Suffix[i]))
		}
	}
	r.outs(w, "]")
}

func (r *Renderer) index(w io.Writer, node *ast.Index) {
	r.outs(w, "(!")
	if node.Primary {
		r.outs(w, "!")
	}
	r.out(w, node.Item)
	if len(node.Subitem) > 0 {
		r.outs(w, ", "+string(node.Subitem))
	}
	r.outs(w, ")")
}

// RenderNode renders markdown node
func (r *Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Text:
		r.text(w, node)
	case *ast.Emph:
		r.outs(w, "*")
	case *ast.Strong:
		r.outs(w, "**")
	case *ast.Del:
		r.outs(w, "~~")
	case *ast.Link:
		r.link(w, node, entering)
	case *ast.Image:
		r.image(w, node, entering)
	case *ast.Code:
		r.code(w, node)
	case *ast.CodeBlock:
		r.codeBlock(w, node)
	case *ast.Document:
	case *ast.Paragraph:
		r.para(w, node, entering)
	case *ast.HTMLSpan:
		r.htmlSpan(w, node)
	case *ast.HTMLBlock:
		r.htmlBlock(w, node)
	case *ast.Heading:
		r.heading(w, node, entering)
	case *ast.List:
		return r.list(w, node, entering)
	case *ast.ListItem:
		return r.listItem(w, node, entering)
	case *ast.Footnotes:
	case *ast.ReferenceDefinition:
		if entering {
			r.outs(w, "[")
			r.out(w, node.Label)
			r.outs(w, "]: ")
			r.out(w, node.Destination)
			if len(node.Title) > 0 {
				r.outs(w, " "+quoteTitle(node.Title))
			}
			r.outs(w, "\n")
		}
	case *ast.BlockQuote:
		return r.prefixedBlock(w, node, entering, "> ")
	case *ast.Aside:
		return r.prefixedBlock(w, node, entering, "A> ")
	case *ast.HorizontalRule:
		r.outs(w, "---\n\n")
	case *ast.Hardbreak:
		r.outs(w, "  \n")
	case *ast.Softbreak:
		r.outs(w, "\n")
	case *ast.NonBlockingSpace:
		r.outs(w, "&nbsp;")
	case *ast.Math:
		r.outs(w, "$"+string(node.Literal)+"$")
	case *ast.MathBlock:
		if entering {
			r.outs(w, "$$"+string(node.Literal)+"$$\n\n")
		}
		return ast.SkipChildren
	case *ast.Subscript:
		r.outs(w, "~"+string(node.Literal)+"~")
	case *ast.Superscript:
		r.outs(w, "^"+string(node.Literal)+"^")
	case *ast.Table:
		return r.table(w, node, entering)
	case *ast.CaptionFigure:
		return r.captionFigure(w, node, entering)
	case *ast.DocumentMatter:
		r.documentMatter(w, node, entering)
	case *ast.CrossReference:
		if entering {
			r.outs(w, "(#"+string(node.Destination))
			if len(node.Suffix) > 0 {
				r.outs(w, ", "+string(node.Suffix))
			}
			r.outs(w, ")")
		}
		return ast.SkipChildren
	case *ast.Citation:
		r.citation(w, node)
	case *ast.Callout:
		r.outs(w, "<<"+string(node.ID)+">>")
	case *ast.Index:
		r.index(w, node)
	default:
		// a node we don't know about, like a custom node added by a parser hook
		if leaf := node.AsLeaf(); leaf != nil {
			r.out(w, leaf.Literal)
		}
	}
	return ast.GoToNext
}

// RenderHeader renders header
func (r *Renderer) RenderHeader(_ io.Writer, _ ast.Node) {}

// RenderFooter renders footer
func (r *Renderer) RenderFooter(w io.Writer, _ ast.Node) {
	if r.C == nil || r.C.Flags&renderLinksInFooter == 0 || len(r.linkcache) == 0 {
		return
	}
	links := make([]string, 0, len(r.linkcache))
	for link := range r.linkcache {
		links = append(links, link)
	}
	sort.Strings(links)
	for _, link := range links {
		r.outs(w, "\n")
		r.outs(w, link)
	}
	r.outs(w, "\n")
	// the renderer can be reused and links must not leak into the next document
	r.linkcache = nil
}

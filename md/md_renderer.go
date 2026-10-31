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

func (r *Renderer) list(w io.Writer, node *ast.List, entering bool) {
	if entering {
		r.listDepth++
		flags := node.ListFlags
		if flags&ast.ListTypeOrdered != 0 {
			if r.orderedListCounter == nil {
				r.orderedListCounter = map[int]int{}
			}
			r.orderedListCounter[r.listDepth] = 1
			if node.Start > 0 {
				r.orderedListCounter[r.listDepth] = node.Start
			}
		}
	} else {
		r.listDepth--
		if _, ok := node.Parent.(*ast.ListItem); !ok {
			fmt.Fprintf(w, "\n")
		}
	}
}

func (r *Renderer) listItem(w io.Writer, node *ast.ListItem, entering bool) {
	if !entering {
		return
	}
	io.WriteString(w, strings.Repeat(" ", (r.listDepth-1)*listIndentSize))
	if node.ListFlags&ast.ListTypeOrdered != 0 {
		delimiter := byte('.')
		if node.Delimiter == ')' {
			delimiter = ')'
		}
		fmt.Fprintf(w, "%d%c ", r.orderedListCounter[r.listDepth], delimiter)
		r.orderedListCounter[r.listDepth]++
	} else {
		io.WriteString(w, string(node.BulletChar)+" ")
	}
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
		r.list(w, node, entering)
	case *ast.ListItem:
		r.listItem(w, node, entering)
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
	default:
		panic(fmt.Sprintf("node %T NYI", node))
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

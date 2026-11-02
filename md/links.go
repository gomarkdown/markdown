package md

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// backslashEscape returns text with a backslash before each byte for which
// needsEscape returns true.
func backslashEscape(text []byte, needsEscape func(i int, c byte) bool) []byte {
	var out []byte
	for i, c := range text {
		if needsEscape(i, c) {
			if out == nil {
				out = append(make([]byte, 0, len(text)+8), text[:i]...)
			}
			out = append(out, '\\')
		}
		if out != nil {
			out = append(out, c)
		}
	}
	if out == nil {
		return text
	}
	return out
}

// hasBalancedParens returns true if the parentheses in text are balanced, which
// is when a link destination can have them unescaped.
func hasBalancedParens(text []byte) bool {
	depth := 0
	for _, c := range text {
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// linkTarget returns the destination and title of a link as written after the
// label. Both parsers unescape destinations. CommonMark also decodes entities
// and needs <> around a destination that is empty or has spaces.
func linkTarget(dest, title []byte, node ast.Node) string {
	commonMark := isCommonMarkDoc(node)
	var target string
	if commonMark && (len(dest) == 0 || bytes.ContainsAny(dest, " \t\r\n")) {
		target = "<" + string(backslashEscape(dest, func(_ int, c byte) bool {
			return c == '\\' || c == '&' || c == '<' || c == '>'
		})) + ">"
	} else {
		parens := !hasBalancedParens(dest)
		target = string(backslashEscape(dest, func(i int, c byte) bool {
			return c == '\\' || (c == '&' && commonMark) || (c == '<' && i == 0) ||
				((c == '(' || c == ')') && parens)
		}))
	}
	if len(title) != 0 {
		target += " " + quoteTitle(title, commonMark)
	}
	return target
}

// refTarget returns the destination and title of a reference definition. The
// legacy parser keeps them in source form.
func refTarget(ref *ast.ReferenceDefinition) string {
	if isCommonMarkDoc(ref) {
		return linkTarget(ref.Destination, ref.Title, ref)
	}
	target := string(ref.Destination)
	if len(ref.Title) != 0 {
		target += " " + quoteTitle(ref.Title, false)
	}
	return target
}

// quoteTitle returns a link title with its delimiters. CommonMark titles have
// decoded escapes and entities; the legacy parser leaves them in source form.
func quoteTitle(title []byte, commonMark bool) string {
	if commonMark {
		title = backslashEscape(title, func(_ int, c byte) bool { return c == '\\' || c == '&' })
	}
	if !bytes.Contains(title, []byte(`"`)) {
		return `"` + string(title) + `"`
	}
	if !bytes.Contains(title, []byte(`'`)) {
		return `'` + string(title) + `'`
	}
	return `"` + string(bytes.ReplaceAll(title, []byte(`"`), []byte(`\"`))) + `"`
}

func (r *Renderer) image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		r.outs(w, "![")
		return
	}
	r.outs(w, "](")
	r.outs(w, linkTarget(node.Destination, node.Title, node))
	r.outs(w, ")")
}

func linkPlainText(node *ast.Link) []byte {
	var text []byte
	for _, child := range node.GetChildren() {
		literal, ok := child.(*ast.Text)
		if !ok {
			return nil
		}
		text = append(text, literal.Literal...)
	}
	return text
}

// footerLinkLabel returns the label for the definition of a link with the given
// rendered text. A label can be the text unless it has something that isn't
// valid in a label in all parsers.
func footerLinkLabel(text []byte) string {
	if len(text) == 0 || text[0] == '^' || bytes.ContainsAny(text, "[]\\\n") {
		return "link"
	}
	return string(text)
}

// normalizeLabel returns the form in which labels are compared by parsers.
func normalizeLabel(label string) string {
	return strings.ToLower(strings.Join(strings.Fields(label), " "))
}

// reserveDefinedLabels records labels of definitions already in the document, as
// the first definition of a label is the one that counts.
func (r *Renderer) reserveDefinedLabels(doc ast.Node) {
	r.footerLabels = map[string]string{}
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if ref, ok := node.(*ast.ReferenceDefinition); ok && entering {
			r.footerLabels[normalizeLabel(string(ref.Label))] = refTarget(ref)
		}
		return ast.GoToNext
	})
}

// uniqueFooterLabel returns a label based on label that refers to target, adding
// a definition to the footer unless there already is one.
func (r *Renderer) uniqueFooterLabel(label, target string) string {
	if r.footerLabels == nil {
		r.footerLabels = map[string]string{}
	}
	unique := label
	for i := 2; ; i++ {
		existing, ok := r.footerLabels[normalizeLabel(unique)]
		if ok && existing == target {
			return unique
		}
		if !ok {
			break
		}
		unique = label + "-" + strconv.Itoa(i)
	}
	r.footerLabels[normalizeLabel(unique)] = target
	if r.linkcache == nil {
		r.linkcache = make(map[string]bool)
	}
	r.linkcache["["+unique+"]: "+target] = true
	return unique
}

// footerLink renders a link whose destination goes to the footer.
func (r *Renderer) footerLink(w io.Writer, node *ast.Link) {
	var text []byte
	for _, child := range node.GetChildren() {
		text = append(text, r.render(child)...)
	}
	target := linkTarget(node.Destination, node.Title, node)
	label := r.uniqueFooterLabel(footerLinkLabel(text), target)

	r.outs(w, "[")
	r.out(w, text)
	r.outs(w, "]")
	if label != string(text) {
		r.outs(w, "["+label+"]")
		return
	}
	if shortcutNeedsBrackets(node) {
		r.outs(w, "[]")
	}
}

// shortcutNeedsBrackets returns true if a shortcut reference [a] needs to be
// written as [a][] because of what follows: ( would make it an inline link, : a
// definition, and the legacy parser takes a following [b] as its label even
// after a space.
func shortcutNeedsBrackets(node *ast.Link) bool {
	next := nextNonEmptyNode(node)
	if text, ok := next.(*ast.Text); ok {
		if text.Literal[0] == '(' || text.Literal[0] == ':' {
			return true
		}
		if len(bytes.TrimSpace(text.Literal)) > 0 {
			return false
		}
		next = nextNonEmptyNode(text)
	}
	_, isLink := next.(*ast.Link)
	return isLink
}

// canBeInFooter returns true if the destination of a link can be in a reference
// definition. The legacy parser doesn't take one that is empty or has spaces.
func canBeInFooter(node *ast.Link) bool {
	return isCommonMarkDoc(node) ||
		(len(node.Destination) != 0 && !bytes.ContainsAny(node.Destination, " \t\r\n"))
}

func (r *Renderer) link(w io.Writer, node *ast.Link, entering bool) ast.WalkStatus {
	if node.NoteID != 0 {
		if entering {
			r.outs(w, "[^"+string(node.Destination)+"]")
		}
		return ast.SkipChildren
	}
	if r.C != nil && r.C.Flags&renderLinksInFooter != 0 && canBeInFooter(node) {
		if entering {
			r.footerLink(w, node)
		}
		return ast.SkipChildren
	}
	if entering {
		r.outs(w, "[")
		return ast.GoToNext
	}
	if len(node.DeferredID) > 0 {
		if bytes.EqualFold(linkPlainText(node), node.DeferredID) {
			r.outs(w, "]")
			if shortcutNeedsBrackets(node) {
				r.outs(w, "[]")
			}
		} else {
			r.outs(w, "]["+string(escape(node.DeferredID))+"]")
		}
		return ast.GoToNext
	}

	r.outs(w, "](")
	r.outs(w, linkTarget(node.Destination, node.Title, node))
	r.outs(w, ")")
	return ast.GoToNext
}

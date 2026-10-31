package md

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// quoteTitle returns a link title with its delimiters. The legacy parser doesn't
// unescape titles, so a backslash is only used if both kinds of quotes are in it.
func quoteTitle(title []byte) string {
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
	r.out(w, escape(node.Destination))
	if len(node.Title) != 0 {
		r.outs(w, " "+quoteTitle(node.Title))
	}
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
			target := string(ref.Destination)
			if len(ref.Title) != 0 {
				target += " " + quoteTitle(ref.Title)
			}
			r.footerLabels[normalizeLabel(string(ref.Label))] = target
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
	target := string(escape(node.Destination))
	if len(node.Title) != 0 {
		target += " " + quoteTitle(node.Title)
	}
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

func (r *Renderer) link(w io.Writer, node *ast.Link, entering bool) ast.WalkStatus {
	if node.NoteID != 0 {
		if entering {
			r.outs(w, "[^"+string(node.Destination)+"]")
		}
		return ast.SkipChildren
	}
	if r.C != nil && r.C.Flags&renderLinksInFooter != 0 {
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
	r.out(w, escape(node.Destination))
	if len(node.Title) != 0 {
		r.outs(w, " "+quoteTitle(node.Title))
	}
	r.outs(w, ")")
	return ast.GoToNext
}

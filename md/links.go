package md

import (
	"bytes"
	"io"

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

// footerLink renders a link whose destination goes to the footer.
func (r *Renderer) footerLink(w io.Writer, node *ast.Link) {
	var text []byte
	for _, child := range node.GetChildren() {
		text = append(text, r.render(child)...)
	}
	label := footerLinkLabel(text)
	definition := "[" + label + "]: " + string(escape(node.Destination))
	if len(node.Title) != 0 {
		definition += " " + quoteTitle(node.Title)
	}
	if r.linkcache == nil {
		r.linkcache = make(map[string]bool)
	}
	r.linkcache[definition] = true

	r.outs(w, "[")
	r.out(w, text)
	r.outs(w, "]")
	if label != string(text) {
		r.outs(w, "["+label+"]")
		return
	}
	// [a] followed by these would be an inline link or a definition
	if next, ok := nextNonEmptyNode(node).(*ast.Text); ok && (next.Literal[0] == '(' || next.Literal[0] == ':') {
		r.outs(w, "[]")
	}
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

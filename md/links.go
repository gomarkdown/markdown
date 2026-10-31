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

func (r *Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if node.NoteID != 0 {
		if entering {
			r.outs(w, "[^"+string(node.Destination)+"]")
		}
		return
	}
	if entering {
		r.outs(w, "[")
		return
	}
	if len(node.DeferredID) > 0 && (r.C == nil || r.C.Flags&renderLinksInFooter == 0) {
		if bytes.EqualFold(linkPlainText(node), node.DeferredID) {
			r.outs(w, "]")
		} else {
			r.outs(w, "]["+string(escape(node.DeferredID))+"]")
		}
		return
	}

	link := string(escape(node.Destination))
	title := ""
	if len(node.Title) != 0 {
		title = " " + quoteTitle(node.Title)
	}
	if r.C == nil || r.C.Flags&renderLinksInFooter == 0 {
		r.outs(w, "](")
		r.outs(w, link)
		r.outs(w, title)
		r.outs(w, ")")
		return
	}

	r.outs(w, "]")
	child, _ := ast.GetFirstChild(node).(*ast.Text)
	definition := "[" + string(escape(child.Literal)) + "]: " + link + title
	if r.linkcache == nil {
		r.linkcache = make(map[string]bool)
	}
	r.linkcache[definition] = true
}

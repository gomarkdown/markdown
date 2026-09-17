package md

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

func (r *Renderer) image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		r.outs(w, "![")
		return
	}
	r.outs(w, "](")
	r.out(w, escape(node.Destination))
	if len(node.Title) != 0 {
		r.outs(w, ` "`+string(node.Title)+`"`)
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
	title := string(node.Title)
	if r.C == nil || r.C.Flags&renderLinksInFooter == 0 {
		r.outs(w, "](")
		r.outs(w, link)
		if title != "" {
			r.outs(w, ` "`+title+`"`)
		}
		r.outs(w, ")")
		return
	}

	r.outs(w, "]")
	child, _ := ast.GetFirstChild(node).(*ast.Text)
	definition := "[" + string(escape(child.Literal)) + "]: " + link
	if title != "" {
		definition += ` "` + title + `"`
	}
	if r.linkcache == nil {
		r.linkcache = make(map[string]bool)
	}
	r.linkcache[definition] = true
}

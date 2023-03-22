package main

// example for https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

import (
	"fmt"
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

// an actual rendering of Paragraph is more complicated
func renderParagraph(w io.Writer, p *ast.Paragraph, entering bool) {
	if entering {
		io.WriteString(w, "<div>")
	} else {
		io.WriteString(w, "</div>")
	}
}

func myRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if para, ok := node.(*ast.Paragraph); ok {
		renderParagraph(w, para, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func newCustomizedRender() *html.Renderer {
	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: myRenderHook,
	}
	return html.NewRenderer(opts)
}

var mds = `foo`

func renderHookExmple() {
	md := []byte(mds)

	renderer := newCustomizedRender()
	html := markdown.ToHTML(md, nil, renderer)

	fmt.Printf("--- Markdown:\n%s\n\n--- HTML:\n%s\n", md, html)
}

func main() {
	renderHookExmple()
}

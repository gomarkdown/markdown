package main

// example for https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"bytes"
	"io"
	"strings"
)

type Gallery struct {
	ast.Leaf
	ImageURLS []string
}

var gallery = []byte(":gallery\n")

func parseGallery(data []byte) (ast.Node, []byte, int) {
	if !bytes.HasPrefix(data, gallery) {
		return nil, nil, 0
	}
	fmt.Printf("Found a gallery!\n\n")
	i := len(gallery)
	// find empty line
	// TODO: should also consider end of document
	end := bytes.Index(data[i:], []byte("\n\n"))
	if end < 0 {
		return nil, data, 0
	}
	end = end + i
	lines := string(data[i:end])
	parts := strings.Split(lines, "\n")
	res := &Gallery{
		ImageURLS: parts,
	}
	return res, nil, end
}

func parserHook(data []byte) (ast.Node, []byte, int) {
	if node, d, n := parseGallery(data); node != nil {
		return node, d, n
	}
	return nil, nil, 0
}

func newMarkdownParser() *parser.Parser {
	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)
	p.Opts.ParserHook = parserHook
	return p
}

func galleryRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if _, ok := node.(*Gallery); ok {
		if entering {
			// note: just for illustration purposes
			// actual implemenation of gallery in HTML / JavaScript is long
			io.WriteString(w, "\n<gallery></gallery>\n\n")
		}
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func newGalleryRender() *html.Renderer {
	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: galleryRenderHook,
	}
	return html.NewRenderer(opts)
}

var mds = `document

:gallery
/img/image-1.png
/img/image-2.png

Rest of the document.`

func parserHookExample() {
	md := []byte(mds)

	p := newMarkdownParser()
	doc := p.Parse([]byte(md))

	renderer := newGalleryRender()
	html := markdown.Render(doc, renderer)

	fmt.Printf("--- Markdown:\n%s\n\n--- HTML:\n%s\n", md, html)
}

func main() {
	parserHookExample()
}

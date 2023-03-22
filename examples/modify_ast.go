package main

// example for https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"strings"
)

func modifyAst(doc ast.Node) ast.Node {
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if img, ok := node.(*ast.Image); ok && entering {
			attr := img.Attribute
			if attr == nil {
				attr = &ast.Attribute{}
			}
			// TODO: might be duplicate
			attr.Classes = append(attr.Classes, []byte("blog-img"))
			img.Attribute = attr
		}

		if link, ok := node.(*ast.Link); ok && entering {
			isExternalURI := func(uri string) bool {
				return (strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://")) && !strings.Contains(uri, "blog.kowalczyk.info")
			}
			if isExternalURI(string(link.Destination)) {
				link.AdditionalAttributes = append(link.AdditionalAttributes, `target="_blank"`)
			}
		}

		return ast.GoToNext
	})
	return doc
}

var mds = `[link](http://example.com)`

func modifyAstExample() {
	md := []byte(mds)

	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	doc = modifyAst(doc)

	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	html := markdown.Render(doc, renderer)

	fmt.Printf("-- Markdown:\n%s\n\n--- HTML:\n%s\n", md, html)
}

func main() {
	modifyAstExample()
}

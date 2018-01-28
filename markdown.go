// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.

package markdown

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Renderer is an interface for implementing custom renderers.
type Renderer interface {
	// RenderNode is the main rendering method. It will be called once for
	// every leaf node and twice for every non-leaf node (first with
	// entering=true, then with entering=false). The method should write its
	// rendition of the node to writer w.
	RenderNode(w io.Writer, node *ast.Node, entering bool) ast.WalkStatus

	// RenderHeader is a method that allows the renderer to produce some
	// content preceding the main body of the output document. The header is
	// understood in the broad sense here. For example, the default HTML
	// renderer will write not only the HTML document preamble, but also the
	// table of contents if it was requested.
	//
	// The method will be passed an entire document tree, in case a particular
	// implementation needs to inspect it to produce output.
	//
	// The output should be written to the supplied writer w. If your
	// implementation has no header to write, supply an empty implementation.
	RenderHeader(w io.Writer, ast *ast.Node)

	// RenderFooter is a symmetric counterpart of RenderHeader.
	RenderFooter(w io.Writer, ast *ast.Node)
}

// Render uses renderer to convert parsed markdown document into a different format.
// For example to convert into HTML, use html.Renderer
func Render(doc *ast.Node, renderer Renderer) []byte {
	var buf bytes.Buffer
	renderer.RenderHeader(&buf, doc)
	doc.WalkFunc(func(node *ast.Node, entering bool) ast.WalkStatus {
		return renderer.RenderNode(&buf, node, entering)
	})
	renderer.RenderFooter(&buf, doc)
	return buf.Bytes()
}

// ToHTML converts markdownDoc to HTML.
//
// You can optionally pass a parser and renderer. This allows to customize
// a parser, use a customized html render or use completely custom renderer.
//
// If you pass nil for both, we use parser configured with parser.CommonExtensions
// and html.Renderer configured with html.CommonFlags.
func ToHTML(markdownDoc []byte, p *parser.Parser, renderer Renderer) []byte {
	if p == nil {
		p = parser.NewWithExtensions(parser.CommonExtensions)
	}
	if renderer == nil {
		opts := html.RendererOptions{
			Flags: html.CommonFlags,
		}
		renderer = html.NewRenderer(opts)
	}
	doc := p.Parse(markdownDoc)
	return Render(doc, renderer)
}

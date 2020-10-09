/*
Package markdown implements markdown parser and HTML renderer.

It parses markdown into AST format which can be serialized to HTML
(using html.Renderer) or possibly other formats (using alternate renderers).

Convert markdown to HTML

The simplest way to convert markdown document to HTML

  md := []byte("## markdown document")
  html := markdown.ToHTML(md, nil, nil)

Customizing parsing and HTML rendering

You can customize parser and HTML renderer:

	import (
		"github.com/moorara/markdown/parser"
		"github.com/moorara/markdown/renderer"
		"github.com/moorara/markdown"
	)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	md := []byte("markdown text")
	html := markdown.ToHTML(md, p, renderer)
*/
package markdown

import (
	"bytes"
	"io"

	"github.com/moorara/markdown/ast"
	"github.com/moorara/markdown/html"
	"github.com/moorara/markdown/parser"
)

// Renderer is an interface for implementing custom renderers.
type Renderer interface {
	// RenderNode renders markdown node to w.
	// It's called once for a leaf node.
	// It's called twice for non-leaf nodes:
	// * first with entering=true
	// * then with entering=false
	//
	// Return value is a way to tell the calling walker to adjust its walk
	// pattern: e.g. it can terminate the traversal by returning Terminate. Or it
	// can ask the walker to skip a subtree of this node by returning SkipChildren.
	// The typical behavior is to return GoToNext, which asks for the usual
	// traversal to the next node.
	RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus

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
	RenderHeader(w io.Writer, ast ast.Node)

	// RenderFooter is a symmetric counterpart of RenderHeader.
	RenderFooter(w io.Writer, ast ast.Node)
}

// Parse parsers a markdown document using provided parser. If parser is nil,
// we use parser configured with parser.CommonExtensions.
//
// It returns AST (abstract syntax tree) that can be converted to another
// format using Render function.
func Parse(markdown []byte, p *parser.Parser) ast.Node {
	if p == nil {
		p = parser.New()
	}
	return p.Parse(markdown)
}

// Render uses renderer to convert parsed markdown document into a different format.
//
// To convert to HTML, pass html.Renderer
func Render(doc ast.Node, renderer Renderer) []byte {
	var buf bytes.Buffer
	renderer.RenderHeader(&buf, doc)
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
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
func ToHTML(markdown []byte, p *parser.Parser, renderer Renderer) []byte {
	doc := Parse(markdown, p)
	if renderer == nil {
		opts := html.RendererOptions{
			Flags: html.CommonFlags,
		}
		renderer = html.NewRenderer(opts)
	}
	return Render(doc, renderer)
}

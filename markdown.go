// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.

package markdown

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/htmlrenderer"
)

// Renderer is an interface for implementing custom renderers.
//
// This package provides HTMLRenderer for markdown => HTML conversion.
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

// ToHTML converts a markdown text in input and converts it to HTML.
//
// You can optionally pass a parser and renderer, which allows to customize
// a parser, a render or provide a renderer other than HTMLRenderer.
//
// If you pass nil for both, we convert with CommonExtensions for
// the parser and HTMLRenderer with CommonFlags for renderer
func ToHTML(input []byte, parser *Parser, renderer Renderer) []byte {
	if parser == nil {
		parser = NewParserWithExtensions(CommonExtensions)
	}
	if renderer == nil {
		params := htmlrenderer.RendererOptions{
			Flags: htmlrenderer.CommonFlags,
		}
		renderer = htmlrenderer.NewHTMLRenderer(params)
	}
	parser.Parse(input)
	return parser.Render(renderer)
}

// Markdown Processor for Go
// Available at https://github.com/gomarkdown/markdown
//
// Copyright © 2011 Russ Ross <russ@russross.com>.
// Distributed under the Simplified BSD License.
// See README.md for details.

package markdown

import (
	"io"
)

// Version string of the package. Appears in the rendered document when
// CompletePage flag is on.
const Version = "2.0"

// Renderer is the rendering interface. This is mostly of interest if you are
// implementing a new rendering format.
//
// Only an HTML implementation is provided in this repository, see the README
// for external implementations.
type Renderer interface {
	// RenderNode is the main rendering method. It will be called once for
	// every leaf node and twice for every non-leaf node (first with
	// entering=true, then with entering=false). The method should write its
	// rendition of the node to the supplied writer w.
	RenderNode(w io.Writer, node *Node, entering bool) WalkStatus

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
	RenderHeader(w io.Writer, ast *Node)

	// RenderFooter is a symmetric counterpart of RenderHeader.
	RenderFooter(w io.Writer, ast *Node)
}

// ToHTML converts a markdown text in input and converts it to HTML.
//
// You can optionally pass a parser and renderer, which allows to customize
// a parser, a render or provide a renderer other than HTMLRenderer.
//
// If you pass nil for both, we convert with CommonExtensions for
// the parser and HTMLRenderer with CommonHTMLFlags.
func ToHTML(input []byte, parser *Parser, renderer Renderer) []byte {
	if parser == nil {
		parser = NewParserWithExtensions(CommonExtensions)
	}
	if renderer == nil {
		params := HTMLRendererParameters{
			Flags: CommonHTMLFlags,
		}
		renderer = NewHTMLRenderer(params)
	}
	parser.Parse(input)
	return parser.Render(renderer)
}

/*
Package notion implements a block builder from markdown.

# Configuring and customizing a renderer

A renderer can be configured with multiple options:

	import "github.com/gomarkdown/markdown/notion"

	opts := notion.RendererOptions{}
	renderer := notion.NewRenderer(opts)

You can also re-use most of the logic and customize rendering of selected nodes
by providing node render hook.
This is most useful for rendering nodes that allow for design choices, like
links or code blocks.

	import (
		"github.com/gomarkdown/markdown/notion"
		"github.com/gomarkdown/markdown/ast"
	)

	// a very dummy render hook that will output "code_replacements" instead of
	// <code>${content}</code> emitted by notion.Renderer
	func renderHookCodeBlock(w io.Writer, node ast.Node) (ast.WalkStatus, bool) {
		_, ok := node.(*ast.CodeBlock)
		if !ok {
			return ast.GoToNext, false
		}
		io.WriteString(w, "code_replacement")
		return ast.GoToNext, true
	}

	opts := notion.RendererOptions{
		RenderNodeHook: renderHookCodeBlock,
	}
	renderer := notion.NewRenderer(opts)
*/
package notion

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

	// a very dummy render hook that will output "> rendered in a custom func!" to each code block
	// it renders, and then adds that block to the internal block slice! (needed to post values to notion)
	func renderHookCodeBlock(w io.Writer, node ast.Node, render *Renderer) (ast.WalkStatus, bool) {
		_, ok := node.(*ast.CodeBlock)
		if !ok {
			return ast.GoToNext, false
		}
		if err != nil {
		  fmt.Println("i can't do shit here")
	  }
		textBlock, err := GetBlock[CodeBlock](CodeType)
		nodeContentAsRichText := RichTextFromString(string(codeBlock.Literal) + "\n> rendered in a custom func!")
		textBlock.Code.RichText = append(textBlock.Code.RichText, nodeContentAsRichText)
		r.AddBlock(textBlock)
		return ast.GoToNext, true
	}

	opts := notion.RendererOptions{
		RenderNodeHook: renderHookCodeBlock,
	}
	renderer := notion.NewRenderer(opts)
*/
package notion

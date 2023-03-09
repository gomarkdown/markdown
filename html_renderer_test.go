package markdown

import (
	"io"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func renderHookEmpty(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	return ast.GoToNext, true
}

func TestRenderNodeHookEmpty(t *testing.T) {
	tests := []string{
		"[foo](gopher://foo.bar)",
		"",

		"[foo](mailto://bar/)\n",
		"",
	}

	htmlParams := html.RendererOptions{
		RenderNodeHook: renderHookEmpty,
	}
	params := TestParams{
		RendererOptions: htmlParams,
	}
	doTestsParam(t, tests, params)
}

func renderHookCodeBlock(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	_, ok := node.(*ast.CodeBlock)
	if !ok {
		return ast.GoToNext, false
	}
	io.WriteString(w, "code_replacement")
	return ast.GoToNext, true
}

func TestRenderNodeHookCode(t *testing.T) {
	tests := []string{
		"a\n```go\ncode\n```\nb",
		"<p>a</p>\ncode_replacement\n<p>b</p>\n",
	}
	opts := html.RendererOptions{
		RenderNodeHook: renderHookCodeBlock,
	}
	params := TestParams{
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

func TestTagParagraphCode(t *testing.T) {
	tests := []string{
		"test",
		"<div>test</div>\n",
	}
	opts := html.RendererOptions{
		ParagraphTag: "div",
	}
	params := TestParams{
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

func TestRenderNodeHookLinkAttrs(t *testing.T) {
	tests := []string{
		`[Click Me](gopher://foo.bar "Click Me")`,
		`<p><a class="button" href="gopher://foo.bar" target="_blank" title="Click Me">Click Me</a></p>` + "\n",
	}
	opts := html.RendererOptions{
		RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			link, isLink := node.(*ast.Link)
			if isLink {
				link.AdditionalAttributes = append(link.AdditionalAttributes, `class="button"`)
			}

			return ast.GoToNext, false
		},
	}
	params := TestParams{
		Flags:           html.HrefTargetBlank,
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

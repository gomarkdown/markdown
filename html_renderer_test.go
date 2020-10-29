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

func renderImagePre(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if _, ok := node.(*ast.Image); ok && entering {
		_, _ = io.WriteString(w, `<figure>`)
	}

	return 0, false
}

func renderImagePost(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if _, ok := node.(*ast.Image); ok && !entering {
		_, _ = io.WriteString(w, `</figure>`)
	}

	return ast.GoToNext
}

func TestRenderImagePrePost(t *testing.T) {
	tests := []string{
		"![some-image](image.png)",
		"<p><figure><img src=\"image.png\" alt=\"some-image\" /></figure></p>\n",
	}
	opts := html.RendererOptions{
		RenderNodeHook:     renderImagePre,
		RenderNodeHookPost: renderImagePost,
	}
	params := TestParams{
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

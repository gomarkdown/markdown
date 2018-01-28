package markdown

import (
	"io"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func renderHookEmpty(w io.Writer, node *ast.Node, entering bool) (ast.WalkStatus, bool) {
	return ast.GoToNext, true
}

func TestRenderNodeHookEmpty(t *testing.T) {
	t.Parallel()
	tests := []string{
		"[foo](gopher://foo.bar)",
		"",

		"[foo](mailto://bar/)\n",
		"",
	}

	htmlParams := HTMLRendererParameters{
		RenderNodeHook: renderHookEmpty,
	}
	params := TestParams{
		HTMLRendererParameters: htmlParams,
	}
	doTestsParam(t, tests, params)
}

func renderHookCodeBlock(w io.Writer, node *ast.Node, entering bool) (ast.WalkStatus, bool) {
	_, ok := node.Data.(*ast.CodeBlockData)
	if !ok {
		return ast.GoToNext, false
	}
	io.WriteString(w, "code_replacement")
	return ast.GoToNext, true
}

func TestRenderNodeHookCode(t *testing.T) {
	t.Parallel()
	tests := []string{
		"a\n```go\ncode\n```\nb",
		"<p>a</p>\ncode_replacement\n<p>b</p>\n",
	}
	htmlParams := HTMLRendererParameters{
		RenderNodeHook: renderHookCodeBlock,
	}
	params := TestParams{
		HTMLRendererParameters: htmlParams,
		extensions:             CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

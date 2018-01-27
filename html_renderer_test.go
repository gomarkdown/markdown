package markdown

import (
	"io"
	"testing"
)

func renderHookEmpty(w io.Writer, node *Node, entering bool) (WalkStatus, bool) {
	return GoToNext, true
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

func renderHookCodeBlock(w io.Writer, node *Node, entering bool) (WalkStatus, bool) {
	_, ok := node.Data.(*CodeBlockData)
	if !ok {
		return GoToNext, false
	}
	io.WriteString(w, "code_replacement")
	return GoToNext, true
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

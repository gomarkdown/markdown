package markdown

import (
	"io"
	"strings"
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

func TestRenderNodeHookLinkAttrs(t *testing.T) {
	tests := []string{
		`[Click Me](gopher://foo.bar "button Click Me")`,
		`<p><a class="mybuttoncls" href="gopher://foo.bar" target="_blank" title="Click Me">Click Me</a></p>` + "\n",
	}
	opts := html.RendererOptions{
		RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			link, isLink := node.(*ast.Link)
			if isLink {
				title := string(link.Title)
				if strings.HasPrefix(title, "button ") {
					link.Title = []byte(strings.TrimPrefix(title, "button "))
					link.Attribute = &ast.Attribute{}
					link.Classes = append(link.Classes, []byte("mybuttoncls"))
				}
			}

			return ast.GoToNext, false
		},
	}
	params := TestParams{
		Flags:          html.HrefTargetBlank,
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

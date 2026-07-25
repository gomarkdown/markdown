package markdown

import (
	"bytes"
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

// GHSA-gc99-qr5c-98ff: heading IDs and fenced-code info must be HTML-escaped
// when emitted into attributes so untrusted markdown cannot inject attributes.
func TestAttributeEscapeHeadingIDAndFencedInfo(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantNot string
	}{
		{
			name:    "heading-id-quote-breakout",
			input:   `# Heading {#x" onclick="alert(1)}` + "\n",
			want:    `id="x&quot; onclick=&quot;alert(1)"`,
			wantNot: `onclick="alert(1)"`,
		},
		{
			name:    "fenced-info-quote-breakout",
			input:   "```x\"onclick=\"alert(1)\ncode\n```\n",
			want:    `class="language-x&quot;onclick=&quot;alert(1)"`,
			wantNot: `onclick="alert(1)"`,
		},
		{
			name:    "fenced-info-entity-bypass",
			input:   "```x&#x22;onclick=&#x22;alert(1)\ncode\n```\n",
			want:    `class="language-x&quot;onclick=&quot;alert(1)"`,
			wantNot: `onclick="alert(1)"`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Default ToHTML path (CommonExtensions + CommonFlags).
			got := string(ToHTML([]byte(tc.input), nil, nil))
			if tc.want != "" && !bytes.Contains([]byte(got), []byte(tc.want)) {
				t.Errorf("expected output to contain %q\ngot:\n%s", tc.want, got)
			}
			if tc.wantNot != "" && bytes.Contains([]byte(got), []byte(tc.wantNot)) {
				t.Errorf("expected output not to contain raw %q\ngot:\n%s", tc.wantNot, got)
			}
		})
	}
}

// TestCodeBlockClassCoalescing verifies that when a code block has both
// a language annotation and a custom class attribute, they are merged into
// a single class attribute. See https://github.com/gomarkdown/markdown/issues/209.
func TestCodeBlockClassCoalescing(t *testing.T) {
	input := "``` yml\ntext: something\n```\n"

	p := parser.NewWithExtensions(parser.FencedCode)
	doc := p.Parse([]byte(input))

	// Walk the AST and add a custom class to the CodeBlock node.
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if cb, ok := node.(*ast.CodeBlock); ok {
			cb.Attribute = &ast.Attribute{
				Classes: [][]byte{[]byte("my-class")},
			}
		}
		return ast.GoToNext
	})

	renderer := html.NewRenderer(html.RendererOptions{})
	got := string(Render(doc, renderer))

	// Before the fix, this produced two separate class= attributes:
	//   <code class="language-yml" class="my-class">
	// After the fix, they should be coalesced:
	//   <code class="language-yml my-class">
	expected := `<pre><code class="language-yml my-class">text: something
</code></pre>
`
	if got != expected {
		t.Errorf("CodeBlock class coalescing failed.\nExpected:\n%s\nGot:\n%s", expected, got)
	}
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

func TestAbsolutePrefixRelativeImagePath(t *testing.T) {
	tests := []string{
		`![](image.png)`,
		`<p><img src="/public/image.png" alt="" /></p>` + "\n",
	}
	opts := html.RendererOptions{
		AbsolutePrefix: "/public",
	}
	params := TestParams{
		RendererOptions: opts,
		extensions:      parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

func TestBlockquoteAfterListWithInsufficientIndent(t *testing.T) {
	tests := []string{
		"1. list\n  > blockquote\n",
		"<ol>\n<li>list</li>\n</ol>\n\n<blockquote>\n<p>blockquote</p>\n</blockquote>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

func TestHardLineBreakDoesNotApplyInsideHTMLBlock(t *testing.T) {
	tests := []string{
		"<div>\n  <a href=\"#\">\n    <img src=\"logo.png\" alt=\"Logo\">\n  </a>\n</div>\n\nHello\nWorld",
		"<div>\n  <a href=\"#\">\n    <img src=\"logo.png\" alt=\"Logo\">\n  </a>\n</div>\n\n<p>Hello<br>\nWorld</p>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions | parser.HardLineBreak,
	}
	doTestsParam(t, tests, params)
}

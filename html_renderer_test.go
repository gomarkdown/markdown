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

// GHSA-g6w5-3rhc-c379: the same unrestricted values reach the table of
// contents, figure IDs, citation links and kramdown block attributes, and
// must be escaped there too.
func TestAttributeEscapeTOCFigureCitationAttrs(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		exts    parser.Extensions
		flags   html.Flags
		want    string
		wantNot string
	}{
		{
			name:    "toc-heading-id-tag-breakout",
			input:   "# T {#x\"><script>alert(1)</script>}\n\nbody\n",
			exts:    parser.CommonExtensions,
			flags:   html.CommonFlags | html.TOC,
			want:    `<a href="#x&quot;&gt;&lt;script&gt;alert(1)&lt;/script&gt;">T</a>`,
			wantNot: `<script>`,
		},
		{
			name:    "toc-heading-id-attr-injection",
			input:   "# T {#\" onmouseover=\"alert(1)}\n",
			exts:    parser.CommonExtensions,
			flags:   html.CommonFlags | html.TOC,
			want:    `<a href="#&quot; onmouseover=&quot;alert(1)">T</a>`,
			wantNot: `onmouseover="alert(1)"`,
		},
		{
			name:    "toc-skiphtml-safelink",
			input:   "# T {#x\"><script>alert(1)</script>}\n",
			exts:    parser.CommonExtensions,
			flags:   html.SkipHTML | html.Safelink | html.TOC,
			want:    `<a href="#x&quot;&gt;&lt;script&gt;alert(1)&lt;/script&gt;">T</a>`,
			wantNot: `<script>`,
		},
		{
			name:    "figure-fenced-caption-id",
			input:   "```\ncode\n```\nFigure: cap {#x\"><script>alert(1)</script>}\n",
			exts:    parser.CommonExtensions | parser.Mmark,
			flags:   html.CommonFlags,
			want:    `<figure id="x&quot;&gt;&lt;script&gt;alert(1)&lt;/script&gt;">`,
			wantNot: `<script>`,
		},
		{
			name:    "figure-block-id",
			input:   "!---\ntext\n!---\nFigure: cap {#x\" onclick=\"alert(1)}\n",
			exts:    parser.CommonExtensions | parser.Mmark,
			flags:   html.CommonFlags,
			want:    `<figure id="x&quot; onclick=&quot;alert(1)">`,
			wantNot: `onclick="alert(1)"`,
		},
		{
			name:    "citation-destination",
			input:   "see [@a\"><svg onload=alert(1) x=\"]\n",
			exts:    parser.CommonExtensions | parser.Mmark,
			flags:   html.CommonFlags,
			want:    `<a href="#a&quot;&gt;&lt;svg onload=alert(1) x=&quot;">`,
			wantNot: `<svg`,
		},
		{
			name:  "block-attrs-id-class",
			input: "para\n{: #x\"><script>alert(1)</script> .c\"><b>}\n",
			exts:  parser.CommonExtensions | parser.Attributes,
			flags: html.CommonFlags,
			// The unbalanced quote makes the parser swallow the rest of the
			// list into the id; what matters is that it is escaped.
			want:    `<p id="x&quot;&gt;&lt;script&gt;alert(1)&lt;/script&gt; .c&quot;&gt;&lt;b&gt;">para</p>`,
			wantNot: `<script>`,
		},
		{
			name:    "block-attrs-class-breakout",
			input:   "para\n{: .c<b>x .d}\n",
			exts:    parser.CommonExtensions | parser.Attributes,
			flags:   html.CommonFlags,
			want:    `<p class="c&lt;b&gt;x d">para</p>`,
			wantNot: `<b>`,
		},
		{
			name:    "block-attrs-value-breakout",
			input:   "para\n{: title=\"x\\\" onclick=\\\"alert(1)\"}\n",
			exts:    parser.CommonExtensions | parser.Attributes,
			flags:   html.CommonFlags,
			wantNot: `onclick="alert(1)"`,
		},
		{
			name:    "block-attrs-unsafe-key-dropped",
			input:   "para\n{: a/b=\"1\" data-x=\"2\"}\n",
			exts:    parser.CommonExtensions | parser.Attributes,
			flags:   html.CommonFlags,
			want:    `<p data-x="2">para</p>`,
			wantNot: `a/b`,
		},
		{
			name:  "ordinary-ids-unchanged",
			input: "# T {#intro}\n\n```\ncode\n```\nFigure: cap {#fig:1}\n\nsee [@rfc1]\n\npara\n{: #p1 .cls data-x=\"1\"}\n",
			exts:  parser.CommonExtensions | parser.Mmark | parser.Attributes,
			flags: html.CommonFlags | html.TOC,
			want:  `<a href="#intro">T</a>`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.NewWithExtensions(tc.exts)
			r := html.NewRenderer(html.RendererOptions{Flags: tc.flags})
			got := string(ToHTML([]byte(tc.input), p, r))
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

func TestMarkdownInHTMLTableCell(t *testing.T) {
	tests := []string{
		"<table>\n<tr>\n<td>\n\n**bold**\n\n</td>\n</tr>\n</table>\n",
		"<table>\n\n<tr>\n\n<td>\n\n<p><strong>bold</strong></p>\n\n</td>\n\n</tr>\n\n</table>\n",

		"<table>\n<tr>\n<td>\n\n```\ncode\n```\n\n</td>\n</tr>\n</table>\n",
		"<table>\n\n<tr>\n\n<td>\n\n<pre><code>code\n</code></pre>\n\n</td>\n\n</tr>\n\n</table>\n",

		"<table>\n<tr>\n<td>\n\n| a | b |\n| - | - |\n| 1 | 2 |\n\n</td>\n</tr>\n</table>\n",
		"<table>\n\n<tr>\n\n<td>\n\n<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n\n</td>\n\n</tr>\n\n</table>\n",

		"<table>\n<thead>\n<tr>\n<td>Input</td>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>\n\n*em*\n\n</td>\n</tr>\n</tbody>\n</table>\n",
		"<table>\n\n<thead>\n\n<tr>\n\n<td>Input</td>\n\n</tr>\n\n</thead>\n\n<tbody>\n\n<tr>\n\n<td>\n\n<p><em>em</em></p>\n\n</td>\n\n</tr>\n\n</tbody>\n\n</table>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions | parser.MarkdownInHTML,
	}
	doTestsParam(t, tests, params)
}

func TestMarkdownInHTMLOffLeavesTableRaw(t *testing.T) {
	tests := []string{
		"<table>\n<tr>\n<td>\n\n**bold**\n\n</td>\n</tr>\n</table>\n",
		"<table>\n<tr>\n<td>\n\n**bold**\n\n</td>\n</tr>\n</table>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions,
	}
	doTestsParam(t, tests, params)
}

func TestAfterBlockAttributes(t *testing.T) {
	tests := []string{
		"## foo\n{: data-line=\"1\"}\n",
		"<h2 data-line=\"1\">foo</h2>\n",

		"hello\n{: .cls}\n",
		"<p class=\"cls\">hello</p>\n",

		"{#myid}\n## foo\n",
		"<h2 id=\"myid\">foo</h2>\n",

		"## foo\n{#next}\n## bar\n",
		"<h2>foo</h2>\n\n<h2 id=\"next\">bar</h2>\n",

		"{#id}\n## foo\n{: data-line=\"1\"}\n",
		"<h2 id=\"id\" data-line=\"1\">foo</h2>\n",

		"{: data-line=\"1\"}\n## foo\n",
		"<h2 data-line=\"1\">foo</h2>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions | parser.Attributes,
	}
	doTestsParam(t, tests, params)
}

func TestAfterBlockAttributesAutoHeadingID(t *testing.T) {
	tests := []string{
		"## foo\n{: data-line=\"1\"}\n",
		"<h2 id=\"foo\" data-line=\"1\">foo</h2>\n",

		"{#myid}\n## foo\n",
		"<h2 id=\"myid\">foo</h2>\n",

		"{#myid}\n## foo\n{: data-line=\"1\"}\n",
		"<h2 id=\"myid\" data-line=\"1\">foo</h2>\n",

		"{#myid .bar}\n## foo\n",
		"<h2 id=\"myid\" class=\"bar\">foo</h2>\n",

		"{#myid}\n## foo\n\n## foo\n",
		"<h2 id=\"myid\">foo</h2>\n\n<h2 id=\"foo\">foo</h2>\n",

		// An empty # is not an id. Keep the generated one, and do not emit id="".
		"{#}\n## foo\n",
		"<h2 id=\"foo\">foo</h2>\n",

		"{# .c}\n## foo\n",
		"<h2 id=\"foo\" class=\"c\">foo</h2>\n",

		"{#myid #}\n## foo\n",
		"<h2 id=\"foo\">foo</h2>\n",

		// The attribute id occupies the auto-id slot, so a later slug is suffixed.
		"{#foo}\n## Introduction\n\n## Foo\n",
		"<h2 id=\"foo\">Introduction</h2>\n\n<h2 id=\"foo-1\">Foo</h2>\n",

		"## Foo\n\n{#foo}\n## Introduction\n",
		"<h2 id=\"foo\">Foo</h2>\n\n<h2 id=\"foo-1\">Introduction</h2>\n",

		"{#foo}\n## One\n\n{#foo}\n## Two\n",
		"<h2 id=\"foo\">One</h2>\n\n<h2 id=\"foo-1\">Two</h2>\n",

		"{#a}\n## foo {#b}\n",
		"<h2 id=\"a\">foo</h2>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions | parser.Attributes | parser.AutoHeadingIDs,
	}
	doTestsParam(t, tests, params)
}

func TestBlockAttributeHeadingIDMatchesTOC(t *testing.T) {
	tests := []string{
		"{#foo}\n## Introduction\n\n## Foo\n",
		"<nav>\n\n<ul>\n<li>\n<ul>\n<li><a href=\"#foo\">Introduction</a></li>\n\n<li><a href=\"#foo-1\">Foo</a></li>\n</ul></li>\n</ul>\n\n</nav>\n\n<h2 id=\"foo\">Introduction</h2>\n\n<h2 id=\"foo-1\">Foo</h2>\n",

		"## Foo\n\n{#foo}\n## Introduction\n",
		"<nav>\n\n<ul>\n<li>\n<ul>\n<li><a href=\"#foo\">Foo</a></li>\n\n<li><a href=\"#foo-1\">Introduction</a></li>\n</ul></li>\n</ul>\n\n</nav>\n\n<h2 id=\"foo\">Foo</h2>\n\n<h2 id=\"foo-1\">Introduction</h2>\n",
	}
	params := TestParams{
		extensions: parser.CommonExtensions | parser.Attributes | parser.AutoHeadingIDs,
		Flags:      html.TOC,
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

func renderAttrs(t *testing.T, input string, flags html.Flags) string {
	t.Helper()
	exts := parser.CommonExtensions | parser.Attributes | parser.InlineAttributes | parser.OrderedListStart
	r := html.NewRenderer(html.RendererOptions{Flags: flags})
	return string(ToHTML([]byte(input), parser.NewWithExtensions(exts), r))
}

// With SkipHTML or Safelink, attributes from the Attributes and
// InlineAttributes extensions could still run script or replace a destination
// that Safelink had checked.
func TestDocumentAttributesUntrusted(t *testing.T) {
	for _, flags := range []html.Flags{html.SkipHTML, html.Safelink} {
		for input, want := range map[string]string{
			"{onmouseover=\"alert(1)\"}\npara\n":        "<p>para</p>\n",
			"[x](/u){onclick=\"alert(1)\"}\n":           "<p><a href=\"/u\">x</a></p>\n",
			"{style=\"color:red\"}\n# h\n":              "<h1>h</h1>\n",
			"![x](/u.png){src=\"javascript:alert(1)\"}": "<p><img src=\"/u.png\" alt=\"x\" /></p>\n",
			"[x](/u){href=\"javascript:alert(1)\"}\n":   "<p><a href=\"/u\">x</a></p>\n",
			// attributes that can't run script or load anything stay
			"{#i .c lang=\"en\" data-x=\"1\" aria-label=\"a\" width=\"5\"}\npara\n": "<p id=\"i\" class=\"c\" aria-label=\"a\" data-x=\"1\" lang=\"en\" width=\"5\">para</p>\n",
			"[x](/u){title=\"t\" data-x=\"1\"}\n":                                   "<p><a href=\"/u\" data-x=\"1\" title=\"t\">x</a></p>\n",
		} {
			if got := renderAttrs(t, input, flags); got != want {
				t.Errorf("flags %d, %q:\ngot  %q\nwant %q", flags, input, got, want)
			}
		}
	}
}

// Without those flags a document can set any attribute, but not one that
// the renderer writes on the element.
func TestDocumentAttributesTrusted(t *testing.T) {
	for input, want := range map[string]string{
		"[x](http://a.b){onclick=\"f()\" target=\"_blank\"}\n": "<p><a href=\"http://a.b\" onclick=\"f()\" target=\"_blank\">x</a></p>\n",
		"[x](/u){href=\"/other\" data-x=\"1\"}\n":              "<p><a href=\"/u\" data-x=\"1\">x</a></p>\n",
		"![x](/u.png){src=\"/other.png\" width=\"10\"}\n":      "<p><img width=\"10\" src=\"/u.png\" alt=\"x\" /></p>\n",
		"{start=\"9\"}\n3. a\n":                                "<ol start=\"3\">\n<li>a</li>\n</ol>\n",
		"{style=\"color:red\"}\n# h\n":                         "<h1 style=\"color:red\">h</h1>\n",
	} {
		if got := renderAttrs(t, input, 0); got != want {
			t.Errorf("%q:\ngot  %q\nwant %q", input, got, want)
		}
	}
	// the renderer's own target takes the place of the document's
	got := renderAttrs(t, "[x](http://a.b){target=\"_self\"}\n", html.HrefTargetBlank)
	if want := "<p><a href=\"http://a.b\" target=\"_blank\">x</a></p>\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// An escaped delimiter ended a super- or subscript anyway.
func TestSuperSubscriptEscapedDelimiter(t *testing.T) {
	for input, want := range map[string]string{
		`H~2\~x~O`: "<p>H<sub>2~x</sub>O</p>\n",
		`a^b\^c^d`: "<p>a<sup>b^c</sup>d</p>\n",
		`~a\~`:     "<p>~a~</p>\n",
		`x^\^^`:    "<p>x<sup>^</sup></p>\n",
		`a^b\\^c`:  "<p>a<sup>b\\</sup>c</p>\n",
		`H~2~O`:    "<p>H<sub>2</sub>O</p>\n",
	} {
		p := parser.NewWithExtensions(parser.CommonExtensions | parser.SuperSubscript)
		if got := string(ToHTML([]byte(input+"\n"), p, nil)); got != want {
			t.Errorf("%q: got %q, want %q", input, got, want)
		}
	}
}

// Super- and subscripts were written without escaping HTML, so they could
// add markup even with SkipHTML.
func TestSuperSubscriptEscapeHTML(t *testing.T) {
	exts := parser.CommonExtensions | parser.SuperSubscript
	r := html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML})
	got := string(ToHTML([]byte("a^<img/src/onerror=alert(1)>^ b~<b>~ c^\\*^\n"), parser.NewWithExtensions(exts), r))
	want := "<p>a<sup>&lt;img/src/onerror=alert(1)&gt;</sup> b<sub>&lt;b&gt;</sub> c<sup>*</sup></p>\n"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

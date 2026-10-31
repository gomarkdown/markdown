package md

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func TestRenderDocument(t *testing.T) {
	source := []byte("# title\n* aaa\n* bbb\n* ccc")
	input := markdown.Parse(source, nil)
	expected := "# title\n\n* aaa\n* bbb\n* ccc\n\n"
	testRendering(t, input, expected)
}

func TestRenderListFollowedByHeadingRoundTrip(t *testing.T) {
	source := []byte("# Header1\n* list item\n\n# Header2\n")
	p := parser.NewWithExtensions(parser.CommonExtensions)
	input := p.Parse(source)

	expected := "# Header1\n\n* list item\n\n# Header2\n\n"
	testRendering(t, input, expected)

	output := markdown.Render(input, NewRenderer())
	p = parser.NewWithExtensions(parser.CommonExtensions)
	roundTripped := p.Parse(output)
	if len(input.GetChildren()) != len(roundTripped.GetChildren()) {
		t.Fatalf("round-trip changed top-level node count: got %d, want %d\n%s",
			len(roundTripped.GetChildren()), len(input.GetChildren()), output)
	}
}

func TestRenderText(t *testing.T) {
	var input ast.Node = &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello"))}}
	expected := "Hello"
	testRendering(t, input, expected)
}

// testSameHTML renders source to Markdown, reparses it and checks that it still
// means the same by comparing the HTML of both.
func testSameHTML(t *testing.T, newParser func() *parser.Parser, source string, opts ...RendererOpt) string {
	t.Helper()
	output := markdown.Render(newParser().Parse([]byte(source)), NewRenderer(opts...))
	want := markdown.ToHTML([]byte(source), newParser(), nil)
	got := markdown.ToHTML(output, newParser(), nil)
	if string(got) != string(want) {
		t.Errorf("rendered %q as %q\nHTML: %q\nwant: %q", source, output, got, want)
	}
	return string(output)
}

var (
	testLegacyParser = func() *parser.Parser { return parser.NewWithExtensions(parser.CommonExtensions) }
	testParsers      = map[string]func() *parser.Parser{"legacy": testLegacyParser, "commonmark": parser.NewCommonMark}
)

func TestRenderTextEscaping(t *testing.T) {
	for _, source := range []string{
		"\\*x\\*\n", "\\_x\\_\n", "\\`x\\`\n", "\\[a\\](b)\n", "a\\\\b\n",
		"\\# x\n", "\\- x\n", "\\+ x\n", "\\> x\n", "1\\. foo\n", "1\\) foo\n", "\\---\n",
		"&lt;script&gt;alert(1)&lt;/script&gt;\n", "\\<b>\n", "_a*b_\n", "*a\\*b*\n",
		"a\\*b *c*\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				testSameHTML(t, p, source)
			})
		}
	}

	for _, source := range []string{
		"\\![a](b)\n", "\\~~x\\~~\n", "\\&amp;\n", "\\&amp; \\&#35;\n", "&amp;amp;\n", "\\$x\\$\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				testSameHTML(t, p, source)
			})
		}
	}

	// nothing to escape
	for _, source := range []string{
		"snake_case and a-b and 1.5 and a < b and C.\n",
		"-5 and +5 and 1.x\n",
		"AT&T and a & b and wow! [a](b)\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				if got := testSameHTML(t, p, source); got != source+"\n" {
					t.Errorf("got %q, want %q", got, source+"\n")
				}
			})
		}
	}

	// the legacy parser keeps image descriptions as they are in the source
	source := "![a*b](c)\n"
	if got := testSameHTML(t, testLegacyParser, source); got != source+"\n" {
		t.Errorf("got %q, want %q", got, source+"\n")
	}
	testSameHTML(t, parser.NewCommonMark, source)
}

func TestRenderStrong(t *testing.T) {
	var input ast.Node = &ast.Strong{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello"))}})
	expected := "**Hello**"
	testRendering(t, input, expected)
}

func TestRenderHeading(t *testing.T) {
	var input ast.Node = &ast.Heading{Level: 3}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello"))}})
	expected := "### Hello\n\n"
	testRendering(t, input, expected)
}

func TestRenderHeadingID(t *testing.T) {
	source := []byte("# T {#myid}\n")
	input := parser.NewWithExtensions(parser.CommonExtensions).Parse(source)
	expected := "# T {#myid}\n\n"
	testRendering(t, input, expected)
}

func TestRenderEmph(t *testing.T) {
	var input ast.Node = &ast.Emph{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello"))}})
	expected := "*Hello*"
	testRendering(t, input, expected)
}

func TestRenderDel(t *testing.T) {
	var input ast.Node = &ast.Del{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello"))}})
	expected := "~~Hello~~"
	testRendering(t, input, expected)
}

func TestRenderLink(t *testing.T) {
	var input ast.Node = &ast.Link{Title: []byte(string("Hello")), Destination: []byte(string("hello.io"))}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello World !"))}})
	expected := "[Hello World !](hello.io \"Hello\")"
	testRendering(t, input, expected)
}

func TestRenderImage(t *testing.T) {
	var input ast.Node = &ast.Image{Title: []byte(string("Hello")), Destination: []byte(string("hello.io"))}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello World !"))}})
	expected := "![Hello World !](hello.io \"Hello\")"
	testRendering(t, input, expected)
}

func TestRenderTitleQuotes(t *testing.T) {
	titleOf := func(n ast.Node) []byte {
		var title []byte
		ast.WalkFunc(n, func(node ast.Node, entering bool) ast.WalkStatus {
			switch node := node.(type) {
			case *ast.Link:
				title = node.Title
			case *ast.Image:
				title = node.Title
			}
			return ast.GoToNext
		})
		return title
	}
	for _, source := range []string{
		"[a](b 'c\"d')\n",
		"![a](b 'c\"d')\n",
		"[a][r]\n\n[r]: /u 'c\"d'\n",
	} {
		for name, p := range map[string]func() *parser.Parser{"legacy": parser.New, "commonmark": parser.NewCommonMark} {
			output := markdown.Render(p().Parse([]byte(source)), NewRenderer())
			if got := titleOf(p().Parse(output)); string(got) != `c"d` {
				t.Errorf("%s: rendered %q as %q; got title %q", name, source, output, got)
			}
		}
	}

	// both kinds of quotes can only be written with a backslash
	input := &ast.Link{Title: []byte(`c"d'e`), Destination: []byte("b")}
	output := markdown.Render(input, NewRenderer())
	if got := titleOf(parser.NewCommonMark().Parse(append([]byte("["+"a"), output[1:]...))); string(got) != `c"d'e` {
		t.Errorf("rendered as %q; got title %q", output, got)
	}
}

func TestRenderCode(t *testing.T) {
	input := &ast.Code{}
	input.Literal = []byte(string("val x : Int = 42"))
	expected := "`val x : Int = 42`"
	testRendering(t, input, expected)
}

// testCodeRoundTrip renders a code node with the given literal, reparses the
// output and checks that it is exactly one code span holding want.
func testCodeRoundTrip(t *testing.T, p *parser.Parser, literal, want string) {
	t.Helper()
	input := &ast.Code{Leaf: ast.Leaf{Literal: []byte(literal)}}
	output := markdown.Render(input, NewRenderer())
	var codes []string
	ast.WalkFunc(markdown.Parse(output, p), func(node ast.Node, entering bool) ast.WalkStatus {
		if code, ok := node.(*ast.Code); ok && entering {
			codes = append(codes, string(code.Literal))
		}
		return ast.GoToNext
	})
	if len(codes) != 1 || codes[0] != want {
		t.Fatalf("rendered %q as %q; parsed code literals: %q, want %q", literal, output, codes, want)
	}
}

func TestRenderCodeRoundTrip(t *testing.T) {
	for _, literal := range []string{
		"plain code", "a`b", "a``b", "a`b``c", "`code", "code`", "`", "``",
		" code ", "  code  ", "   ", " a", "a ", "` ", " `",
	} {
		t.Run(literal, func(t *testing.T) {
			testCodeRoundTrip(t, parser.NewCommonMark(), literal, literal)
		})
	}
}

// CommonMark turns line endings in a code span into spaces, so that is the
// most that can survive. A literal ending in a line ending also loses that
// trailing space: padding after it is stripped as the next line's indentation.
func TestRenderCodeRoundTripNewlines(t *testing.T) {
	for _, tt := range []struct{ literal, want string }{
		{"a\nb", "a b"},
		{"a``\nb", "a`` b"},
		{"\na ", " a "},
		{" a\n", " a"},
		{"\na\n", " a"},
	} {
		t.Run(tt.literal, func(t *testing.T) {
			testCodeRoundTrip(t, parser.NewCommonMark(), tt.literal, tt.want)
		})
	}
}

// The legacy parser trims spaces at the edges of a code span, so those can't
// round trip. It also treats a multi-line span with a 3 backtick delimiter as
// a code block.
func TestRenderCodeRoundTripLegacy(t *testing.T) {
	for _, literal := range []string{
		"plain code", "a`b", "a``b", "a`b``c", "`code", "code`", "`", "``",
		"a\nb", "a``\nb", "a`b``\nc", "\na\n",
	} {
		t.Run(literal, func(t *testing.T) {
			testCodeRoundTrip(t, parser.New(), literal, literal)
		})
	}
}

func TestRenderCodeEmpty(t *testing.T) {
	testRendering(t, &ast.Code{}, "")
}

func TestRenderCodeBlock(t *testing.T) {
	input := &ast.CodeBlock{Info: []byte(string("scala"))}
	input.Literal = []byte(string("val x : Int = 42"))
	expected := "\n```scala\nval x : Int = 42\n```\n\n"
	testRendering(t, input, expected)
}

func TestRenderCodeBlockFence(t *testing.T) {
	for _, source := range []string{
		"````\n```\nx\n```\n````\n",
		"    a\n    ```\n    b\n",
	} {
		for name, p := range map[string]func() *parser.Parser{"legacy": parser.New, "commonmark": parser.NewCommonMark} {
			doc := p().Parse([]byte(source))
			want := ast.GetFirstChild(doc).(*ast.CodeBlock).Literal
			output := markdown.Render(doc, NewRenderer())
			children := p().Parse(output).GetChildren()
			if len(children) != 1 {
				t.Fatalf("%s: rendered %q as %q; got %d blocks", name, source, output, len(children))
			}
			if got := children[0].(*ast.CodeBlock).Literal; string(got) != string(want) {
				t.Errorf("%s: rendered %q as %q; got literal %q, want %q", name, source, output, got, want)
			}
		}
	}
}

func TestRenderCodeBlockInfo(t *testing.T) {
	for _, tt := range []struct{ info, expected string }{
		{"go", "\n```go\nz\n```\n\n"},
		{".go", "\n```go\nz\n```\n\n"},
		{" go title=x ", "\n```go title=x\nz\n```\n\n"},
		{"a`b", "\n~~~a`b\nz\n~~~\n\n"},
	} {
		input := &ast.CodeBlock{Info: []byte(tt.info), Leaf: ast.Leaf{Literal: []byte("z\n")}}
		testRendering(t, input, tt.expected)
	}

	input := parser.NewCommonMark().Parse([]byte("~~~ a`b\nz\n~~~\n"))
	output := markdown.Render(input, NewRenderer())
	reparsed := parser.NewCommonMark().Parse(output)
	code, ok := ast.GetFirstChild(reparsed).(*ast.CodeBlock)
	if !ok || string(code.Info) != "a`b" {
		t.Errorf("rendered as %q, which doesn't keep the info string", output)
	}
}

func TestRenderParagraph(t *testing.T) {
	input := &ast.Paragraph{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte(string("Hello World !"))}})
	expected := "Hello World !\n\n"
	testRendering(t, input, expected)
}

func TestRenderDoubleParagraph(t *testing.T) {
	input := markdown.Parse([]byte("Paragraph 1\n\nParagraph 2"), nil)

	expected := "Paragraph 1\n\nParagraph 2\n\n"

	testRendering(t, input, expected)
}

func TestRenderCodeWithParagraph(t *testing.T) {
	input := markdown.Parse([]byte("```\nnpm run build\nnpm run test:e2e:dev\n```"), nil)

	expected := "\n```\nnpm run build\nnpm run test:e2e:dev\n```\n\n"

	testRendering(t, input, expected)
}

func TestRenderHTMLSpan(t *testing.T) {
	input := &ast.HTMLSpan{}
	input.Literal = []byte(string("hello"))
	expected := "hello"
	testRendering(t, input, expected)
}

func TestRenderHTMLBlock(t *testing.T) {
	input := &ast.HTMLBlock{}
	input.Literal = []byte(string("hello"))
	expected := "\nhello\n\n"
	testRendering(t, input, expected)
}

func TestRenderList(t *testing.T) {
	source := []byte("* aaa\n* bbb\n* ccc\n* ddd\n")
	input := markdown.Parse(source, nil)
	expected := "* aaa\n* bbb\n* ccc\n* ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("+ aaa\n+ bbb\n+ ccc\n+ ddd\n")
	input = markdown.Parse(source, nil)
	expected = "+ aaa\n+ bbb\n+ ccc\n+ ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("- aaa\n- bbb\n- ccc\n- ddd\n")
	input = markdown.Parse(source, nil)
	expected = "- aaa\n- bbb\n- ccc\n- ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("1. aaa\n2. bbb\n3. ccc\n4. ddd\n")
	input = markdown.Parse(source, nil)
	expected = "1. aaa\n2. bbb\n3. ccc\n4. ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("1. aaa\n1. bbb\n1. ccc\n1. ddd\n")
	input = markdown.Parse(source, nil)
	expected = "1. aaa\n2. bbb\n3. ccc\n4. ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("1. aaa\n3. bbb\n8. ccc\n1. ddd\n")
	input = markdown.Parse(source, nil)
	expected = "1. aaa\n2. bbb\n3. ccc\n4. ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("* aaa\n    * aaa1\n    * aaa2\n* bbb\n* ccc\n* ddd\n")
	input = markdown.Parse(source, nil)
	expected = "* aaa\n    * aaa1\n    * aaa2\n* bbb\n* ccc\n* ddd\n\n"
	testRendering(t, input, expected)

	source = []byte("* aaa\n    * aaa1\n\n")
	input = markdown.Parse(source, nil)
	expected = "* aaa\n    * aaa1\n\n"
	testRendering(t, input, expected)

	source = []byte("* aaa\n    * aaa1\n        * aaa1.1\n\n")
	input = markdown.Parse(source, nil)
	expected = "* aaa\n    * aaa1\n        * aaa1.1\n\n"
	testRendering(t, input, expected)

	source = []byte("1. aaa\n    1. aaa1\n\n")
	input = markdown.Parse(source, nil)
	expected = "1. aaa\n    1. aaa1\n\n"
	testRendering(t, input, expected)

	source = []byte("1. aaa\n    1. aaa1\n        1. aaa1.1\n\n")
	input = markdown.Parse(source, nil)
	expected = "1. aaa\n    1. aaa1\n        1. aaa1.1\n\n"
	testRendering(t, input, expected)

	source = []byte("This is an [example](https://example.com) and another [website](https://github.com).")
	input = markdown.Parse(source, nil)
	rendererOpts := []RendererOpt{WithRenderInFooter(true)}
	expected = "This is an [example] and another [website].\n\n\n[example]: https://example.com\n[website]: https://github.com\n"
	testRendering(t, input, expected, rendererOpts...)
}

func TestRenderListItemBlocks(t *testing.T) {
	for _, source := range []string{
		"* a\n\n    para2\n\n* c\n",
		"* a\n\n    ```\n    x\n\n    y\n    ```\n* c\n",
		"* a\n    * n\n\n    tail\n",
		"* a\n\n* b\n",
		"1. a\n\n    b\n\n2. c\n",
		"* a\n\n    * n\n\n        deep\n\n    tail\n\n* b\n",
		"* a\n\n    <div>x</div>\n\n* b\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				testSameHTML(t, p, source)
			})
		}
	}
	testSameHTML(t, parser.NewCommonMark, "100. a\n\n     b\n")
}

func TestRenderDefinitionList(t *testing.T) {
	for _, source := range []string{
		"Term\n: def\n",
		"T1\n: d1\n\nT2\n: d2\n",
		"T\n: d1\n: d2\n",
		"T\n\n: d\n\n    para\n",
		"*T*\n: *d*\n\nafter\n",
	} {
		t.Run(source, func(t *testing.T) {
			testSameHTML(t, testLegacyParser, source)
		})
	}
	testRendering(t, testLegacyParser().Parse([]byte("Term\n: def\n")), "Term\n: def\n\n")
}

func TestRenderFootnotes(t *testing.T) {
	footnotes := func() *parser.Parser {
		return parser.NewWithExtensions(parser.CommonExtensions | parser.Footnotes)
	}
	for _, source := range []string{
		"a[^1]\n\n[^1]: note\n",
		"a[^x] b[^y] c[^x]\n\n[^x]: *one*\n\n[^y]: two\n",
		"a[^long-id]\n\n[^long-id]: para\n\n    more\n\n    * list\n\nafter\n",
		"a^[inline note]\n",
		"* item[^1]\n\n[^1]: note\n",
	} {
		t.Run(source, func(t *testing.T) {
			testSameHTML(t, footnotes, source)
		})
	}
	input := footnotes().Parse([]byte("a[^1]\n\n[^1]: note\n"))
	testRendering(t, input, "a[^1]\n\n[^1]: note\n\n")
}

func TestRenderBlockQuote(t *testing.T) {
	for _, source := range []string{
		"> q\n",
		"> a\n>\n> b\n",
		"> a\n>\n> > nested\n>\n> ```\n> x\n>\n> y\n> ```\n",
		"> * a\n> * b\n\nafter\n",
		"* a\n\n    > q\n\n* b\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				testSameHTML(t, p, source)
			})
		}
	}
	input := parser.NewCommonMark().Parse([]byte("> a\n>\n> b\n"))
	testRendering(t, input, "> a\n>\n> b\n\n")
}

func TestRenderInlineAndLeafNodes(t *testing.T) {
	for _, source := range []string{
		"a\n\n---\n\nb\n",
		"foo  \nbar\n",
		"foo  \n\\# bar\n",
		"* a  \n  b\n",
	} {
		for name, p := range testParsers {
			t.Run(name+" "+source, func(t *testing.T) {
				testSameHTML(t, p, source)
			})
		}
	}
	for _, source := range []string{
		"a $x+y$ b\n",
		"$$\nx\n$$\n",
		"foo\\\nbar\n",
	} {
		t.Run(source, func(t *testing.T) {
			testSameHTML(t, testLegacyParser, source)
		})
	}
	withExtensions := func(e parser.Extensions) func() *parser.Parser {
		return func() *parser.Parser { return parser.NewWithExtensions(e) }
	}
	testSameHTML(t, withExtensions(parser.SuperSubscript), "H~2~O and x^2^\n")
	testRendering(t, withExtensions(parser.NonBlockingSpace)().Parse([]byte("a\\ b\n")), "a&nbsp;b\n\n")
	testRendering(t, &ast.Softbreak{}, "\n")
}

func TestRenderTable(t *testing.T) {
	for _, source := range []string{
		"a|b\n-|-\n1|2\n",
		"| a | b | c | d |\n|:--|--:|:-:|---|\n| 1 | 2 | 3 | 4 |\n",
		"| a | b |\n|---|---|\n| x \\| y | `p|q` |\n",
		"| *a* | [l](u) |\n|---|---|\n| | z |\n\nafter\n",
		"| a | b | c |\n|---|---|---|\n| 1 || 3 |\n",
		"| a | b |\n|---|---|\n| 1 | 2 |\n===|===\n| f | g |\n",
		"|---|---|\n| 1 | 2 |\n",
		"* x\n\n    | a | b |\n    |---|---|\n    | 1 | 2 |\n\n* y\n",
	} {
		t.Run(source, func(t *testing.T) {
			testSameHTML(t, testLegacyParser, source)
		})
	}
	input := testLegacyParser().Parse([]byte("a|b\n:-|-:\n1|2\n"))
	testRendering(t, input, "| a | b |\n| :--- | ---: |\n| 1 | 2 |\n\n")
}

func TestRenderMmarkNodes(t *testing.T) {
	mmark := func() *parser.Parser {
		return parser.NewWithExtensions(parser.CommonExtensions | parser.Mmark)
	}
	for _, source := range []string{
		"A> aside\nA>\nA> more\n",
		"{frontmatter}\n\na\n\n{mainmatter}\n\nb\n\n{backmatter}\n\nc\n",
		"```\ncode\n```\nFigure: cap *x* {#id}\n\nafter\n",
		"> q\n\nQuote: cap\n\nafter\n",
		"| a | b |\n|---|---|\n| 1 | 2 |\nTable: cap\n\nafter\n",
		"!---\n![a](b)\n\n![c](d)\n!---\nFigure: cap\n\nafter\n",
		"see (#ref) and (#ref, sec 1)\n",
		"[@a] and [@!a, p. 3; @-b]\n",
		"(!item) (!!item, sub)\n",
		"```\ncode <<1>>\n```\n",
	} {
		t.Run(source, func(t *testing.T) {
			testSameHTML(t, mmark, source)
		})
	}
}

// customLeaf is the kind of node a parser hook can add.
type customLeaf struct{ ast.Leaf }

func TestRenderUnknownNodes(t *testing.T) {
	testRendering(t, &customLeaf{ast.Leaf{Literal: []byte("x")}}, "x")

	type customContainer struct{ ast.Container }
	input := &customContainer{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte("y")}})
	testRendering(t, input, "y")
}

func TestRenderReferenceDefinitionRoundTrip(t *testing.T) {
	source := []byte("Hello [world][wiki-world]\n\n[wiki-world]: https://wikipedia.com/blah-blah-blah \"Wiki\"\n")
	p := parser.NewWithExtensions(parser.CommonExtensions)
	input := p.Parse(source)
	expected := "Hello [world][wiki-world]\n\n[wiki-world]: https://wikipedia.com/blah-blah-blah \"Wiki\"\n"
	testRendering(t, input, expected)
}

func TestRenderShortcutReferenceRoundTrip(t *testing.T) {
	source := []byte("[ref]\n\n[ref]: /url/\n")
	p := parser.NewWithExtensions(parser.CommonExtensions)
	input := p.Parse(source)
	expected := "[ref]\n\n[ref]: /url/\n"
	testRendering(t, input, expected)
}

func TestRenderOrderedListMarker(t *testing.T) {
	input := parser.NewCommonMark().Parse([]byte("3. a\n4. b\n"))
	testRendering(t, input, "3. a\n4. b\n\n")

	input = parser.NewCommonMark().Parse([]byte("1) a\n2) b\n"))
	testRendering(t, input, "1) a\n2) b\n\n")

	input = parser.NewWithExtensions(parser.CommonExtensions | parser.OrderedListStart).Parse([]byte("3. a\n4. b\n"))
	testRendering(t, input, "3. a\n4. b\n\n")
}

func TestRenderParagraphSeparator(t *testing.T) {
	// the last child writing nothing doesn't make the paragraph empty
	input := &ast.Paragraph{}
	ast.AppendChild(input, &ast.Text{Leaf: ast.Leaf{Literal: []byte("a")}})
	ast.AppendChild(input, &ast.HTMLSpan{})
	testRendering(t, input, "a\n\n")

	// output before an empty paragraph doesn't make it non-empty
	doc := &ast.Document{}
	heading := &ast.Heading{Level: 1}
	ast.AppendChild(heading, &ast.Text{Leaf: ast.Leaf{Literal: []byte("t")}})
	ast.AppendChild(doc, heading)
	ast.AppendChild(doc, &ast.Paragraph{})
	testRendering(t, doc, "# t\n\n")
}

func TestRenderZeroValueRenderer(t *testing.T) {
	input := markdown.Parse([]byte("1. a\n2. b\n"), nil)
	got := string(markdown.Render(input, &Renderer{}))
	if expected := "1. a\n2. b\n\n"; got != expected {
		t.Errorf("[%s] is not equal to [%s]", got, expected)
	}
}

func TestRenderReusedRendererFooter(t *testing.T) {
	r := NewRenderer(WithRenderInFooter(true))
	markdown.Render(markdown.Parse([]byte("[a](x)\n"), nil), r)
	got := string(markdown.Render(markdown.Parse([]byte("[b](y)\n"), nil), r))
	if expected := "[b]\n\n\n[b]: y\n"; got != expected {
		t.Errorf("[%s] is not equal to [%s]", got, expected)
	}
}

func testRendering(t *testing.T, input ast.Node, expected string, opts ...RendererOpt) {
	renderer := NewRenderer(opts...)
	result := string(markdown.Render(input, renderer))
	if strings.Compare(result, expected) != 0 {
		t.Errorf("[%s] is not equal to [%s]", result, expected)
	}
}

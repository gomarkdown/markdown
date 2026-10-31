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

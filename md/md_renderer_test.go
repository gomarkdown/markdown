package md

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

func TestRenderDocument(t *testing.T) {
	source := []byte("# title\n* aaa\n* bbb\n* ccc")
	input := markdown.Parse(source, nil)
	expected := "# title\n\n* aaa\n* bbb\n* ccc\n\n"
	testRendering(t, input, expected)
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

func TestRenderCodeBlock(t *testing.T) {
	input := &ast.CodeBlock{Info: []byte(string("scala"))}
	input.Literal = []byte(string("val x : Int = 42"))
	expected := "\n```scala\nval x : Int = 42\n```\n"
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
	expected = "* aaa\n    * aaa1\n    * aaa2\n\n* bbb\n* ccc\n* ddd\n\n"
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

func testRendering(t *testing.T, input ast.Node, expected string, opts ...RendererOpt) {
	renderer := NewRenderer(opts...)
	result := string(markdown.Render(input, renderer))
	if strings.Compare(result, expected) != 0 {
		t.Errorf("[%s] is not equal to [%s]", result, expected)
	}
}

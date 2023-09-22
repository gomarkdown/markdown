package markdown

import (
	"bytes"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"testing"
)

func TestPrefixHeaderNoExtensions(t *testing.T) {
	doTestsBlock(t, "PrefixHeaderNoExtensions.tests", 0)
}

func TestPrefixHeaderSpaceExtension(t *testing.T) {
	doTestsBlock(t, "PrefixHeaderSpaceExtension.tests", parser.SpaceHeadings)
}

func TestPrefixHeaderIdExtension(t *testing.T) {
	doTestsBlock(t, "PrefixHeaderIdExtension.tests", parser.HeadingIDs)
}

func TestPrefixHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderIdExtensionWithPrefixAndSuffix.tests")

	parameters := html.RendererOptions{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	doTestsParam(t, tests, TestParams{
		extensions:      parser.HeadingIDs,
		Flags:           html.UseXHTML,
		RendererOptions: parameters,
	})
}

func TestPrefixAutoHeaderIdExtension(t *testing.T) {
	doTestsBlock(t, "PrefixAutoHeaderIdExtension.tests", parser.AutoHeadingIDs)
}

func TestPrefixAutoHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	tests := readTestFile2(t, "PrefixAutoHeaderIdExtensionWithPrefixAndSuffix.tests")
	parameters := html.RendererOptions{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	doTestsParam(t, tests, TestParams{
		extensions:      parser.AutoHeadingIDs,
		Flags:           html.UseXHTML,
		RendererOptions: parameters,
	})
}

func TestPrefixMultipleHeaderExtensions(t *testing.T) {
	doTestsBlock(t, "PrefixMultipleHeaderExtensions.tests", parser.AutoHeadingIDs|parser.HeadingIDs)
}

func TestPrefixHeaderMmarkExtension(t *testing.T) {
	doTestsBlock(t, "PrefixHeaderMmarkExtension.tests", parser.Mmark)
}

func TestUnderlineHeaders(t *testing.T) {
	doTestsBlock(t, "UnderlineHeaders.tests", 0)
}

func TestUnderlineHeadersAutoIDs(t *testing.T) {
	doTestsBlock(t, "UnderlineHeadersAutoIDs.tests", parser.AutoHeadingIDs)
}

func TestHorizontalRule(t *testing.T) {
	doTestsBlock(t, "HorizontalRule.tests", 0)
}

func TestUnorderedList(t *testing.T) {
	doTestsBlock(t, "UnorderedList.tests", 0)
}

func TestOrderedList(t *testing.T) {
	doTestsBlock(t, "OrderedList.tests", 0)
}

func TestDefinitionList(t *testing.T) {
	doTestsBlock(t, "DefinitionList.tests", parser.DefinitionLists)
}

func TestNestedDefinitionList(t *testing.T) {
	doTestsBlock(t, "NestedDefinitionList.tests", parser.DefinitionLists)
}

func TestPreformattedHtml(t *testing.T) {
	doTestsBlock(t, "PreformattedHtml.tests", 0)
}

func TestPreformattedHtmlLax(t *testing.T) {
	doTestsBlock(t, "PreformattedHtmlLax.tests", parser.LaxHTMLBlocks)
}

func TestFencedCodeBlock(t *testing.T) {
	doTestsBlock(t, "FencedCodeBlock.tests", parser.FencedCode)
}

func TestFencedCodeInsideBlockquotes(t *testing.T) {
	doTestsBlock(t, "FencedCodeInsideBlockquotes.tests", parser.FencedCode)
}

func TestTable(t *testing.T) {
	doTestsBlock(t, "Table.tests", parser.Tables)
}

func TestUnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	doTestsBlock(t, "UnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests", parser.NoEmptyLineBeforeBlock)
}

func TestOrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	doTestsBlock(t, "OrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests", parser.NoEmptyLineBeforeBlock)
}

func TestFencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	doTestsBlock(t, "FencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests", parser.FencedCode|parser.NoEmptyLineBeforeBlock)
}

func TestMathBlock(t *testing.T) {
	doTestsBlock(t, "MathBlock.tests", parser.CommonExtensions)
}

func TestDefinitionListWithFencedCodeBlock(t *testing.T) {
	doTestsBlock(t, "DefinitionListWithFencedCodeBlock.tests", parser.FencedCode|parser.DefinitionLists)
}

func TestListWithFencedCodeBlockAndHeader(t *testing.T) {
	doTestsBlock(t, "ListWithFencedCodeBlockAndHeader.tests", parser.FencedCode)
}

func TestTitleBlock_EXTENSION_TITLEBLOCK(t *testing.T) {
	doTestsBlock(t, "TitleBlock_EXTENSION_TITLEBLOCK.tests", parser.Titleblock)
}

func TestBlockComments(t *testing.T) {
	doTestsBlock(t, "BlockComments.tests", 0)
}

func TestBug168(t *testing.T) {
	doTestsBlock(t, "bug168.tests", 0)
}

func TestBug196(t *testing.T) {
	extensions := parser.AutoHeadingIDs
	doTestsBlock(t, "bug196.tests", extensions)
}

func TestBug242(t *testing.T) {
	extensions := parser.AutoHeadingIDs
	doTestsBlock(t, "bug242.tests", extensions)
}

func TestTOC(t *testing.T) {
	tests := readTestFile2(t, "TOC.tests")
	doTestsParam(t, tests, TestParams{
		extensions: parser.HeadingIDs,
		Flags:      html.UseXHTML | html.TOC,
	})
}

func TestCompletePage(t *testing.T) {
	tests := readTestFile2(t, "CompletePage.tests")
	doTestsParam(t, tests, TestParams{Flags: html.UseXHTML | html.CompletePage})
}

func TestSpaceHeadings(t *testing.T) {
	tests := readTestFile2(t, "SpaceHeadings.tests")
	doTestsParam(t, tests, TestParams{extensions: parser.SpaceHeadings})
}

func TestCodeInList(t *testing.T) {
	tests := readTestFile2(t, "code_in_list.test")
	exts := parser.CommonExtensions
	doTestsParam(t, tests, TestParams{extensions: exts})
}

func TestLists(t *testing.T) {
	tests := readTestFile2(t, "Lists.tests")
	exts := parser.CommonExtensions
	doTestsParam(t, tests, TestParams{extensions: exts})
}

func TestBug126(t *testing.T) {
	// there's a space after end of table header, which used to break table parsing
	input := "> ```\n> fenced pre block 1\n> ```\n\n```\nfenced pre block 2\n````\n"
	p := parser.NewWithExtensions(parser.CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	// TODO: needs fixing https://github.com/gomarkdown/markdown/issues/126
	exp := "BlockQuote\n  CodeBlock: '> fenced pre block 1\\n> ```\\n\\n'\n  Paragraph\n    Text 'fenced pre block 2\\n````'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

func TestPull288(t *testing.T) {
	input := "```go\nfmt.Println(\"Hello world!\")\n```\n"
	p := parser.NewWithExtensions(parser.CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "CodeBlock:go 'fmt.Println(\"Hello world!\")\\n'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

func TestSec1(t *testing.T) {
	ext := parser.CommonExtensions |
		parser.Attributes |
		parser.OrderedListStart |
		parser.SuperSubscript |
		parser.Mmark

	tests := []string{
		"[@]", "[#]", "[@", "[#", "[@@]", "[@#]",
	}
	for _, test := range tests {
		p := parser.NewWithExtensions(ext)
		inp := []byte(test)
		ToHTML(inp, p, nil)
	}
}

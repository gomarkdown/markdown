package markdown

import (
	"testing"

	"github.com/moorara/markdown/parser"
	"github.com/moorara/markdown/render/html"
)

func TestPrefixHeaderNoExtensions(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderNoExtensions.tests")
	runBlockTests(t, tests, 0)
}

func TestPrefixHeaderSpaceExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderSpaceExtension.tests")
	runBlockTests(t, tests, parser.SpaceHeadings)
}

func TestPrefixHeaderIdExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderIdExtension.tests")
	runBlockTests(t, tests, parser.HeadingIDs)
}

func TestPrefixHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderIdExtensionWithPrefixAndSuffix.tests")

	parameters := html.RendererOptions{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	runParamTests(t, tests, TestParams{
		extensions:      parser.HeadingIDs,
		Flags:           html.UseXHTML,
		RendererOptions: parameters,
	})
}

func TestPrefixAutoHeaderIdExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixAutoHeaderIdExtension.tests")
	runBlockTests(t, tests, parser.AutoHeadingIDs)
}

func TestPrefixAutoHeaderIdExtensionWithPrefixAndSuffix(t *testing.T) {
	tests := readTestFile2(t, "PrefixAutoHeaderIdExtensionWithPrefixAndSuffix.tests")
	parameters := html.RendererOptions{
		HeadingIDPrefix: "PRE:",
		HeadingIDSuffix: ":POST",
	}

	runParamTests(t, tests, TestParams{
		extensions:      parser.AutoHeadingIDs,
		Flags:           html.UseXHTML,
		RendererOptions: parameters,
	})
}

func TestPrefixMultipleHeaderExtensions(t *testing.T) {
	tests := readTestFile2(t, "PrefixMultipleHeaderExtensions.tests")
	runBlockTests(t, tests, parser.AutoHeadingIDs|parser.HeadingIDs)
}

func TestPrefixHeaderMmarkExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderMmarkExtension.tests")
	runBlockTests(t, tests, parser.Mmark)
}

func TestUnderlineHeaders(t *testing.T) {
	tests := readTestFile2(t, "UnderlineHeaders.tests")
	runBlockTests(t, tests, 0)
}

func TestUnderlineHeadersAutoIDs(t *testing.T) {
	tests := readTestFile2(t, "UnderlineHeadersAutoIDs.tests")
	runBlockTests(t, tests, parser.AutoHeadingIDs)
}

func TestHorizontalRule(t *testing.T) {
	tests := readTestFile2(t, "HorizontalRule.tests")
	runBlockTests(t, tests, 0)
}

func TestUnorderedList(t *testing.T) {
	tests := readTestFile2(t, "UnorderedList.tests")
	runBlockTests(t, tests, 0)
}

func TestOrderedList(t *testing.T) {
	tests := readTestFile2(t, "OrderedList.tests")
	runBlockTests(t, tests, 0)
}

func TestDefinitionList(t *testing.T) {
	tests := readTestFile2(t, "DefinitionList.tests")
	runBlockTests(t, tests, parser.DefinitionLists)
}

func TestNestedDefinitionList(t *testing.T) {
	tests := readTestFile2(t, "NestedDefinitionList.tests")
	runBlockTests(t, tests, parser.DefinitionLists)
}

func TestPreformattedHtml(t *testing.T) {
	tests := readTestFile2(t, "PreformattedHtml.tests")
	runBlockTests(t, tests, 0)
}

func TestPreformattedHtmlLax(t *testing.T) {
	tests := readTestFile2(t, "PreformattedHtmlLax.tests")
	runBlockTests(t, tests, parser.LaxHTMLBlocks)
}

func TestFencedCodeBlock(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeBlock.tests")
	runBlockTests(t, tests, parser.FencedCode)
}

func TestFencedCodeInsideBlockquotes(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeInsideBlockquotes.tests")
	runBlockTests(t, tests, parser.FencedCode)
}

func TestTable(t *testing.T) {
	tests := readTestFile2(t, "Table.tests")
	runBlockTests(t, tests, parser.Tables)
}

func TestUnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "UnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	runBlockTests(t, tests, parser.NoEmptyLineBeforeBlock)
}

func TestOrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "OrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	runBlockTests(t, tests, parser.NoEmptyLineBeforeBlock)
}

func TestFencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	runBlockTests(t, tests, parser.FencedCode|parser.NoEmptyLineBeforeBlock)
}

func TestMathBlock(t *testing.T) {
	tests := readTestFile2(t, "MathBlock.tests")
	runBlockTests(t, tests, parser.CommonExtensions)
}

func TestDefinitionListWithFencedCodeBlock(t *testing.T) {
	tests := readTestFile2(t, "DefinitionListWithFencedCodeBlock.tests")
	runBlockTests(t, tests, parser.FencedCode|parser.DefinitionLists)
}

func TestListWithFencedCodeBlockAndHeader(t *testing.T) {
	tests := readTestFile2(t, "ListWithFencedCodeBlockAndHeader.tests")
	runBlockTests(t, tests, parser.FencedCode)
}

func TestTitleBlock_EXTENSION_TITLEBLOCK(t *testing.T) {
	tests := readTestFile2(t, "TitleBlock_EXTENSION_TITLEBLOCK.tests")
	runBlockTests(t, tests, parser.Titleblock)
}

func TestBlockComments(t *testing.T) {
	tests := readTestFile2(t, "BlockComments.tests")
	runBlockTests(t, tests, 0)
}

func TestBug168(t *testing.T) {
	tests := readTestFile2(t, "bug168.tests")
	runBlockTests(t, tests, 0)
}

func TestTOC(t *testing.T) {
	tests := readTestFile2(t, "TOC.tests")
	runParamTests(t, tests, TestParams{
		extensions: parser.HeadingIDs,
		Flags:      html.UseXHTML | html.TOC,
	})
}

func TestCompletePage(t *testing.T) {
	tests := readTestFile2(t, "CompletePage.tests")
	runParamTests(t, tests, TestParams{Flags: html.UseXHTML | html.CompletePage})
}

func TestSpaceHeadings(t *testing.T) {
	tests := readTestFile2(t, "SpaceHeadings.tests")
	runParamTests(t, tests, TestParams{extensions: parser.SpaceHeadings})
}

func TestCodeInList(t *testing.T) {
	tests := readTestFile2(t, "code_in_list.test")
	exts := parser.CommonExtensions
	runParamTests(t, tests, TestParams{extensions: exts})
}

func TestLists(t *testing.T) {
	tests := readTestFile2(t, "Lists.tests")
	exts := parser.CommonExtensions
	runParamTests(t, tests, TestParams{extensions: exts})
}

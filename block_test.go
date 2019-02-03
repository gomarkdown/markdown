package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func writeTest(file string, tests []string) {
	path := filepath.Join("testdata", file)
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	lastIdx := len(tests) - 1
	for i, s := range tests {
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		fmt.Fprint(f, s)
		if i != lastIdx {
			fmt.Fprint(f, "+++\n")
		}
	}
}

func TestPrefixHeaderNoExtensions(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderNoExtensions.tests")
	doTestsBlock(t, tests, 0)
}

func TestPrefixHeaderSpaceExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderSpaceExtension.tests")
	doTestsBlock(t, tests, parser.SpaceHeadings)
}

func TestPrefixHeaderIdExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderIdExtension.tests")
	doTestsBlock(t, tests, parser.HeadingIDs)
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
	tests := readTestFile2(t, "PrefixAutoHeaderIdExtension.tests")
	doTestsBlock(t, tests, parser.AutoHeadingIDs)
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
	tests := readTestFile2(t, "PrefixMultipleHeaderExtensions.tests")
	doTestsBlock(t, tests, parser.AutoHeadingIDs|parser.HeadingIDs)
}

func TestPrefixHeaderMmarkExtension(t *testing.T) {
	tests := readTestFile2(t, "PrefixHeaderMmarkExtension.tests")
	doTestsBlock(t, tests, parser.Mmark)
}

func TestUnderlineHeaders(t *testing.T) {
	tests := readTestFile2(t, "UnderlineHeaders.tests")
	doTestsBlock(t, tests, 0)
}

func TestUnderlineHeadersAutoIDs(t *testing.T) {
	tests := readTestFile2(t, "UnderlineHeadersAutoIDs.tests")
	doTestsBlock(t, tests, parser.AutoHeadingIDs)
}

func TestHorizontalRule(t *testing.T) {
	tests := readTestFile2(t, "HorizontalRule.tests")
	doTestsBlock(t, tests, 0)
}

func TestUnorderedList(t *testing.T) {
	tests := readTestFile2(t, "UnorderedList.tests")
	doTestsBlock(t, tests, 0)
}

func TestOrderedList(t *testing.T) {
	tests := readTestFile2(t, "OrderedList.tests")
	doTestsBlock(t, tests, 0)
}

func TestDefinitionList(t *testing.T) {
	tests := readTestFile2(t, "DefinitionList.tests")
	doTestsBlock(t, tests, parser.DefinitionLists)
}

func TestNestedDefinitionList(t *testing.T) {
	tests := readTestFile2(t, "NestedDefinitionList.tests")
	doTestsBlock(t, tests, parser.DefinitionLists)
}

func TestPreformattedHtml(t *testing.T) {
	tests := readTestFile2(t, "PreformattedHtml.tests")
	doTestsBlock(t, tests, 0)
}

func TestPreformattedHtmlLax(t *testing.T) {
	tests := readTestFile2(t, "PreformattedHtmlLax.tests")
	doTestsBlock(t, tests, parser.LaxHTMLBlocks)
}

func TestFencedCodeBlock(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeBlock.tests")
	doTestsBlock(t, tests, parser.FencedCode)
}

func TestFencedCodeInsideBlockquotes(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeInsideBlockquotes.tests")
	doTestsBlock(t, tests, parser.FencedCode)
}

func TestTable(t *testing.T) {
	tests := readTestFile2(t, "Table.tests")
	doTestsBlock(t, tests, parser.Tables)
}

func TestUnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "UnorderedListWith_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	doTestsBlock(t, tests, parser.NoEmptyLineBeforeBlock)
}

func TestOrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "OrderedList_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	doTestsBlock(t, tests, parser.NoEmptyLineBeforeBlock)
}

func TestFencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	tests := readTestFile2(t, "FencedCodeBlock_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK.tests")
	doTestsBlock(t, tests, parser.FencedCode|parser.NoEmptyLineBeforeBlock)
}

func TestMathBlock(t *testing.T) {
	tests := readTestFile2(t, "MathBlock.tests")
	doTestsBlock(t, tests, parser.CommonExtensions)
}

func TestDefinitionListWithFencedCodeBlock(t *testing.T) {
	tests := readTestFile2(t, "DefinitionListWithFencedCodeBlock.tests")
	doTestsBlock(t, tests, parser.FencedCode|parser.DefinitionLists)
}

func TestListWithFencedCodeBlockAndHeader(t *testing.T) {
	tests := readTestFile2(t, "ListWithFencedCodeBlockAndHeader.tests")
	doTestsBlock(t, tests, parser.FencedCode)
}

func TestTitleBlock_EXTENSION_TITLEBLOCK(t *testing.T) {
	tests := readTestFile2(t, "TitleBlock_EXTENSION_TITLEBLOCK.tests")
	doTestsBlock(t, tests, parser.Titleblock)
}

func TestBlockComments(t *testing.T) {
	tests := readTestFile2(t, "BlockComments.tests")
	doTestsBlock(t, tests, 0)
}

func TestTOC(t *testing.T) {
	tests := readTestFile2(t, "TOC.tests")
	doTestsParam(t, tests, TestParams{
		Flags: html.UseXHTML | html.TOC,
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

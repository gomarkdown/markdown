package markdown

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

// Markdown 1.0.3 reference tests

var (
	refFiles = []string{
		"Amps and angle encoding",
		"Auto links",
		"Backslash escapes",
		"Blockquotes with code blocks",
		"Code Blocks",
		"Code Spans",
		"Horizontal rules",
		"Inline HTML (Advanced)",
		"Inline HTML (Simple)",
		"Inline HTML comments",
		"Links, inline style",
		"Links, reference style",
		"Links, shortcut references",
		"Literal quotes in titles",
		"Markdown Documentation - Basics",
		"Markdown Documentation - Syntax",
		"Nested blockquotes",
		"Ordered and unordered lists",
		"Strong and em together",
		"Tabs",
		"Tidyness",
		// New tests added for gomarkdown
		"Entities",
	}
)

func TestReference(t *testing.T) {
	files := append(refFiles, "Hard-wrapped paragraphs with list-like lines")
	doTestsReference(t, files, 0)
}

func TestReference_EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK(t *testing.T) {
	files := append(refFiles, "Hard-wrapped paragraphs with list-like lines no empty line before block")
	doTestsReference(t, files, parser.NoEmptyLineBeforeBlock)
}

// benchResultAnchor is an anchor variable to store the result of a benchmarked
// code so that compiler could never optimize away the call to runMarkdown()
var benchResultAnchor string

func benchFile(b *testing.B, basename string) {
	params := TestParams{extensions: parser.CommonExtensions}
	filename := filepath.Join("testdata", basename+".text")
	inputBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		b.Errorf("Couldn't open '%s', error: %v\n", filename, err)
		return
	}

	test := string(inputBytes)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchResultAnchor = runMarkdown(test, params)
	}
}

func BenchmarkReferenceAmps(b *testing.B) {
	benchFile(b, "Amps and angle encoding")
}

func BenchmarkReferenceAutoLinks(b *testing.B) {
	benchFile(b, "Auto links")
}

func BenchmarkReferenceBackslashEscapes(b *testing.B) {
	benchFile(b, "Backslash escapes")
}

func BenchmarkReferenceBlockquotesWithCodeBlocks(b *testing.B) {
	benchFile(b, "Blockquotes with code blocks")
}

func BenchmarkReferenceCodeBlocks(b *testing.B) {
	benchFile(b, "Code Blocks")
}

func BenchmarkReferenceCodeSpans(b *testing.B) {
	benchFile(b, "Code Spans")
}

func BenchmarkIssue265(b *testing.B) {
	benchFile(b, "issue265-slow-binary")
}

func BenchmarkReferenceHardWrappedPara(b *testing.B) {
	benchFile(b, "Hard-wrapped paragraphs with list-like lines")
}
func BenchmarkReferenceHorizontalRules(b *testing.B) {
	benchFile(b, "Horizontal rules")
}
func BenchmarkReferenceInlineHTMLAdvances(b *testing.B) {
	benchFile(b, "Inline HTML (Advanced)")
}
func BenchmarkReferenceInlineHTMLSimple(b *testing.B) {
	benchFile(b, "Inline HTML (Simple)")
}
func BenchmarkReferenceInlineHTMLComments(b *testing.B) {
	benchFile(b, "Inline HTML comments")
}
func BenchmarkReferenceLinksInline(b *testing.B) {
	benchFile(b, "Links, inline style")
}
func BenchmarkReferenceLinksReference(b *testing.B) {
	benchFile(b, "Links, reference style")
}
func BenchmarkReferenceLinksShortcut(b *testing.B) {
	benchFile(b, "Links, shortcut references")
}
func BenchmarkReferenceLiterQuotesInTitles(b *testing.B) {
	benchFile(b, "Literal quotes in titles")
}

func BenchmarkReferenceMarkdownBasics(b *testing.B) {
	benchFile(b, "Markdown Documentation - Basics")
}

func BenchmarkReferenceMarkdownSyntax(b *testing.B) {
	benchFile(b, "Markdown Documentation - Syntax")
}

func BenchmarkReferenceNestedBlockquotes(b *testing.B) {
	benchFile(b, "Nested blockquotes")
}

func BenchmarkReferenceOrderedAndUnorderedLists(b *testing.B) {
	benchFile(b, "Ordered and unordered lists")
}

func BenchmarkReferenceStrongAndEm(b *testing.B) {
	benchFile(b, "Strong and em together")
}

func BenchmarkReferenceTabs(b *testing.B) {
	benchFile(b, "Tabs")
}

func BenchmarkReferenceTidyness(b *testing.B) {
	benchFile(b, "Tidyness")
}

func BenchmarkReference(b *testing.B) {
	params := TestParams{extensions: parser.CommonExtensions}
	files := append(refFiles, "Hard-wrapped paragraphs with list-like lines")
	var tests []string
	for _, basename := range files {
		filename := filepath.Join("testdata", basename+".text")
		inputBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			b.Errorf("Couldn't open '%s', error: %v\n", filename, err)
			continue
		}
		tests = append(tests, string(inputBytes))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, test := range tests {
			benchResultAnchor = runMarkdown(test, params)
		}
	}
}

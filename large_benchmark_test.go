package markdown

import (
	"bytes"
	"os"
	"testing"

	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var largeDocumentResult []byte

func readLargeBenchmarkDocument(b *testing.B) []byte {
	b.Helper()
	files := []string{
		"testdata/Markdown Documentation - Syntax.text",
		"testdata/Markdown Documentation - Basics.text",
		"testdata/issue330-crasher.md",
		"testdata/Backslash escapes.text",
		"testdata/Inline HTML (Simple).text",
		"testdata/Inline HTML (Advanced).text",
		"testdata/Links, inline style.text",
		"testdata/Links, reference style.text",
		"testdata/Ordered and unordered lists.text",
		"testdata/md1.md",
		"testdata/md2.md",
	}

	var buf bytes.Buffer
	for i := 0; i < 16; i++ {
		buf.WriteString("\n\n<!-- large fixture iteration -->\n\n")
		for _, file := range files {
			buf.WriteString("\n\n<!-- source: ")
			buf.WriteString(file)
			buf.WriteString(" -->\n\n")
			data, err := os.ReadFile(file)
			if err != nil {
				b.Fatal(err)
			}
			buf.Write(data)
		}
	}
	return buf.Bytes()
}

func BenchmarkLargeDocumentParse(b *testing.B) {
	data := readLargeBenchmarkDocument(b)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := parser.NewWithExtensions(parser.CommonExtensions).Parse(data)
		if doc == nil {
			b.Fatal("nil document")
		}
	}
}

func BenchmarkLargeDocumentToHTML(b *testing.B) {
	data := readLargeBenchmarkDocument(b)
	opts := html.RendererOptions{Flags: html.CommonFlags}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewWithExtensions(parser.CommonExtensions)
		renderer := html.NewRenderer(opts)
		largeDocumentResult = ToHTML(data, p, renderer)
	}
}

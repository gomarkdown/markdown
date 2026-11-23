//go:build gofuzz

package markdown

import (
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Fuzz is the entry point for https://github.com/dvyukov/go-fuzz.
// Returning 1 marks a non-empty document so the fuzzer prefers it.
// Rendering is part of the call so a panic in the HTML renderer is a crash.
func Fuzz(data []byte) int {
	legacy := parser.NewWithExtensions(parser.CommonExtensions)
	commonMark := parser.New()
	commonMark.Opts.Flags = parser.CommonMark
	if fuzzOne(data, legacy) || fuzzOne(data, commonMark) {
		return 1
	}
	return 0
}

func fuzzOne(data []byte, p *parser.Parser) bool {
	doc := Parse(data, p)
	Render(doc, html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags}))
	return doc != nil && len(doc.GetChildren()) > 0
}

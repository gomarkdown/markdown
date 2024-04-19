package main

// example for https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"

	"bytes"
	"net/url"
)

// mds shows the two extensions provided in this example. The tricky
// part about the wiki link example is that we need to defer to the
// original code in the library if the opening square bracket we're
// looking at is not a wiki link.
var mds = `This is a [[wiki link]].

#markdown #example`

// wikiLink returns an inline parser function. This indirection is
// required because we want to call the previous definition in case
// this is not a wikiLink.
func wikiLink(p *parser.Parser,	fn parser.InlineParser) parser.InlineParser {
	return func (p *parser.Parser, original []byte, offset int) (int, ast.Node) {
		data := original[offset:]
		n := len(data)
		// minimum: [[X]]
		if n < 5 || data[1] != '[' {
			return fn(p, original, offset)
		}
		i := 2
		for i+1 < n && data[i] != ']' && data[i+1] != ']' {
			i++
		}
		text := data[2:i+1]
		link := &ast.Link{
			Destination: []byte(url.PathEscape(string(text))),
		}
		ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
		return i+3, link
	}
}

// hashtag links every hashtag to a search for that tag. How this
// search is then handled is of course a different problem. For your
// own code, you probably need to use a different destination URL.
func hashtag(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]
	i := 0
	n := len(data)
	for i < n && !parser.IsSpace(data[i]) {
		i++
	}
	if i == 0 {
		return 0, nil
	}
	link := &ast.Link{
		Destination: append([]byte("/search?q=%23"), data[1:i]...),
	}
	text := bytes.ReplaceAll(data[0:i], []byte("_"), []byte(" "))
	ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: text}})
	return i, link
}

func inlineExample() {
	md := []byte(mds)

	parser := parser.New()
	prev := parser.RegisterInline('[', nil)
	parser.RegisterInline('[', wikiLink(parser, prev))
	parser.RegisterInline('#', hashtag)
	html := markdown.ToHTML(md, parser, nil)

	fmt.Printf("--- Markdown:\n%s\n\n--- HTML:\n%s\n", md, html)
}

func main() {
	inlineExample()
}

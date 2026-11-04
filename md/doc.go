/*
Package md implements a renderer that writes a parsed markdown document back
as Markdown.

	p := parser.NewWithExtensions(parser.CommonExtensions)
	doc := p.Parse(input)
	output := markdown.Render(doc, md.NewRenderer())

Use it to reformat a document, or to write one after changing its AST.

The output parses to the same HTML as the document did, with the same kind of
parser: the parser of the parser package with extensions, or the CommonMark
parser from parser.NewCommonMark. It escapes what would otherwise be read as
Markdown syntax, which differs between those parsers.

The output is a normalized form of the document: the lines of a paragraph are
joined, emphasis is written with *, headings with # and code blocks with
fences. These don't round-trip:

  - whitespace that HTML ignores
  - the leading . of a code block's info string: {.go} is written as go
  - emphasis nested three levels deep with punctuation next to the delimiters

# Links in the footer

With WithRenderInFooter(true), links are written as [text][label] with the
definitions of the labels at the end of the document, instead of inline. Links
that the parser of the parser package can't read from a definition, because
their destination is empty or has spaces, stay inline.
*/
package md

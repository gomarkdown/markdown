package parser

import (
	"bufio"
	"bytes"
	"fmt"
	stdhtml "html"

	"github.com/gomarkdown/markdown/ast"
	"github.com/yuin/goldmark"
	goldast "github.com/yuin/goldmark/ast"
	goldparser "github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// CommonMark uses a separate grammar so fixes to its block and delimiter rules
// cannot change the historical parser. Convert to our AST before returning so
// callers retain the same traversal, transformation, and renderer APIs.
func (p *Parser) parseCommonMark(input []byte) ast.Node {
	input = bytes.ReplaceAll(NormalizeNewlines(input), []byte{0}, []byte("\ufffd"))
	context := goldparser.NewContext()
	if p.ReferenceOverride != nil {
		context = &commonMarkContext{Context: context, override: p.ReferenceOverride}
	}
	root := goldmark.DefaultParser().Parse(text.NewReader(input), goldparser.WithContext(context))
	doc := p.Doc.(*ast.Document)
	doc.CommonMark = true
	commonMarkChildren(doc, root, input)
	p.tip = nil
	return doc
}

type commonMarkContext struct {
	goldparser.Context
	override ReferenceOverrideFunc
}

func (c *commonMarkContext) Reference(label string) (goldparser.Reference, bool) {
	if ref, overridden := c.override(label); overridden {
		if ref == nil {
			return nil, false
		}
		return goldparser.NewReference([]byte(label), []byte(ref.Link), []byte(ref.Title)), true
	}
	return c.Context.Reference(label)
}

// The upstream text writer resolves escapes and entities in a single pass.
// Decode its HTML escaping to obtain literal text for our renderer. In
// particular, an escaped ampersand must not start another entity expansion.
func commonMarkLiteral(source []byte) []byte {
	if source == nil {
		return nil
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	goldhtml.DefaultWriter.Write(w, source)
	w.Flush()
	return []byte(stdhtml.UnescapeString(buf.String()))
}

func appendCommonMarkText(parent ast.Node, literal []byte) {
	if len(literal) == 0 {
		return
	}
	if previous, ok := ast.GetLastChild(parent).(*ast.Text); ok {
		previous.Literal = append(previous.Literal, literal...)
		return
	}
	ast.AppendChild(parent, &ast.Text{Leaf: ast.Leaf{Literal: bytes.Clone(literal)}})
}

func commonMarkChildren(parent ast.Node, source goldast.Node, input []byte) {
	for child := source.FirstChild(); child != nil; child = child.NextSibling() {
		var node ast.Node
		switch n := child.(type) {
		case *goldast.Text:
			literal := n.Value(input)
			if !n.IsRaw() {
				literal = commonMarkLiteral(literal)
			}
			appendCommonMarkText(parent, literal)
			if n.HardLineBreak() {
				ast.AppendChild(parent, &ast.Hardbreak{})
			} else if n.SoftLineBreak() {
				appendCommonMarkText(parent, []byte{'\n'})
			}
			continue
		case *goldast.String:
			literal := n.Value
			if !n.IsRaw() && !n.IsCode() {
				literal = commonMarkLiteral(literal)
			}
			appendCommonMarkText(parent, literal)
			continue
		case *goldast.Paragraph, *goldast.TextBlock:
			node = &ast.Paragraph{}
		case *goldast.Heading:
			node = &ast.Heading{Level: n.Level}
		case *goldast.Blockquote:
			node = &ast.BlockQuote{}
		case *goldast.ThematicBreak:
			node = &ast.HorizontalRule{}
		case *goldast.List:
			list := &ast.List{Tight: n.IsTight, BulletChar: n.Marker}
			if n.IsOrdered() {
				list.ListFlags = ast.ListTypeOrdered
				list.Delimiter = n.Marker
				list.Start = n.Start
			}
			node = list
		case *goldast.ListItem:
			list := parent.(*ast.List)
			item := &ast.ListItem{ListFlags: list.ListFlags, Tight: list.Tight,
				BulletChar: list.BulletChar, Delimiter: list.Delimiter}
			if n.PreviousSibling() == nil {
				item.ListFlags |= ast.ListItemBeginningOfList
			}
			if n.NextSibling() == nil {
				item.ListFlags |= ast.ListItemEndOfList
			}
			if !list.Tight {
				item.ListFlags |= ast.ListItemContainsBlock
			}
			node = item
		case *goldast.CodeBlock:
			node = &ast.CodeBlock{Leaf: ast.Leaf{Literal: n.Lines().Value(input)}}
		case *goldast.FencedCodeBlock:
			code := &ast.CodeBlock{IsFenced: true, Leaf: ast.Leaf{Literal: n.Lines().Value(input)}}
			if n.Info != nil {
				code.Info = commonMarkLiteral(n.Info.Value(input))
			}
			node = code
		case *goldast.HTMLBlock:
			literal := n.Lines().Value(input)
			if n.HasClosure() {
				literal = append(literal, n.ClosureLine.Value(input)...)
			}
			node = &ast.HTMLBlock{Leaf: ast.Leaf{Literal: literal}}
		case *goldast.RawHTML:
			node = &ast.HTMLSpan{Leaf: ast.Leaf{Literal: n.Segments.Value(input)}}
		case *goldast.CodeSpan:
			var literal []byte
			for part := n.FirstChild(); part != nil; part = part.NextSibling() {
				literal = append(literal, part.(*goldast.Text).Value(input)...)
			}
			literal = bytes.ReplaceAll(literal, []byte{'\n'}, []byte{' '})
			ast.AppendChild(parent, &ast.Code{Leaf: ast.Leaf{Literal: literal}})
			continue
		case *goldast.Emphasis:
			if n.Level == 2 {
				node = &ast.Strong{}
			} else {
				node = &ast.Emph{}
			}
		case *goldast.Link:
			node = &ast.Link{Destination: util.URLEscape(n.Destination, true), Title: commonMarkLiteral(n.Title)}
		case *goldast.Image:
			node = &ast.Image{Destination: util.URLEscape(n.Destination, true), Title: commonMarkLiteral(n.Title)}
		case *goldast.AutoLink:
			destination := util.URLEscape(n.URL(input), false)
			if n.AutoLinkType == goldast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(destination), []byte("mailto:")) {
				destination = append([]byte("mailto:"), destination...)
			}
			link := &ast.Link{Destination: destination}
			ast.AppendChild(parent, link)
			appendCommonMarkText(link, n.Label(input))
			continue
		case *goldast.LinkReferenceDefinition:
			node = &ast.ReferenceDefinition{Label: bytes.Clone(n.Label),
				Destination: util.URLEscape(n.Destination, true), Title: commonMarkLiteral(n.Title),
				Leaf: ast.Leaf{Literal: n.Lines().Value(input)}}
		default:
			panic(fmt.Sprintf("unsupported CommonMark node %T", child))
		}
		ast.AppendChild(parent, node)
		if node.AsContainer() != nil {
			commonMarkChildren(node, child, input)
		}
	}
}

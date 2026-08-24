package parser

import (
	"bytes"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestReferenceDefinitionInAST(t *testing.T) {
	p := New()
	doc := p.Parse([]byte("Hello [world][wiki-world]\n\n[wiki-world]: https://wikipedia.com/blah-blah-blah \"Wiki\"\n"))
	var def *ast.ReferenceDefinition
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if d, ok := node.(*ast.ReferenceDefinition); ok && entering {
			def = d
		}
		return ast.GoToNext
	})
	if def == nil {
		t.Fatal("missing ReferenceDefinition node")
	}
	if got := string(def.Label); got != "wiki-world" {
		t.Errorf("label %q", got)
	}
	if got := string(def.Destination); got != "https://wikipedia.com/blah-blah-blah" {
		t.Errorf("destination %q", got)
	}
	if got := string(def.Title); got != "Wiki" {
		t.Errorf("title %q", got)
	}

	var link *ast.Link
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if l, ok := node.(*ast.Link); ok && entering {
			link = l
		}
		return ast.GoToNext
	})
	if link == nil {
		t.Fatal("missing Link")
	}
	if got := string(link.Destination); got != "https://wikipedia.com/blah-blah-blah" {
		t.Errorf("link destination %q", got)
	}
	if got := string(link.DeferredID); got != "wiki-world" {
		t.Errorf("deferred id %q", got)
	}
}

func TestReferenceDefinitionNotInHTML(t *testing.T) {
	// sanity: node type name is stable for debug printers
	var buf bytes.Buffer
	n := &ast.ReferenceDefinition{Label: []byte("x")}
	ast.Print(&buf, n)
	if !bytes.Contains(buf.Bytes(), []byte("ReferenceDefinition")) {
		t.Errorf("print %q", buf.Bytes())
	}
}

func TestReferenceDefinitionOrder(t *testing.T) {
	p := New()
	doc := p.Parse([]byte("Hello [world][wiki-world]\n[wiki-world]: https://example.com\n"))
	kids := doc.GetChildren()
	if len(kids) < 2 {
		t.Fatalf("children %d", len(kids))
	}
	if _, ok := kids[0].(*ast.Paragraph); !ok {
		t.Errorf("first child %T, want Paragraph", kids[0])
	}
	if _, ok := kids[1].(*ast.ReferenceDefinition); !ok {
		t.Errorf("second child %T, want ReferenceDefinition", kids[1])
	}
}

func TestReferenceDefinitionSkipsFootnotes(t *testing.T) {
	p := NewWithExtensions(CommonExtensions | Footnotes)
	doc := p.Parse([]byte("Hi[^a]\n\n[^a]: a note\n"))
	var def *ast.ReferenceDefinition
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if d, ok := node.(*ast.ReferenceDefinition); ok && entering {
			def = d
		}
		return ast.GoToNext
	})
	if def != nil {
		t.Fatalf("footnote should not create ReferenceDefinition, got %+v", def)
	}
}

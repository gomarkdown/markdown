package markdown

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func TestCommonMark(t *testing.T) {
	data, err := os.ReadFile("testdata/commonmark-0.31.2.json")
	if err != nil {
		t.Fatal(err)
	}
	var examples []struct {
		Markdown string
		HTML     string
		Example  int
		Section  string
	}
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatal(err)
	}
	if len(examples) != 652 {
		t.Fatalf("expected the complete 652-example suite, got %d", len(examples))
	}
	for _, extensions := range []parser.Extensions{parser.NoExtensions, parser.CommonExtensions} {
		t.Run(fmt.Sprintf("extensions_%d", extensions), func(t *testing.T) {
			for _, example := range examples {
				t.Run(fmt.Sprintf("%s/%03d", example.Section, example.Example), func(t *testing.T) {
					p := parser.NewWithExtensions(extensions)
					p.Opts.Flags = parser.CommonMark
					r := html.NewRenderer(html.RendererOptions{Flags: html.UseXHTML})
					got := string(ToHTML([]byte(example.Markdown), p, r))
					if got != example.HTML {
						t.Errorf("markdown: %q\nwant: %q\n got: %q", example.Markdown, example.HTML, got)
					}
					p = parser.NewWithExtensions(extensions)
					p.Opts.Flags = parser.CommonMark
					if got := string(ToHTML([]byte(example.Markdown), p, nil)); got != example.HTML {
						t.Errorf("default renderer, markdown: %q\nwant: %q\n got: %q", example.Markdown, example.HTML, got)
					}
				})
			}
		})
	}
}

func TestCommonMarkASTIntegration(t *testing.T) {
	p := parser.New()
	p.Opts.Flags = parser.CommonMark
	doc := Parse([]byte("# heading\n\n[old][ref]\n\n[ref]: /original \"title\"\n"), p)
	var foundHeading, foundLink, foundDefinition bool
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}
		for i, child := range node.GetChildren() {
			if child.GetParent() != node {
				t.Fatal("incorrect AST parent")
			}
			if i > 0 && ast.GetPrevNode(child) != node.GetChildren()[i-1] {
				t.Fatal("incorrect AST previous sibling")
			}
			if i+1 < len(node.GetChildren()) && ast.GetNextNode(child) != node.GetChildren()[i+1] {
				t.Fatal("incorrect AST next sibling")
			}
		}
		switch n := node.(type) {
		case *ast.Heading:
			foundHeading = true
		case *ast.Link:
			foundLink = true
			n.Destination = []byte("/changed")
		case *ast.ReferenceDefinition:
			foundDefinition = true
			if string(n.Label) != "ref" || string(n.Destination) != "/original" || string(n.Title) != "title" {
				t.Fatalf("incorrect reference definition: %+v", n)
			}
		}
		return ast.GoToNext
	})
	if !foundHeading || !foundLink || !foundDefinition {
		t.Fatal("missing standard AST nodes")
	}
	r := html.NewRenderer(html.RendererOptions{
		RenderNodeHook: func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			if _, ok := node.(*ast.Heading); ok {
				if entering {
					io.WriteString(w, "<header>custom</header>\n")
				}
				return ast.SkipChildren, true
			}
			return ast.GoToNext, false
		},
	})
	want := "<header>custom</header>\n<p><a href=\"/changed\" title=\"title\">old</a></p>\n"
	if got := string(Render(doc, r)); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestCommonMarkReferenceOverride(t *testing.T) {
	p := parser.New()
	p.Opts.Flags = parser.CommonMark
	p.ReferenceOverride = func(label string) (*parser.Reference, bool) {
		switch label {
		case "override":
			return &parser.Reference{Link: "/override", Title: "new title"}, true
		case "reject":
			return nil, true
		default:
			return nil, false
		}
	}
	input := "[override] [reject] [fallback]\n\n[override]: /old\n[reject]: /old\n[fallback]: /fallback\n"
	want := "<p><a href=\"/override\" title=\"new title\">override</a> [reject] <a href=\"/fallback\">fallback</a></p>\n"
	if got := string(ToHTML([]byte(input), p, nil)); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestCommonMarkNewlinesAndNUL(t *testing.T) {
	for _, input := range []string{"a\x00b\nc", "a\x00b\rc", "a\x00b\r\nc"} {
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		want := "<p>a\ufffdb\nc</p>\n"
		if got := string(ToHTML([]byte(input), p, nil)); got != want {
			t.Errorf("input %q: want %q, got %q", input, want, got)
		}
	}
}

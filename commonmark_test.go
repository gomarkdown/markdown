package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type commonMarkExample struct {
	Markdown string
	HTML     string
	Example  int
	Section  string
}

func commonMarkExamples(t testing.TB) []commonMarkExample {
	t.Helper()
	data, err := os.ReadFile("testdata/commonmark-0.31.2.json")
	if err != nil {
		t.Fatal(err)
	}
	var examples []commonMarkExample
	if err := json.Unmarshal(data, &examples); err != nil {
		t.Fatal(err)
	}
	if len(examples) != 652 {
		t.Fatalf("expected the complete 652-example suite, got %d", len(examples))
	}
	return examples
}

func TestCommonMark(t *testing.T) {
	examples := commonMarkExamples(t)
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

func TestCommonMarkEmptyTitles(t *testing.T) {
	for _, input := range []string{"[link](/url \"\")", "[link][ref]\n\n[ref]: /url \"\""} {
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		want := "<p><a href=\"/url\" title=\"\">link</a></p>\n"
		if got := string(ToHTML([]byte(input), p, nil)); got != want {
			t.Errorf("input %q: want %q, got %q", input, want, got)
		}
	}
}

func TestCommonMarkUnicodeReferences(t *testing.T) {
	for _, pair := range [][2]string{{"ß", "SS"}, {"ẞ", "ss"}, {"ς", "Σ"}, {"ſ", "S"}, {"ﬃ", "FFI"}, {"İ", "i\u0307"}, {"ǰ", "j\u030c"}} {
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		input := "[" + pair[0] + "]\n\n[" + pair[1] + "]: /url"
		want := "<p><a href=\"/url\">" + pair[0] + "</a></p>\n"
		if got := string(ToHTML([]byte(input), p, nil)); got != want {
			t.Errorf("input %q: want %q, got %q", input, want, got)
		}
	}
	for _, label := range []string{"ı", "i", "a\u00a0b"} {
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		definition := "I"
		if label == "i" {
			definition = "İ"
		}
		if label == "a\u00a0b" {
			definition = "a b"
		}
		input := "[" + label + "]\n\n[" + definition + "]: /url"
		want := "<p>[" + label + "]</p>\n"
		if got := string(ToHTML([]byte(input), p, nil)); got != want {
			t.Errorf("distinct labels matched: input %q, got %q", input, got)
		}
	}
}

func TestCommonMarkEdgeCases(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"\v", "<p>\v</p>\n"},
		{"&notit; &semi;", "<p>&amp;notit; ;</p>\n"},
		{"[outer <https://example.com>](/url)", "<p>[outer <a href=\"https://example.com\">https://example.com</a>](/url)</p>\n"},
		{"[ref]\n\n[ref]: /url\n\n\"title\"", "<p><a href=\"/url\">ref</a></p>\n<p>&quot;title&quot;</p>\n"},
	} {
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		if got := string(ToHTML([]byte(tc.input), p, nil)); got != tc.want {
			t.Errorf("input %q: want %q, got %q", tc.input, tc.want, got)
		}
	}
}

func TestCommonMarkAdversarialInputs(t *testing.T) {
	for name, input := range map[string]string{
		"open brackets":                    strings.Repeat("[", 32768),
		"balanced brackets":                strings.Repeat("[", 16384) + strings.Repeat("]", 16384),
		"unclosed angles":                  "text " + strings.Repeat("<", 32768),
		"unclosed comments":                "text " + strings.Repeat("<!-- ", 8192),
		"unclosed processing instructions": "text " + strings.Repeat("<? ", 8192),
		"unclosed CDATA":                   "text " + strings.Repeat("<![CDATA[ ", 8192),
		"unclosed destinations":            strings.Repeat("[a](", 8192),
		"unclosed titles":                  strings.Repeat("[a](b \" ", 8192),
		"unmatched delimiters":             strings.Repeat("a_b*c ", 8192),
		"nested emphasis":                  strings.Repeat("*", 16384) + "a" + strings.Repeat("*", 16384),
		"backtick runs":                    strings.Repeat("`", 32768),
		"nested containers":                strings.Repeat("> ", 8192) + "a\n",
	} {
		t.Run(name, func(t *testing.T) {
			p := parser.New()
			p.Opts.Flags = parser.CommonMark
			parseWithParserShortTimeout(t, input, p)
		})
	}
}

func FuzzCommonMark(f *testing.F) {
	for _, example := range commonMarkExamples(f) {
		f.Add([]byte(example.Markdown))
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		original := bytes.Clone(input)
		p := parser.New()
		p.Opts.Flags = parser.CommonMark
		doc := p.Parse(input)
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
			return ast.GoToNext
		})
		Render(doc, html.NewRenderer(html.RendererOptions{Flags: html.UseXHTML}))
		if !bytes.Equal(input, original) {
			t.Fatal("parser modified the input")
		}
	})
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

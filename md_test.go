package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
)

// collapseSpace replaces each run of whitespace with one space, as HTML
// rendering does. md.Renderer joins the lines of a paragraph.
func collapseSpace(b []byte) string {
	return strings.Join(strings.Fields(string(b)), " ")
}

func TestMd(t *testing.T) {
	newParser := func() *parser.Parser {
		return parser.NewWithExtensions(parser.CommonExtensions | parser.OrderedListStart)
	}
	for _, file := range []string{"md1", "md2"} {
		pathOrig := filepath.Join("testdata", file+".md")
		d, err := os.ReadFile(pathOrig)
		if err != nil {
			t.Fatalf("os.ReadFile('%s') failed with %s", pathOrig, err)
		}
		pathExp := filepath.Join("testdata", file+"_exp.md")
		exp, err := os.ReadFile(pathExp)
		if err != nil {
			t.Fatalf("os.ReadFile('%s') failed with %s", pathExp, err)
		}
		// working copies may be CRLF (core.autocrlf)
		d, exp = NormalizeNewlines(d), NormalizeNewlines(exp)
		doc := Parse(d, newParser())
		got := Render(doc, md.NewRenderer())

		wantHTML := ToHTML(d, newParser(), nil)
		gotHTML := ToHTML(got, newParser(), nil)
		if collapseSpace(gotHTML) != collapseSpace(wantHTML) {
			t.Errorf("%s: rendered Markdown has different HTML\nwant: %s\n got: %s", pathOrig, wantHTML, gotHTML)
		}

		if !bytes.Equal(got, exp) {
			pathGot := filepath.Join("testdata", file+"_got.md")
			if err := os.WriteFile(pathGot, got, 0644); err != nil {
				t.Errorf("os.WriteFile('%s') failed with %s", pathGot, err)
			}
			pathAST := filepath.Join("testdata", file+"_ast.txt")
			f, err := os.Create(pathAST)
			if err != nil {
				t.Errorf("os.Create('%s') failed with %s", pathAST, err)
				continue
			}
			ast.PrintWithPrefix(f, doc, "  ")
			f.Close()
			t.Errorf("Failed md-to-md on %s. Compare expected %s to got %s and ast %s", pathOrig, pathExp, pathGot, pathAST)
		}
	}
}

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
		Title: data[0:i],
	}
	ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: data[0:i]}})
	return i + 1, link
}

func TestInlineParser(t *testing.T) {
	md := []byte(`#Haiku`)
	p := parser.New()
	p.RegisterInline('#', hashtag)
	html := ToHTML(md, p, nil)

	r := `<p><a href="/search?q=%23Haiku" title="#Haiku">#Haiku</a></p>
`
	if r != string(html) {
		t.Errorf("`%s`\n!=\n`%s`\n", string(html), r)
	}
}
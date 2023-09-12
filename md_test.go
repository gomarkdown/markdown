package markdown

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
)

func TestMd(t *testing.T) {
	if true {
		// disabled for now because md render is not finished
		return
	}
	files := []string{
		"md1",
		"md2",
	}
	for _, file := range files {
		pathOrig := filepath.Join("testdata", file+".md")
		d, err := ioutil.ReadFile(pathOrig)
		if err != nil {
			t.Fatalf("ioutil.ReadFile('%s') failed with %s", pathOrig, err)
		}
		pathExp := filepath.Join("testdata", file+"_exp.md")
		exp, err := ioutil.ReadFile(pathExp)
		if err != nil {
			t.Fatalf("ioutil.ReadFile('%s') failed with %s", pathExp, err)
		}
		exts := parser.CommonExtensions | parser.OrderedListStart
		parser := parser.NewWithExtensions(exts)
		renderer := md.NewRenderer()
		doc := Parse(d, parser)
		got := Render(doc, renderer)
		if !bytes.Equal(got, exp) {
			pathGot := filepath.Join("testdata", file+"_got.md")
			err = ioutil.WriteFile(pathGot, got, 0644)
			if err != nil {
				t.Errorf("ioutil.WriteFile('%s') failed with %s", pathGot, err)
			}
			pathAST := filepath.Join("testdata", file+"_ast.txt")
			f, err := os.Create(pathAST)
			if err != nil {
				t.Errorf("ioutil.WriteFile('%s') failed with %s", pathAST, err)
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
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// This prints AST of parsed markdown document.
// Usage: printast <markdown-file>

func usageAndExit() {
	fmt.Printf("Usage: printast [-to-html] <markdown-file>\n")
	os.Exit(1)
}

func main() {
	var (
		flgToHTML bool
	)
	{
		flag.BoolVar(&flgToHTML, "to-html", false, "convert to HTML")
		flag.Parse()
	}

	files := flag.Args()
	if len(files) < 1 {
		usageAndExit()
	}
	for _, fileName := range files {
		d, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open '%s', error: '%s'\n", fileName, err)
			continue
		}
		d = markdown.NormalizeNewlines(d)

		exts := parser.CommonExtensions // parser.OrderedListStart | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(exts)
		doc := markdown.Parse(d, p)

		if flgToHTML {
			htmlFlags := mdhtml.Smartypants |
				mdhtml.SmartypantsFractions |
				mdhtml.SmartypantsDashes |
				mdhtml.SmartypantsLatexDashes
			htmlOpts := mdhtml.RendererOptions{
				Flags: htmlFlags,
			}
			renderer := mdhtml.NewRenderer(htmlOpts)
			html := markdown.Render(doc, renderer)
			fmt.Printf("HTML of file '%s':\n%s\n", fileName, string(html))

		} else {
			fmt.Printf("Ast of file '%s':\n", fileName)
			ast.PrintWithPrefix(os.Stdout, doc, " ")
			fmt.Print("\n")
		}
	}
}

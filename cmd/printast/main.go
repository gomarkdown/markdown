package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// This prints AST of parsed markdown document.
// Usage: printast <markdown-file>

func usageAndExit() {
	fmt.Printf("Usage: printast <markdown-file>\n")
	os.Exit(1)
}

func main() {
	nFiles := len(os.Args) - 1
	if nFiles < 1 {
		usageAndExit()
	}
	for i := 0; i < nFiles; i++ {
		fileName := os.Args[i+1]
		d, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open '%s', error: '%s'\n", fileName, err)
			continue
		}
		exts := parser.CommonExtensions // parser.OrderedListStart | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(exts)
		doc := markdown.Parse(d, p)
		fmt.Printf("Ast of file '%s':\n", fileName)
		ast.PrintWithPrefix(os.Stdout, doc, " ")
		fmt.Print("\n")
	}
}

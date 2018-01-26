package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
)

/*
This checks that markdown parser doesn't crash by parsing every substring
of markdown string coming from test files.
*/

var files = []string{
	"Amps and angle encoding",
	"Auto links",
	"Backslash escapes",
	"Blockquotes with code blocks",
	"Code Blocks",
	"Code Spans",
	"Hard-wrapped paragraphs with list-like lines no empty line before block",
	"Horizontal rules",
	"Inline HTML (Advanced)",
	"Inline HTML (Simple)",
	"Inline HTML comments",
	"Links, inline style",
	"Links, reference style",
	"Links, shortcut references",
	"Literal quotes in titles",
	"Markdown Documentation - Basics",
	"Markdown Documentation - Syntax",
	"Nested blockquotes",
	"Ordered and unordered lists",
	"Strong and em together",
	"Tabs",
	"Tidyness",
}

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

var (
	wg sync.WaitGroup
	// semaphore to control concurrency level
	sem chan bool
)

func runMarkdown(d []byte) string {
	html := markdown.ToHTML(d, nil, nil)
	return string(html)
}

func testFile(basename string) {
	filename := filepath.Join("testdata", basename+".text")
	input, err := ioutil.ReadFile(filename)
	panicIfErr(err)

	// test every prefix of every input to check for bounds checking
	n := len(input)
	for end := 1; end < n; end++ {
		d := input[0:end]
		sem <- true
		wg.Add(1)
		go func(md []byte) {
			go runMarkdown(md)
			wg.Done()
			<-sem
		}(d)
	}
}

func main() {
	// set concurrency level equal to number of processors
	nProcs := runtime.NumCPU()
	sem = make(chan bool, nProcs)

	fmt.Printf("Running crash tests using %d processors\n", nProcs)

	timeStart := time.Now()
	for _, basename := range files {
		testFile(basename)
	}
	wg.Wait()
	fmt.Printf("Success! We didn't crash!\nTests took %s\n", time.Since(timeStart))
}

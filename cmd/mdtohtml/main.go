package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const defaultTitle = ""

func main() {
	var page, toc, xhtml, latex, smartypants, latexdashes, fractions bool
	var css, cpuprofile string
	var repeat int
	flag.BoolVar(&page, "page", false,
		"Generate a standalone HTML page (implies -latex=false)")
	flag.BoolVar(&toc, "toc", false,
		"Generate a table of contents (implies -latex=false)")
	flag.BoolVar(&xhtml, "xhtml", true,
		"Use XHTML-style tags in HTML output")
	//flag.BoolVar(&latex, "latex", false,
	//	"Generate LaTeX output instead of HTML")
	flag.BoolVar(&smartypants, "smartypants", true,
		"Apply smartypants-style substitutions")
	flag.BoolVar(&latexdashes, "latexdashes", true,
		"Use LaTeX-style dash rules for smartypants")
	flag.BoolVar(&fractions, "fractions", true,
		"Use improved fraction rules for smartypants")
	flag.StringVar(&css, "css", "",
		"Link to a CSS stylesheet (implies -page)")
	flag.StringVar(&cpuprofile, "cpuprofile", "",
		"Write cpu profile to a file")
	flag.IntVar(&repeat, "repeat", 1,
		"Process the input multiple times (for benchmarking)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Markdown Processor "+
			"\nAvailable at http://github.com/gomarkdown/markdown/cmd/mdtohtml\n\n"+
			"Copyright © 2011 Russ Ross <russ@russross.com>\n"+
			"Copyright © 2018 Krzysztof Kowalczyk <https://blog.kowalczyk.info>\n"+
			"Distributed under the Simplified BSD License\n"+
			"Usage:\n"+
			"  %s [options] [inputfile [outputfile]]\n\n"+
			"Options:\n",
			os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// enforce implied options
	if css != "" {
		page = true
	}
	if page {
		latex = false
	}
	if toc {
		latex = false
	}

	// turn on profiling?
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// read the input
	var input []byte
	var err error
	args := flag.Args()
	switch len(args) {
	case 0:
		if input, err = ioutil.ReadAll(os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from Stdin:", err)
			os.Exit(-1)
		}
	case 1, 2:
		if input, err = ioutil.ReadFile(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from", args[0], ":", err)
			os.Exit(-1)
		}
	default:
		flag.Usage()
		os.Exit(-1)
	}

	// set up options
	var extensions = parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings

	var renderer markdown.Renderer
	if latex {
		// render the data into LaTeX
		//renderer = markdown.LatexRenderer(0)
	} else {
		// render the data into HTML
		var htmlFlags html.Flags
		if xhtml {
			htmlFlags |= html.UseXHTML
		}
		if smartypants {
			htmlFlags |= html.Smartypants
		}
		if fractions {
			htmlFlags |= html.SmartypantsFractions
		}
		if latexdashes {
			htmlFlags |= html.SmartypantsLatexDashes
		}
		title := ""
		if page {
			htmlFlags |= html.CompletePage
			title = getTitle(input)
		}
		if toc {
			htmlFlags |= html.TOC
		}
		params := html.RendererOptions{
			Flags: htmlFlags,
			Title: title,
			CSS:   css,
		}
		renderer = html.NewRenderer(params)
	}

	// parse and render
	var output []byte
	for i := 0; i < repeat; i++ {
		parser := parser.NewWithExtensions(extensions)
		output = markdown.ToHTML(input, parser, renderer)
	}

	// output the result
	var out *os.File
	if len(args) == 2 {
		if out, err = os.Create(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v", args[1], err)
			os.Exit(-1)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	if _, err = out.Write(output); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing output:", err)
		os.Exit(-1)
	}
}

// try to guess the title from the input buffer
// just check if it starts with an <h1> element and use that
func getTitle(input []byte) string {
	i := 0

	// skip blank lines
	for i < len(input) && (input[i] == '\n' || input[i] == '\r') {
		i++
	}
	if i >= len(input) {
		return defaultTitle
	}
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}

	// find the first line
	start := i
	for i < len(input) && input[i] != '\n' && input[i] != '\r' {
		i++
	}
	line1 := input[start:i]
	if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
		i++
	}
	i++

	// check for a prefix header
	if len(line1) >= 3 && line1[0] == '#' && (line1[1] == ' ' || line1[1] == '\t') {
		return strings.TrimSpace(string(line1[2:]))
	}

	// check for an underlined header
	if i >= len(input) || input[i] != '=' {
		return defaultTitle
	}
	for i < len(input) && input[i] == '=' {
		i++
	}
	for i < len(input) && (input[i] == ' ' || input[i] == '\t') {
		i++
	}
	if i >= len(input) || (input[i] != '\n' && input[i] != '\r') {
		return defaultTitle
	}

	return strings.TrimSpace(string(line1))
}

// Package markdown implements markdown parser and HTML renderer.
//
// It parses markdown into AST format which can be serialized to HTML
// (using HTMLRenderer) or possibly other formats (using alternate renderers).
//
// The simplest way to connvert markdown text into HTML is to call
// ToHTML(input, nil, nil) function with nil parser and renderer.
//
// To have more control over parsing, create parser using
// NewParserWithExtensions(extensions) and call ToHTML(input, parser, nil)
//
// To have more control over rendering, create renderer using NewHTMLRenderer(params)
// and call ToHTML(input, nil, renderer).
//
// For a cmd-line tool, use cmd/mdtohtml
package markdown

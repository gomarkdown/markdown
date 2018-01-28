// Package markdown implements markdown parser and HTML renderer.
//
// It parses markdown into AST format which can be serialized to HTML
// (using html.Renderer) or possibly other formats (using alternate renderers).
//
//
// Convert markdown to HTML
//
// The simplest way to convert markdown document to HTML
//
//  md := []byte("## markdown document")
//  html := markdown.ToHTML(md, nil, nil)
//
// Customizing parsing and HTML rendering
//
// You can customize parser and HTML renderer:
//
//  md := []byte("markdown document")
//  extensions := parser.CommonExtensions | parser.AutoHeadingIDs
//  parser := parser.NewWithExensions(extensions)
//  htmlParams := html.CommonFlags | html.HrefTargetBlank
//  renderer := html.NewRenderer(htmlParams)
//  html := markdown.ToHTML(md, parser, renderer)
//
// For a cmd-line tool see https://github.com/gomarkdown/markdown/tree/master/cmd/mdtohtml
package markdown

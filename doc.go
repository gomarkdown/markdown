// Package markdown implements markdown parser and HTML renderer.
//
// It parses markdown into AST format which can be serialized to HTML
// (using HTMLRenderer) or possibly other formats (using alternate renderers).
//
//
// Convert markdown to HTML
//
//  md := []byte("## markdown document")
//  html := ToHTML(md, nil, nil)
//
// Customizing parsing and HTML rendering
//
// md := []byte("markdown document")
// extensions := CommonExtensions | AutoHeadingIDs
// parser := NewParserWithExensions(extensions)
// htmlParams := CommonHTMLFlags | HrefTargetBlank
// renderer := NewHTMLRenderer(htmlParams)
// html := ToHTML(md, parser, renderer)
//
// For a cmd-line tool see https://github.com/gomarkdown/markdown/tree/master/cmd/mdtohtml
package markdown

// Package matter names the document matters (front, main and back), so the
// parser and the renderers read and write the same tags.
package matter

import "github.com/gomarkdown/markdown/ast"

var names = [...]struct{ name, tag string }{
	ast.DocumentMatterFront: {"front", "{frontmatter}"},
	ast.DocumentMatterMain:  {"main", "{mainmatter}"},
	ast.DocumentMatterBack:  {"back", "{backmatter}"},
}

// Kinds lists the document matters that have a name.
var Kinds = [...]ast.DocumentMatters{ast.DocumentMatterFront, ast.DocumentMatterMain, ast.DocumentMatterBack}

// Name returns "front", "main" or "back", or "" for any other value.
func Name(m ast.DocumentMatters) string {
	if m < 0 || int(m) >= len(names) {
		return ""
	}
	return names[m].name
}

// Tag returns the markdown tag that starts a matter, like "{frontmatter}",
// or "" for any other value.
func Tag(m ast.DocumentMatters) string {
	if m < 0 || int(m) >= len(names) {
		return ""
	}
	return names[m].tag
}

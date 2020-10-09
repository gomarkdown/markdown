[![Go Doc][godoc-image]][godoc-url]
[![Build Status][workflow-image]][workflow-url]

# Markdown

This is a fork of https://github.com/gomarkdown/markdown

This package is a very fast Go library for parsing Markdown documents and rendering them to HTML.
It's fast and supports common extensions.

## License

[Simplified BSD License](./LICENSE)

## Usage

To convert markdown text to HTML using reasonable defaults:

```go
md := []byte("## markdown document")
output := markdown.ToHTML(md, nil, nil)
```

### Customizing Markdown Parser

Markdown format is loosely specified and there are multiple extensions invented after original specification was created.

The parser supports several [extensions](https://godoc.org/github.com/moorara/markdown/parser#Extensions).

Default parser uses most common `parser.CommonExtensions` but you can easily use parser with custom extension:

```go
import (
  "github.com/moorara/markdown"
  "github.com/moorara/markdown/parser"
)

extensions := parser.CommonExtensions | parser.AutoHeadingIDs
parser := parser.NewWithExtensions(extensions)

md := []byte("markdown text")
html := markdown.ToHTML(md, parser, nil)
```

### Customizing HTML Renderer

Similarly, HTML renderer can be configured with different [options](https://godoc.org/github.com/moorara/markdown/html#RendererOptions)

Here's how to use a custom renderer:

```go
import (
  "github.com/moorara/markdown"
  "github.com/moorara/markdown/html"
)

htmlFlags := html.CommonFlags | html.HrefTargetBlank
opts := html.RendererOptions{Flags: htmlFlags}
renderer := html.NewRenderer(opts)

md := []byte("markdown text")
html := markdown.ToHTML(md, nil, renderer)
```


[godoc-url]: https://pkg.go.dev/github.com/moorara/markdown
[godoc-image]: https://godoc.org/github.com/moorara/markdown?status.svg
[workflow-url]: https://github.com/moorara/markdown/actions
[workflow-image]: https://github.com/moorara/markdown/workflows/Main/badge.svg

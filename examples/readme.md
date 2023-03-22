Here you can find examples of advanced uses of this library.

You can use them as base for your own code.

They are described in more detail in https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

You can run each of them with: `go run <program.go>`.

The examples:
* `basic.go` : simplest markdown => HTML example
* `render_hook.go` : shows how to customize HTML renderer with render hook function
* `code_highlight.go` : shows how to syntax highlight code blocks using `github.com/alecthomas/chroma`
* `parser_hook.go` : shows how to extend parser to recognize custom block-level syntax
* `modify_ast.go` : shows how to modify AST after parsing but before HTML rendering

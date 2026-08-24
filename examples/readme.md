Here you can find examples of advanced uses of this library.

You can use them as a base for your own code.

They are described in more detail in https://blog.kowalczyk.info/article/cxn3/advanced-markdown-processing-in-go.html

You can run each of them with: `go run <program.go>`.

The examples:
* `basic.go` : simplest markdown => HTML example
* `inline.go` : extend the parser with custom inline syntax (wiki links, hashtags)
* `render_hook.go` : shows how to customize the HTML renderer with a render hook function
* `code_highlight.go` : shows how to syntax-highlight code blocks using `github.com/alecthomas/chroma`
* `parser_hook.go` : shows how to extend the parser to recognize custom block-level syntax
* `modify_ast.go` : shows how to modify the AST after parsing but before HTML rendering

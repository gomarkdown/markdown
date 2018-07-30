package parser

import "github.com/gomarkdown/markdown/ast"

// ParserOptions is a collection of supplementary parameters tweaking the behavior of various parts of the parser.
type ParserOptions struct {
	ParserHook BlockFunc
}

// BlockFunc allows to registration of a parser function. If successful it
// returns an ast.Node, a buffer that should be parsed as a block and the the number of bytes consumed.
type BlockFunc func(data []byte) (ast.Node, []byte, int)

package parser

import "github.com/gomarkdown/markdown/ast"

// ParserOptions is a collection of supplementary parameters tweaking the behavior of various parts of the parser.
type ParserOptions struct {
	ParserHook    BlockFunc
	ReadIncludeFn ReadIncludeFunc
}

// BlockFunc allows to registration of a parser function. If successful it
// returns an ast.Node, a buffer that should be parsed as a block and the the number of bytes consumed.
type BlockFunc func(data []byte) (ast.Node, []byte, int)

// ReadIncludeFunc reads the file under path and returns the read bytes. If path is not absolute it is taken
// relative to the currently parsed file. If this not set no data will be read from the filesystem.
// address is the optional address specifier of which lines of the file to return.
type ReadIncludeFunc func(path string, address []byte) []byte

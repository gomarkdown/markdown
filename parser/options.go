package parser

// ParserOptions is a collection of supplementary parameters tweaking
// the behavior of various parts of parser.
type ParserOptions struct {
	ParserHook BlockFunc
}

// BlockFunc allows to registration of a parser function. If successful it
// returns the number of bytes consumed.
type BlockFunc func(p *Parser, data []byte) int

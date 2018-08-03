package parser

import (
	"path"
	"path/filepath"
)

// updateWd updates the working directory. If new is an absolute
// path we just use that, relative paths are taken relative of cur.
func updateWd(cur, new string) string {
	if path.IsAbs(new) {
		return path.Dir(new)
	}

	return path.Dir(filepath.Join(cur, new))
}

// {{...}}
func (p *Parser) isInclude(data []byte) (filename string, address []byte, consumed int) {
	i := 0
	if len(data) < 3 {
		return "", nil, 0
	}
	if data[i] != '{' || data[i+1] != '{' {
		return "", nil, 0
	}

	// find the end delimiter
	i = skipUntilChar(data, i, '}')
	if i+1 >= len(data) {
		return "", nil, 0
	}
	if data[i+1] != '}' {
		return "", nil, 0
	}
	return string(data[2:i]), nil, i + 1
}

func (p *Parser) include(file string, address []byte) []byte {
	// p.cwd holds containing dir, already tailored to the path, means path.Base should give us the leaf.
	fullPath := filepath.Join(p.cwd, path.Base(file))
	if p.Opts.IncludeHook != nil {
		return p.Opts.IncludeHook(fullPath, address)
	}

	return nil
}

// <{{...}}
func (p *Parser) isCodeInclude(data []byte) (filename string, address []byte, consumed int) {
	if len(data) < 3 {
		return "", nil, 0
	}
	if data[0] != '<' {
		return "", nil, 0
	}
	return p.isInclude(data[1:])
}

// codeInclude acts like include except the returned bytes are wrapped in a fenced code block.
func (p *Parser) codeInclude(file string, address []byte) []byte {
	data := p.include(file, address)
	if data == nil {
		return nil
	}
	// possible some fiddling to set the language etc.
	data = append([]byte("```\n"), data...)
	data = append(data, []byte("```")...)
	return data
}

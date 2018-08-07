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

// isInclude parses {{...}}[...], that contains a path between the {{, the [...] syntax contains
// an address to select which lines to include. It is treated as an opaque string and just given
// to readInclude.
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
	end := i
	i++
	if data[i] != '}' {
		return "", nil, 0
	}

	if i+1 < len(data) && data[i+1] == '[' { // potential address specification
		i++
		start := i + 1

		i = skipUntilChar(data, start, ']')
		if i >= len(data) {
			return "", nil, 0
		}
		address = data[start:i]
	}
	return string(data[2:end]), address, i + 1
}

func (p *Parser) readInclude(file string, address []byte) []byte {
	// p.cwd holds containing dir, already tailored to the path, means path.Base should give us the leaf.
	fullPath := filepath.Join(p.cwd, path.Base(file))
	if p.Opts.ReadIncludeFn != nil {
		return p.Opts.ReadIncludeFn(fullPath, address)
	}

	return nil
}

// isCodeInclude parses <{{...}} which is similar to isInclude the returned bytes are, however wrapped in a code block.
func (p *Parser) isCodeInclude(data []byte) (filename string, address []byte, consumed int) {
	if len(data) < 3 {
		return "", nil, 0
	}
	if data[0] != '<' {
		return "", nil, 0
	}
	return p.isInclude(data[1:])
}

// readCodeInclude acts like include except the returned bytes are wrapped in a fenced code block.
func (p *Parser) readCodeInclude(file string, address []byte) []byte {
	data := p.readInclude(file, address)
	if data == nil {
		return nil
	}
	// possible some fiddling to set the language etc.
	data = append([]byte("```\n"), data...)
	data = append(data, []byte("```")...)
	return data
}

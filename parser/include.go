package parser

import (
	"io/ioutil"
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
func (p *Parser) isInclude(data []byte) (filename string, consumed int) {
	i := 0
	if len(data) < 3 {
		return "", 0
	}
	if data[i] != '{' || data[i+1] != '{' {
		return "", 0
	}

	i = skipUntilChar(data, i, '}')
	// find the end delimiter
	end, j := 0, 0
	for end = i; end < len(data) && j < 2; end++ {
		if data[end] == '}' {
			j++
		} else {
			j = 0
		}
	}
	if j < 2 && end >= len(data) {
		return "", 0
	}
	return string(data[i+2 : end-2]), end
}

func (p *Parser) include(file string) ([]byte, error) {
	// p.cwd holds containing dir, already tailed to path
	data, err := ioutil.ReadFile(filepath.Join(p.cwd, path.Base(file)))

	// TODO: address, prefix and stuff, becomes more full fledged function latter.

	return data, err
}

// <{{...}}
func (p *Parser) isCodeInclude(data []byte) (filename string, consumed int) {
	if len(data) < 3 {
		return "", 0
	}
	if data[0] != '<' {
		return "", 0
	}
	return p.isInclude(data[1:])
}

// codeInclude acts like include except the returns bytes are wrapped in a fenced code block.
func (p *Parser) codeInclude(file string) ([]byte, error) {
	data, err := p.include(file)
	// possible some fiddling to set the language etc.
	data = append([]byte("```\n"), data...)
	data = append(data, []byte("```")...)
	return data, err
}

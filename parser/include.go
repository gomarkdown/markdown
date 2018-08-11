package parser

import (
	"bytes"
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
	filename = string(data[2:end])

	if i+1 < len(data) && data[i+1] == '[' { // potential address specification
		start := i + 2

		end = skipUntilChar(data, start, ']')
		if end >= len(data) {
			return "", nil, 0
		}
		address = data[start:end]
		return filename, address, end + 1
	}

	return filename, address, i + 1
}

func (p *Parser) readInclude(file string, address []byte) []byte {
	if p.Opts.ReadIncludeFn != nil {
		return p.Opts.ReadIncludeFn(file, address)
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
	filename, address, consumed = p.isInclude(data[1:])
	if consumed == 0 {
		return "", nil, 0
	}
	return filename, address, consumed + 1
}

// readCodeInclude acts like include except the returned bytes are wrapped in a fenced code block.
func (p *Parser) readCodeInclude(file string, address []byte) []byte {
	data := p.readInclude(file, address)
	if data == nil {
		return nil
	}
	ext := path.Ext(file)
	buf := &bytes.Buffer{}
	buf.Write([]byte("```"))
	if ext != "" {
		buf.WriteString(" " + ext[1:] + "\n")
	} else {
		buf.WriteByte('\n')
	}
	buf.Write(data)
	buf.WriteString("```\n")
	return buf.Bytes()
}

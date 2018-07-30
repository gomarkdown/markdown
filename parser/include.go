package parser

// linesUntilEmpty scans up to the first empty line.
func (p *Parser) linesUntilEmpty(data []byte) int {
	line, i := 0, 0

	for line < len(data) {
		i++

		// find the end of this line
		for i < len(data) && data[i-1] != '\n' {
			i++
		}

		if p.isEmpty(data[line:i]) == 0 {
			line = i
			continue
		}

		break
	}
	return i
}

package parser

// IsEmpty returns the length of a blank line at the start of data,
// or 0 if the line is not blank. It is okay to call on an empty buffer.
func IsEmpty(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	var i int
	for i = 0; i < len(data) && data[i] != '\n'; i++ {
		if data[i] != ' ' && data[i] != '\t' {
			return 0
		}
	}
	i = skipCharN(data, i, '\n', 1)
	return i
}

// skipChar advances i as long as data[i] == c
func skipChar(data []byte, i int, c byte) int {
	n := len(data)
	for i < n && data[i] == c {
		i++
	}
	return i
}

// like skipChar but only skips up to max characters
func skipCharN(data []byte, i int, c byte, max int) int {
	n := len(data)
	for i < n && max > 0 && data[i] == c {
		i++
		max--
	}
	return i
}

// skipUntilChar advances i as long as data[i] != c
func skipUntilChar(data []byte, i int, c byte) int {
	n := len(data)
	for i < n && data[i] != c {
		i++
	}
	return i
}

func skipAlnum(data []byte, i int) int {
	n := len(data)
	for i < n && IsAlnum(data[i]) {
		i++
	}
	return i
}

func skipSpace(data []byte, i int) int {
	n := len(data)
	for i < n && IsSpace(data[i]) {
		i++
	}
	return i
}

func backChar(data []byte, i int, c byte) int {
	for i > 0 && data[i-1] == c {
		i--
	}
	return i
}

func backUntilChar(data []byte, i int, c byte) int {
	for i > 0 && data[i-1] != c {
		i--
	}
	return i
}

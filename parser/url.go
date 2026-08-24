package parser

import "bytes"

var URIs = [][]byte{
	[]byte("http://"),
	[]byte("https://"),
	[]byte("ftp://"),
	[]byte("mailto:"),
}

var Paths = [][]byte{
	[]byte("/"),
	[]byte("./"),
	[]byte("../"),
}

// IsSafeURL returns true if url starts with one of the valid schemes or is a relative path.
func IsSafeURL(url []byte) bool {
	nLink := len(url)
	for _, path := range Paths {
		nPath := len(path)
		// Length must be checked before slicing; short urls (empty, ".", "..")
		// are shorter than some path prefixes and would panic on url[:nPath].
		// See GHSA-cv23-7vc5-jfh7.
		if nLink < nPath {
			continue
		}
		if bytes.Equal(url[:nPath], path) {
			if nLink == nPath {
				return true
			} else if IsAlnum(url[nPath]) {
				return true
			}
		}
	}

	for _, prefix := range URIs {
		// TODO: handle unicode here
		// case-insensitive prefix test
		nPrefix := len(prefix)
		if nLink > nPrefix {
			linkPrefix := bytes.ToLower(url[:nPrefix])
			if bytes.Equal(linkPrefix, prefix) && IsAlnum(url[nPrefix]) {
				return true
			}
		}
	}

	return false
}

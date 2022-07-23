package utils

import "github.com/gomarkdown/markdown/ast"

// IsAlnum returns true if c is a digit or letter
// TODO: check when this is looking for ASCII alnum and when it should use unicode
func IsAlnum(c byte) bool {
	return IsDigit(c) || IsLetter(c)
}

// IsDigit returns true if c is a digit
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// IsSpace returns true if c is a white-space charactr
func IsSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

// IsLetter returns true if c is ascii letter
func IsLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// IsPunctuation returns true if c is a punctuation symbol.
func IsPunctuation(c byte) bool {
	for _, r := range []byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~") {
		if c == r {
			return true
		}
	}
	return false
}

func IsListItem(node ast.Node) bool {
	_, ok := node.(*ast.ListItem)
	return ok
}

func IsListItemTerm(node ast.Node) bool {
	data, ok := node.(*ast.ListItem)
	return ok && data.ListFlags&ast.ListTypeTerm != 0
}

func IsList(node ast.Node) bool {
	_, ok := node.(*ast.List)
	return ok
}

func IsListTight(node ast.Node) bool {
	if list, ok := node.(*ast.List); ok {
		return list.Tight
	}
	return false
}

package markdown

import (
	"bytes"
	"testing"
)

func TestDocument(t *testing.T) {
	var tests = []string{
		// Empty document.
		"",
		"",

		" ",
		"",

		// This shouldn't panic.
		// https://github.com/russross/blackfriday/issues/172
		"[]:<",
		"<p>[]:&lt;</p>\n",

		// This shouldn't panic.
		// https://github.com/russross/blackfriday/issues/173
		"   [",
		"<p>[</p>\n",
	}
	doTests(t, tests)
}

func TestNormalizeNewlines(t *testing.T) {
	var tests = [][]byte{
		{},
		{10},
		{13},
		{13, 10, 13},
		{'a', 13, 10, 13, 'b'},
		{'a', 13, 10},
		{'b', 13},
		{13, 10, 13, 10},
		{13, 10, 13, 10, 'a'},
		{13, 10, 13, 10, 10},
		{13, 10, 13, 10, 13},
	}
	ref := func(d []byte) []byte {
		d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1) // CRLF => LF
		d = bytes.Replace(d, []byte{13}, []byte{10}, -1)     // CR => LF
		return d
	}
	dup := func(d []byte) []byte {
		return append([]byte{}, d...)
	}
	for i := 0; i < len(tests); i++ {
		d := tests[i]
		got := NormalizeNewlines(dup(d))
		exp := ref(dup(d))
		if !bytes.Equal(got, exp) {
			t.Errorf("got: %v, exp: %v, i: %d\n", got, exp, i)
		}

	}
}

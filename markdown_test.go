package markdown

import "testing"

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

func TestLineEndings(t *testing.T) {
	var tests = []string{
		// https://github.com/gomarkdown/markdown/issues/154
		"\r\nsomething",
		"<p>something</p>\n",

		"\nsomething else",
		"<p>something else</p>\n",

		"something\r\n",
		"<p>something</p>\n",

		"something else\n",
		"<p>something else</p>\n",

		"something else",
		"<p>something else</p>\n",
	}
	doTests(t, tests)
}

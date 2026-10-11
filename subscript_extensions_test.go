package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestSuperSubscriptExtensions(t *testing.T) {
	cases := []struct {
		name        string
		extensions  parser.Extensions
		input, want string
	}{
		{"disabled", parser.NoExtensions, "H~2~O and 2^10^\n", "<p>H~2~O and 2^10^</p>\n"},
		{"alone", parser.SuperSubscript, "H~2~O and 2^10^\n", "<p>H<sub>2</sub>O and 2<sup>10</sup></p>\n"},
		{"with common", parser.CommonExtensions | parser.SuperSubscript, "H~2~O and 2^10^\n", "<p>H<sub>2</sub>O and 2<sup>10</sup></p>\n"},
		{"strikethrough disabled", parser.SuperSubscript, "~~deleted~~\n", "<p>~~deleted~~</p>\n"},
		{"strikethrough enabled", parser.SuperSubscript | parser.Strikethrough, "~~deleted~~ and H~2~O\n", "<p><del>deleted</del> and H<sub>2</sub>O</p>\n"},
		{"unmatched subscript", parser.SuperSubscript, "H~2O\n", "<p>H~2O</p>\n"},
		{"escaped subscript", parser.SuperSubscript, "H\\~2\\~O\n", "<p>H~2~O</p>\n"},
		{"heading", parser.SuperSubscript, "# H~2~O\n", "<h1>H<sub>2</sub>O</h1>\n"},
		{"link", parser.SuperSubscript, "[H~2~O](u)\n", "<p><a href=\"u\">H<sub>2</sub>O</a></p>\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runMarkdown(tc.input, TestParams{extensions: tc.extensions})
			if got != tc.want {
				t.Errorf("input %q: got %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

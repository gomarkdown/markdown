package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestSingleTildeDoesNotEmphasize(t *testing.T) {
	configs := []struct {
		name       string
		extensions parser.Extensions
	}{
		{"strikethrough", parser.Strikethrough},
		{"common", parser.CommonExtensions},
	}
	cases := []struct {
		name, input, want string
	}{
		{"single", "a ~word~ b\n", "<p>a ~word~ b</p>\n"},
		{"mixed", "~literal~ and ~~deleted~~\n", "<p>~literal~ and <del>deleted</del></p>\n"},
		{"inside strikethrough", "~~deleted ~literal~ text~~\n", "<p><del>deleted ~literal~ text</del></p>\n"},
		{"other emphasis", "~literal~ *italic* **bold**\n", "<p>~literal~ <em>italic</em> <strong>bold</strong></p>\n"},
	}
	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					got := runMarkdown(tc.input, TestParams{extensions: config.extensions})
					if got != tc.want {
						t.Errorf("input %q: got %q, want %q", tc.input, got, tc.want)
					}
				})
			}
		})
	}
}

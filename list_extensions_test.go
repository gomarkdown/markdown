package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestDefinitionListSyntaxRequiresExtensionInLists(t *testing.T) {
	configs := []struct {
		name       string
		extensions parser.Extensions
	}{
		{"none", parser.NoExtensions},
		{"fenced code", parser.FencedCode},
		{"common without definitions", parser.CommonExtensions &^ parser.DefinitionLists},
	}
	cases := []struct {
		name, input, want string
	}{
		{"unordered continuation", "* item\n: literal\n* next\n", "<ul>\n<li>item\n: literal</li>\n<li>next</li>\n</ul>\n"},
		{"unordered nested", "* item\n  term\n  : literal\n* next\n", "<ul>\n<li>item\nterm\n: literal</li>\n<li>next</li>\n</ul>\n"},
		{"ordered continuation", "1. item\n: literal\n2. next\n", "<ol>\n<li>item\n: literal</li>\n<li>next</li>\n</ol>\n"},
		{"ordered nested", "1. item\n  term\n  : literal\n2. next\n", "<ol>\n<li>item\nterm\n: literal</li>\n<li>next</li>\n</ol>\n"},
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

package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestTableCaptionRequiresMmark(t *testing.T) {
	const tableInput = "a | b\n--- | ---\nx | y\n"
	const tableHTML = "<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td>x</td>\n<td>y</td>\n</tr>\n</tbody>\n</table>\n"
	configs := []struct {
		name       string
		extensions parser.Extensions
	}{
		{"tables", parser.Tables},
		{"common", parser.CommonExtensions},
		{"tables with heading IDs", parser.Tables | parser.HeadingIDs},
	}
	cases := []struct {
		name, caption, plain, special string
	}{
		{"text", "Table: caption\n", "<p>Table: caption</p>\n", "<figure>\n" + tableHTML + "<figcaption>caption\n</figcaption>\n</figure>\n"},
		{"ID and inline markup", "Table: *caption* {#table-id}\n", "<p>Table: <em>caption</em> {#table-id}</p>\n", "<figure id=\"table-id\">\n" + tableHTML + "<figcaption><em>caption</em> </figcaption>\n</figure>\n"},
	}
	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			for _, tc := range cases {
				for _, mmark := range []bool{false, true} {
					name := tc.name + "/disabled"
					extensions := config.extensions
					want := tableHTML + tc.plain
					if mmark {
						name = tc.name + "/enabled"
						extensions |= parser.Mmark
						want = tc.special
					}
					t.Run(name, func(t *testing.T) {
						input := tableInput + tc.caption + "\nAfter\n"
						want += "\n<p>After</p>\n"
						got := runMarkdown(input, TestParams{extensions: extensions})
						if got != want {
							t.Errorf("input %q: got %q, want %q", input, got, want)
						}
					})
				}
			}
		})
	}
}

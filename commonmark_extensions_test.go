package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func commonMarkExtHTML(input string, extensions parser.Extensions) string {
	return string(ToHTML([]byte(input), parser.NewCommonMarkWithExtensions(extensions), nil))
}

func TestCommonMarkExtensions(t *testing.T) {
	all := parser.CommonMarkExtensions
	tests := []struct {
		name, input, want string
		extensions        parser.Extensions
	}{
		{"table", "| a | b |\n|:--|--:|\n| 1 | *2* |\n", "<table>\n<thead>\n<tr>\n<th align=\"left\">a</th>\n<th align=\"right\">b</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td align=\"left\">1</td>\n<td align=\"right\"><em>2</em></td>\n</tr>\n</tbody>\n</table>\n", all},
		{"table interrupts paragraph", "text\na | b\n--|--\n1 | 2\n", "<p>text</p>\n\n<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n", all},
		{"table off", "a | b\n--|--\n", "<p>a | b\n--|--</p>\n", parser.NoExtensions},
		{"table in list", "- a | b\n  --|--\n  1 | 2\n", "<ul>\n<li>\n\n<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n</li>\n</ul>\n", all},
		{"strikethrough", "a ~~b~~ c\n", "<p>a <del>b</del> c</p>\n", all},
		{"strikethrough with emphasis", "~~*b*~~ **~~c~~**\n", "<p><del><em>b</em></del> <strong><del>c</del></strong></p>\n", all},
		{"single tilde", "~5 MB to ~10 MB ~a~\n", "<p>~5 MB to ~10 MB ~a~</p>\n", all},
		{"escaped tildes", "\\~~a~~\n", "<p>~~a~~</p>\n", all},
		{"strikethrough off", "~~b~~\n", "<p>~~b~~</p>\n", parser.NoExtensions},
		{"autolink", "see https://example.com/a_b_c.\n", "<p>see <a href=\"https://example.com/a_b_c\">https://example.com/a_b_c</a>.</p>\n", all},
		{"autolink parens", "(see https://en.wikipedia.org/wiki/A_(b))\n", "<p>(see <a href=\"https://en.wikipedia.org/wiki/A_(b)\">https://en.wikipedia.org/wiki/A_(b)</a>)</p>\n", all},
		{"autolink emphasis", "*http://a.com*\n", "<p><em><a href=\"http://a.com\">http://a.com</a></em></p>\n", all},
		{"autolink entity", "http://a.com/?x&amp;\n", "<p><a href=\"http://a.com/?x\">http://a.com/?x</a>&amp;</p>\n", all},
		{"autolink in link", "[go to http://a.com](http://b.com)\n", "<p><a href=\"http://b.com\">go to http://a.com</a></p>\n", all},
		{"autolink in anchor", "<a href=\"http://b.com\">http://a.com</a>\n", "<p><a href=\"http://b.com\">http://a.com</a></p>\n", all},
		{"autolink in code", "`http://a.com`\n", "<p><code>http://a.com</code></p>\n", all},
		{"autolink mid word", "xhttp://a.com\n", "<p>xhttp://a.com</p>\n", all},
		{"autolink scheme only", "http:// and mailto:\n", "<p>http:// and mailto:</p>\n", all},
		{"angle autolink", "<http://a.com>\n", "<p><a href=\"http://a.com\">http://a.com</a></p>\n", all},
		{"autolink off", "http://a.com\n", "<p>http://a.com</p>\n", parser.NoExtensions},
		{"heading ids", "# Hello World\n\nSetext\n---\n\n# Hello World\n", "<h1 id=\"hello-world\">Hello World</h1>\n<h2 id=\"setext\">Setext</h2>\n<h1 id=\"hello-world-1\">Hello World</h1>\n", all},
		{"heading ids off", "# Hello\n", "<h1>Hello</h1>\n", parser.NoExtensions},
		{"unsupported extensions ignored", "Term\n: definition\n", "<p>Term\n: definition</p>\n", parser.CommonExtensions},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := commonMarkExtHTML(tc.input, tc.extensions); got != tc.want {
				t.Errorf("markdown: %q\nwant: %q\n got: %q", tc.input, tc.want, got)
			}
		})
	}
}

type testDirective struct {
	ast.Container
}

func TestCommonMarkParserHook(t *testing.T) {
	marker := []byte(":box\n")
	hook := func(data []byte) (ast.Node, []byte, int) {
		if !bytes.HasPrefix(data, marker) {
			return nil, nil, 0
		}
		end := bytes.Index(data[len(marker):], marker)
		if end < 0 {
			return nil, nil, 0
		}
		inner := data[len(marker) : len(marker)+end]
		return &testDirective{}, inner, 2*len(marker) + end
	}
	input := ":box\n# Title\n\n- *item*\n:box\n\n- a\n\n  :box\n  b\n  :box\n"
	for _, withHook := range []bool{true, false} {
		p := parser.NewCommonMarkWithExtensions(parser.NoExtensions)
		if !withHook {
			p = parser.NewCommonMark()
		}
		p.Opts.ParserHook = hook
		doc := p.Parse([]byte(input))
		var got []string
		ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
			if entering {
				if _, ok := node.(*testDirective); ok {
					var kids []string
					for _, c := range node.GetChildren() {
						kids = append(kids, typeName(c))
					}
					got = append(got, strings.Join(kids, ","))
				}
			}
			return ast.GoToNext
		})
		want := []string{"Heading,List", "Paragraph"}
		if !withHook {
			want = nil
		}
		if strings.Join(got, ";") != strings.Join(want, ";") {
			t.Errorf("hook %v: got %q, want %q", withHook, got, want)
		}
	}
}

func typeName(n ast.Node) string {
	switch n.(type) {
	case *ast.Heading:
		return "Heading"
	case *ast.List:
		return "List"
	case *ast.Paragraph:
		return "Paragraph"
	}
	return "other"
}

// Apart from heading ids, the extensions change only the CommonMark examples
// with bare URLs, which become links.
func TestCommonMarkExtensionsSpec(t *testing.T) {
	bareURLs := map[int]bool{602: true, 608: true, 611: true}
	for _, example := range commonMarkExamples(t) {
		got := commonMarkExtHTML(example.Markdown, parser.CommonMarkExtensions&^parser.AutoHeadingIDs)
		if (got != example.HTML) != bareURLs[example.Example] {
			t.Errorf("%d: %q\nwant: %q\n got: %q", example.Example, example.Markdown, example.HTML, got)
		}
	}
}

func TestCommonMarkExtensionsAdversarialInputs(t *testing.T) {
	for name, input := range map[string]string{
		"tildes":                 strings.Repeat("~~a ", 16384),
		"nested strikethrough":   strings.Repeat("~~", 16384) + "a" + strings.Repeat("~~", 16384),
		"closing parens":         "http://a" + strings.Repeat(")", 32768),
		"opening parens":         "http://a" + strings.Repeat("(", 32768),
		"semicolons":             "http://a" + strings.Repeat("&b;", 16384),
		"many urls":              strings.Repeat("http://a ", 16384),
		"urls in brackets":       strings.Repeat("[http://a", 16384),
		"anchors":                strings.Repeat("<a>http://a ", 8192),
		"pipes":                  strings.Repeat("a|b\n", 16384),
		"table rows":             "a|b\n-|-\n" + strings.Repeat("|", 32768),
		"backticks in pipe rows": strings.Repeat("|`", 16384) + "\n-|-\n",
		"headings":               strings.Repeat("# a\n", 16384),
	} {
		t.Run(name, func(t *testing.T) {
			parseWithParserShortTimeout(t, input, parser.NewCommonMarkWithExtensions(parser.CommonMarkExtensions))
		})
	}
}

package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// inlineShape describes the inline nodes of a document, quoting text
// literals exactly (ast.Print trims their spaces).
func inlineShape(doc ast.Node) string {
	var parts []string
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.Document, *ast.Paragraph:
		case *ast.Text:
			if entering {
				parts = append(parts, fmt.Sprintf("%q", n.Literal))
			}
		default:
			name := strings.TrimPrefix(fmt.Sprintf("%T", node), "*ast.")
			if !entering {
				name = "/" + name
			}
			parts = append(parts, name)
		}
		return ast.GoToNext
	})
	return strings.Join(parts, " ")
}

// TestCommonMarkInlineSpaces covers the spaces that end a text token in
// CommonMark mode: a single space between words stays in the token, while
// a space run, or a space before a tab or a line break, has its own rules.
func TestCommonMarkInlineSpaces(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a b c", `"a b c"`},
		{"a  b", `"a  b"`},
		{"a \tb", `"a \tb"`},
		{"a \nb", `"a\nb"`},
		{"a b\nc d\ne", `"a b\nc d\ne"`},
		{"a\n*b*", `"a\n" Emph "b" /Emph`},
		{"a  \nb", `"a" Hardbreak "b"`},
		{"a b *c* d", `"a b " Emph "c" /Emph " d"`},
		{"a *b *c", `"a *b *c"`},
		{"[a b ](u)", `Link "a b " /Link`},
	}
	for _, tc := range tests {
		got := inlineShape(NewCommonMark().Parse([]byte(tc.input)))
		if got != tc.want {
			t.Errorf("input %q\nExpected[%s]\nGot     [%s]", tc.input, tc.want, got)
		}
	}
}

func TestCommonMarkInlineSpaceBeforeURL(t *testing.T) {
	p := NewCommonMarkWithExtensions(Autolink)
	got := inlineShape(p.Parse([]byte("see http://example.com now")))
	want := `"see " Link "http://example.com" /Link " now"`
	if got != want {
		t.Errorf("\nExpected[%s]\nGot     [%s]", want, got)
	}
}

// A paragraph with no indented line uses the input bytes rather than a
// copy. Parsing must leave the input as it was either way.
func TestCommonMarkParagraphInput(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a b\nc *d*\n", `"a b\nc " Emph "d" /Emph`},
		{"a b\n  c\n d", `"a b\nc\nd"`},
		{"  a\nb  \n", `"a\nb"`},
		{"a\nb", `"a\nb"`},
		{"a b\n c\n===\n", `Heading "a b\nc" /Heading`},
		{"a b\nc\n---\nd", `Heading "a b\nc" /Heading "d"`},
	}
	for _, tc := range tests {
		input := []byte(tc.input)
		got := inlineShape(NewCommonMark().Parse(input))
		if got != tc.want {
			t.Errorf("input %q\nExpected[%s]\nGot     [%s]", tc.input, tc.want, got)
		}
		if string(input) != tc.input {
			t.Errorf("input %q changed to %q", tc.input, input)
		}
	}
}

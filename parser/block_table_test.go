package parser

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

// skipCodeSpanNaive is the retry the inline parser performs byte by byte:
// try the whole backtick run, then the run minus its first backtick, and so
// on. skipCodeSpan must find the same cell separators in one pass.
func skipCodeSpanNaive(data []byte, i int) int {
	if data[i] != '`' || isEscape(data, i) {
		return i
	}
	for j := i; j < len(data) && data[j] == '`'; j++ {
		if end := codeSpanEnd(data[j:]); end > 0 {
			return j + end - 1
		}
	}
	return i
}

// pipesAfterSkipping runs the tableRow scan loop with the given skipper and
// reports where it found unescaped cell separators.
func pipesAfterSkipping(skip func([]byte, int) int, data []byte) []int {
	var pipes []int
	for i := 0; i < len(data); i++ {
		i = skip(data, i)
		if data[i] == '|' && !isEscape(data, i) {
			pipes = append(pipes, i)
		}
	}
	return pipes
}

func TestSkipCodeSpanMatchesInlineRetry(t *testing.T) {
	alphabet := []byte("```` a|\\")
	rng := rand.New(rand.NewSource(1))
	for n := 0; n < 200000; n++ {
		data := make([]byte, 1+rng.Intn(24))
		for i := range data {
			data[i] = alphabet[rng.Intn(len(alphabet))]
		}
		got := pipesAfterSkipping(skipCodeSpan, data)
		want := pipesAfterSkipping(skipCodeSpanNaive, data)
		if len(got) != len(want) {
			t.Fatalf("%q: got pipes at %v, want %v", data, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%q: got pipes at %v, want %v", data, got, want)
			}
		}
	}
}

func BenchmarkSkipCodeSpanBacktickRun(b *testing.B) {
	// A run of unmatched backticks followed by text: every backtick used to
	// trigger its own scan to the end of the row.
	row := []byte("| " + strings.Repeat("`", 4000) + strings.Repeat("a", 4000) + " | c |\n")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pipesAfterSkipping(skipCodeSpan, row)
	}
}

func TestBug195(t *testing.T) {
	input := "| a | b |\n| - | - |\n|`foo|bar` | types |\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	// TODO: change expectations for https://github.com/gomarkdown/markdown/issues/195
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text\n        Code 'foo|bar'\n      TableCell\n        Text 'types'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

func TestBug198(t *testing.T) {
	// there's a space after end of table header, which used to break table parsing
	input := `| a | b|
| :--- | ---: |
| c | d |`
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text 'c'\n      TableCell\n        Text 'd'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

// https://github.com/gomarkdown/markdown/issues/274
func TestIssue274(t *testing.T) {
	input := "| a | b |\n| - | - |\n|	foo | bar |\n"
	p := NewWithExtensions(CommonExtensions)
	doc := p.Parse([]byte(input))
	var buf bytes.Buffer
	ast.Print(&buf, doc)
	got := buf.String()
	exp := "Table\n  TableHeader\n    TableRow\n      TableCell\n        Text 'a'\n      TableCell\n        Text 'b'\n  TableBody\n    TableRow\n      TableCell\n        Text '\\tfoo'\n      TableCell\n        Text 'bar'\n"
	if got != exp {
		t.Errorf("\nInput   [%#v]\nExpected[%#v]\nGot     [%#v]\n",
			input, exp, got)
	}
}

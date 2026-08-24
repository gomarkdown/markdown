package markdown

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// crashes found with go-fuzz

func TestCrash1(t *testing.T) {
	tests := []string{
		": \n\n0\n00",
		">>>0```\n\n:\n```",
		">0```\n: \n\n0\n```",
		">>>>0```\n\n:\n```",
		"0\n\n:\n00",
		">>0```\n\n:\n```",
		"[0]:<",
		">0\n>\n:\n00",
		": : \n\n\t0\n00",
		"0\n: : \n\n\t0\n00",
		"0\n\n:\n00",
		"0\n\n: [0]:<",
		"[0]:<",
	}
	for _, test := range tests {
		Parse([]byte(test), nil)
	}
}

func parseWithShortTimeout(t *testing.T, test string) {
	c := make(chan bool, 1)
	go func() {
		Parse([]byte(test), nil)
		c <- true
	}()
	select {
	case <-c:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out parsing %#v\n", test)
	}
}
func TestInfinite1(t *testing.T) {
	test := "[[[[[[\n\t: ]]]]]]\n\n: " + "\n\n:(()"
	parseWithShortTimeout(t, test)
}

func TestInfinite2(t *testing.T) {
	test := ":\x00\x00\x00\x01V\n>* \x00\x80e\n\t* \n\n:\t"
	parseWithShortTimeout(t, test)
}

func TestInfinite3(t *testing.T) {
	test := "\xa2 \n\t: \n: "
	parseWithShortTimeout(t, test)
}

func TestIssue330FuzzCrasher(t *testing.T) {
	data, err := os.ReadFile("testdata/issue330-crasher.md")
	if err != nil {
		t.Fatal(err)
	}
	parseWithShortTimeout(t, string(data))
}

// GHSA-cv23-7vc5-jfh7: empty/short link destinations panic IsSafeURL when Safelink is on.
func TestGHSA_cv23_7vc5_jfh7_SafelinkEmptyDestination(t *testing.T) {
	opts := html.RendererOptions{Flags: html.CommonFlags | html.Safelink}
	renderer := html.NewRenderer(opts)
	inputs := []string{
		"[x]()",
		"[x](.)",
		"[x](..)",
		"[x](/)",
		"[x](http://example.com)",
		"[x](javascript:alert(1))",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			// Must not panic.
			_ = ToHTML([]byte(input), nil, renderer)
		})
	}
}

// tableRow indexed data[i] after skipChar advanced i to len(data) on a row
// whose cell is only whitespace, panicking with index out of range.
func TestTableRowTrailingSpaceCrasher(t *testing.T) {
	inputs := []string{
		" -|\n| ",
		"a|b\n-|-\n| ",
		" -|\n|  \t",
	}
	for _, input := range inputs {
		// Must not panic (Tables is enabled via the default CommonExtensions).
		_ = ToHTML([]byte(input), nil, nil)
	}
}

// GHSA-85vw-wvf9-r522: a run of unmatched '[' made link() rescan the rest of
// the buffer from each '[', which is quadratic. Same for a balanced run of
// '[' ... ']' that does not resolve to a reference.
func TestGHSA_85vw_wvf9_r522_UnmatchedOpenBrackets(t *testing.T) {
	n := 32 * 1024
	input := strings.Repeat("[", n)
	parseWithShortTimeout(t, input)
	got := string(ToHTML([]byte(input), nil, nil))
	want := "<p>" + strings.Repeat("[", n) + "</p>\n"
	if got != want {
		t.Fatalf("got %d bytes, want %d", len(got), len(want))
	}
}

func TestGHSA_85vw_wvf9_r522_BalancedBrackets(t *testing.T) {
	n := 16 * 1024
	input := strings.Repeat("[", n) + strings.Repeat("]", n)
	parseWithShortTimeout(t, input)
}

func TestGHSA_85vw_wvf9_r522_InnerShortcutStillResolves(t *testing.T) {
	input := "[[[foo]]]\n\n[foo]: /url\n"
	got := string(ToHTML([]byte(input), nil, nil))
	if !strings.Contains(got, `<a href="/url">foo</a>`) {
		t.Fatalf("inner shortcut was not resolved:\n%s", got)
	}
}

func TestGHSA_85vw_wvf9_r522_InnerInlineStillResolves(t *testing.T) {
	input := "[[foo](/url)]\n"
	got := string(ToHTML([]byte(input), nil, nil))
	if !strings.Contains(got, `<a href="/url">foo</a>`) {
		t.Fatalf("inner inline link was not resolved:\n%s", got)
	}
}

func TestGHSA_85vw_wvf9_r522_NestedReferenceStyle(t *testing.T) {
	input := "[foo [bar] baz][ref]\n\n[ref]: /url\n"
	got := string(ToHTML([]byte(input), nil, nil))
	if !strings.Contains(got, `<a href="/url">`) {
		t.Fatalf("nested reference-style link was not resolved:\n%s", got)
	}
}

func TestGHSA_85vw_wvf9_r522_NestedReferenceOverride(t *testing.T) {
	p := parser.New()
	p.ReferenceOverride = func(reference string) (*parser.Reference, bool) {
		if reference == "[foo]" {
			return &parser.Reference{Link: "/url"}, true
		}
		return nil, false
	}
	got := string(ToHTML([]byte("[[foo]]\n"), p, nil))
	if !strings.Contains(got, `<a href="/url">`) {
		t.Fatalf("ReferenceOverride was not applied to nested label:\n%s", got)
	}
}

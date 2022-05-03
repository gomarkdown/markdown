package parser

import (
	"testing"
)

func TestIsFenceLine(t *testing.T) {
	tests := []struct {
		data            []byte
		syntaxRequested bool
		wantEnd         int
		wantMarker      string
		wantSyntax      string
	}{
		{
			data:       []byte("```"),
			wantEnd:    3,
			wantMarker: "```",
		},
		{
			data:       []byte("```\nstuff here\n"),
			wantEnd:    4,
			wantMarker: "```",
		},
		{
			data:            []byte("```\nstuff here\n"),
			syntaxRequested: true,
			wantEnd:         4,
			wantMarker:      "```",
		},
		{
			data:    []byte("stuff here\n```\n"),
			wantEnd: 0,
		},
		{
			data:            []byte("```"),
			syntaxRequested: true,
			wantEnd:         3,
			wantMarker:      "```",
		},
		{
			data:            []byte("``` go"),
			syntaxRequested: true,
			wantEnd:         6,
			wantMarker:      "```",
			wantSyntax:      "go",
		},
	}

	for _, test := range tests {
		var syntax *string
		if test.syntaxRequested {
			syntax = new(string)
		}
		end, marker := isFenceLine(test.data, syntax, "")
		if got, want := end, test.wantEnd; got != want {
			t.Errorf("got end %v, want %v", got, want)
		}
		if got, want := marker, test.wantMarker; got != want {
			t.Errorf("got marker %q, want %q", got, want)
		}
		if test.syntaxRequested {
			if got, want := *syntax, test.wantSyntax; got != want {
				t.Errorf("got syntax %q, want %q", got, want)
			}
		}
	}
}

func TestSanitizedAnchorName(t *testing.T) {
	tests := []string{
		"This is a header",
		"this-is-a-header",

		"This is also          a header",
		"this-is-also-a-header",

		"main.go",
		"main-go",

		"Article 123",
		"article-123",

		"<- Let's try this, shall we?",
		"let-s-try-this-shall-we",

		"        ",
		"empty",

		"Hello, 世界",
		"hello-世界",

		"世界",
		"世界",

		"⌥",
		"empty",
	}
	n := len(tests)
	for i := 0; i < n; i += 2 {
		text := tests[i]
		want := tests[i+1]
		if got := sanitizeHeadingID(text); got != want {
			t.Errorf("SanitizedAnchorName(%q):\ngot %q\nwant %q", text, got, want)
		}
	}
}

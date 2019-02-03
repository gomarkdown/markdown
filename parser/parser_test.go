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
		end, marker := isFenceLine(test.data, syntax, "```")
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
	tests := []struct {
		text string
		want string
	}{
		{
			text: "This is a header",
			want: "this-is-a-header",
		},
		{
			text: "This is also          a header",
			want: "this-is-also-a-header",
		},
		{
			text: "main.go",
			want: "main-go",
		},
		{
			text: "Article 123",
			want: "article-123",
		},
		{
			text: "<- Let's try this, shall we?",
			want: "let-s-try-this-shall-we",
		},
		{
			text: "        ",
			want: "",
		},
		{
			text: "Hello, 世界",
			want: "hello-世界",
		},
	}
	for _, test := range tests {
		if got := sanitizeAnchorName(test.text); got != test.want {
			t.Errorf("SanitizedAnchorName(%q):\ngot %q\nwant %q", test.text, got, test.want)
		}
	}
}

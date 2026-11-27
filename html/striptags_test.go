package html

import (
	"bytes"
	"strings"
	"testing"
)

// outStripped renders s through Out and Outs with tags disabled and checks
// both give the same text.
func outStripped(t *testing.T, s string) string {
	t.Helper()
	r := NewRenderer(RendererOptions{})
	r.DisableTags = 1
	var b, str bytes.Buffer
	r.Out(&b, []byte(s))
	r.Outs(&str, s)
	if b.String() != str.String() {
		t.Fatalf("Out(%q) = %q, Outs = %q", s, b.String(), str.String())
	}
	return b.String()
}

func TestOutStripsLeadingTag(t *testing.T) {
	tests := map[string]string{
		"":                  "",
		"plain":             "plain",
		"<em>":              "",
		"</em>":             "",
		"<EM>text":          "text",
		"<br/>":             "",
		"<!-- c -->":        "",
		"<?pi?>":            "",
		"<![CDATA[x]]>":     "",
		"<!DOCTYPE html>":   "",
		"<em><b>":           "<b>", // only the tag at the start
		"a<em>":             "a<em>",
		" <em>":             " <em>",
		"<":                 "<",
		"< em>":             "< em>",
		`<img src="`:        `<img src="`,
		`<a href="x">link`:  "link",
		"<a href='x'>":      "",
		"<a\nhref=x>":       "",
		"&lt;em&gt;":        "&lt;em&gt;",
		"é<em>":             "é<em>",
		"<em>\n<b>":         "\n<b>",
		"x\n<em>":           "x\n<em>",
		"<123>":             "<123>",
		"<a b='1' c=\"2\">": "",
	}
	for in, want := range tests {
		if got := outStripped(t, in); got != want {
			t.Errorf("stripped %q = %q, want %q", in, got, want)
		}
	}
}

// TestStripTagsNeedsLeadingAngle checks that htmlTagRe matches nothing in
// text that does not start with '<', so Out and Outs may skip it.
func TestStripTagsNeedsLeadingAngle(t *testing.T) {
	for _, s := range []string{
		"a<em>", "\n<em>", " </b>", "x<!-- c -->", "x<?pi?>", "x<![CDATA[y]]>",
		"x<!DOCTYPE html>", "x\n<br/>",
	} {
		if htmlTagRe.MatchString(s) {
			t.Errorf("htmlTagRe matches %q", s)
		}
	}
}

func FuzzOutStripTags(f *testing.F) {
	for _, s := range []string{"plain", "<em>x", "a<em>", "<!-- c -->", "\n<b>", "<a href='x'>"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		want := htmlTagRe.ReplaceAllString(s, "")
		if got := outStripped(t, s); got != want {
			t.Errorf("stripped %q = %q, want %q", s, got, want)
		}
		if !strings.HasPrefix(s, "<") && want != s {
			t.Errorf("htmlTagRe changed %q, which does not start with '<'", s)
		}
	})
}

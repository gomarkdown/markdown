package markdown

import (
	"net/url"
	"testing"

	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func renderHTML(input string, flags html.Flags, commonMark bool, override func([]byte) bool) string {
	p := parser.New()
	if commonMark {
		p.Opts.Flags = parser.CommonMark
	}
	r := html.NewRenderer(html.RendererOptions{Flags: flags})
	r.IsSafeURLOverride = override
	return string(ToHTML([]byte(input), p, r))
}

func TestSafelinkLinks(t *testing.T) {
	tests := []struct{ input, want string }{
		{"[x](javascript:alert(1))", "<p><tt>x</tt></p>\n"},
		{"[x](&#106;avascript:alert(1))", "<p><tt>x</tt></p>\n"},
		{"[x](&#x6a;avascript:alert(1))", "<p><tt>x</tt></p>\n"},
		{"[x](java&Tab;script:alert(1))", "<p><tt>x</tt></p>\n"},
		{"[x](javascript&colon;alert(1))", "<p><tt>x</tt></p>\n"},
		{"[x](/&#47;evil.com)", "<p><tt>x</tt></p>\n"},
		{"[x](&#47;&#47;evil.com)", "<p><tt>x</tt></p>\n"},
		// the decoded URL is what gets checked, and it's safe
		{"[x](&#47;path)", "<p><a href=\"/path\">x</a></p>\n"},
		{"[x](&#104;ttp://a.com/)", "<p><a href=\"http://a.com/\">x</a></p>\n"},
		// Safelink doesn't check hosts, so encoding this changes nothing
		{"[x](http://a.com@evil.com/)", "<p><a href=\"http://a.com@evil.com/\">x</a></p>\n"},
		{"[x](http://a.com&#64;evil.com/)", "<p><a href=\"http://a.com@evil.com/\">x</a></p>\n"},
	}
	for _, tc := range tests {
		if got := renderHTML(tc.input, html.Safelink, false, nil); got != tc.want {
			t.Errorf("input %q\n got %q\nwant %q", tc.input, got, tc.want)
		}
	}
}

// A host allowlist must judge the URL that is written to the output: raw
// http://evil.com&sol;@trusted.com/ has host trusted.com, but it's output as
// http://evil.com/@trusted.com/ whose host is evil.com.
func TestSafelinkOverrideSeesOutputURL(t *testing.T) {
	allowTrusted := func(dest []byte) bool {
		u, err := url.Parse(string(dest))
		return err == nil && u.Hostname() == "trusted.com"
	}
	tests := []struct{ input, want string }{
		{"[x](http://trusted.com/)", "<p><a href=\"http://trusted.com/\">x</a></p>\n"},
		{"[x](http://evil.com/)", "<p><tt>x</tt></p>\n"},
		{"[x](http://evil.com/@trusted.com/)", "<p><tt>x</tt></p>\n"},
		{"[x](http://evil.com&sol;@trusted.com/)", "<p><tt>x</tt></p>\n"},
		{"![x](http://evil.com&sol;@trusted.com/a.png)", "<p><tt>x</tt></p>\n"},
	}
	for _, tc := range tests {
		if got := renderHTML(tc.input, html.Safelink, false, allowTrusted); got != tc.want {
			t.Errorf("input %q\n got %q\nwant %q", tc.input, got, tc.want)
		}
	}
}

func TestSafelinkImages(t *testing.T) {
	blocked := []string{
		"javascript:alert(1)",
		"JaVaScRiPt:alert(1)",
		"&#106;avascript:alert(1)",
		"java&Tab;script:alert(1)",
		"java&#10;script:alert(1)",
		"&#32;javascript:alert(1)",
		"javascript&colon;alert(1)",
		"vbscript:x",
		"data:text/html,x",
		"data:image/svg+xml,x",
		"data:image/pngx,x",
		"//evil.com/a.png",
		"&#47;&#47;evil.com/a.png",
	}
	for _, commonMark := range []bool{false, true} {
		for _, dest := range blocked {
			input := "![a](" + dest + ")"
			got := renderHTML(input, html.Safelink, commonMark, nil)
			want := "<p><tt>a</tt></p>\n"
			if got != want {
				t.Errorf("commonMark=%v input %q\n got %q\nwant %q", commonMark, input, got, want)
			}
		}
	}

	allowed := []string{
		"http://a.com/b.png",
		"HTTPS://a.com/b.png",
		"/b.png",
		"./b.png",
		"../b.png",
		"b.png",
		"dir/b.png",
		"dir/b:c.png",
		"b.png?x=1:2",
		"#frag",
		"data:image/png;base64,AAAA",
		"DATA:image/jpeg,AAAA",
		"data:image/webp;base64,AAAA",
		"data:image/gif;base64,AAAA",
	}
	for _, commonMark := range []bool{false, true} {
		for _, dest := range allowed {
			input := "![a](" + dest + ")"
			got := renderHTML(input, html.Safelink, commonMark, nil)
			want := renderHTML(input, html.FlagsNone, commonMark, nil)
			if got != want || got == "<p><tt>a</tt></p>\n" {
				t.Errorf("commonMark=%v input %q\n got %q\nwant %q", commonMark, input, got, want)
			}
		}
	}
}

func TestSafelinkImageSurroundings(t *testing.T) {
	tests := []struct {
		input      string
		commonMark bool
		want       string
	}{
		// rendering continues normally after a blocked image
		{"![a](javascript:x) *b* ![c](/d.png)", false, "<p><tt>a</tt> <em>b</em> <img src=\"/d.png\" alt=\"c\" /></p>\n"},
		{"[![a](javascript:x)](/b)", false, "<p><a href=\"/b\"><tt>a</tt></a></p>\n"},
		// alt text is escaped, and only CommonMark parses it as markdown
		{"![<b>](javascript:x)", false, "<p><tt>&lt;b&gt;</tt></p>\n"},
		{"![&lt;b&gt;](javascript:x)", true, "<p><tt>&lt;b&gt;</tt></p>\n"},
		{"![<b>](javascript:x)", true, "<p><tt></tt></p>\n"},
		{"![a *b*](javascript:x)", false, "<p><tt>a *b*</tt></p>\n"},
		{"![a *b*](javascript:x)", true, "<p><tt>a b</tt></p>\n"},
		// images nested in CommonMark alt text are flattened to text
		{"![a ![b](javascript:x) c](/d.png)", true, "<p><img src=\"/d.png\" alt=\"a b c\"></p>\n"},
		{"![a ![b](/c.png) d](javascript:x)", true, "<p><tt>a b d</tt></p>\n"},
	}
	for _, tc := range tests {
		if got := renderHTML(tc.input, html.Safelink, tc.commonMark, nil); got != tc.want {
			t.Errorf("commonMark=%v input %q\n got %q\nwant %q", tc.commonMark, tc.input, got, tc.want)
		}
	}
}

// Without Safelink images are unrestricted, as before.
func TestImagesWithoutSafelink(t *testing.T) {
	got := renderHTML("![a](data:text/html,x)", html.FlagsNone, false, nil)
	want := "<p><img src=\"data:text/html,x\" alt=\"a\" /></p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLinkDestinationEntities(t *testing.T) {
	tests := []struct{ input, want string }{
		{"[x](/a?b=1&amp;c=2)", "<p><a href=\"/a?b=1&amp;c=2\">x</a></p>\n"},
		// a legacy entity name without a semicolon is query string text
		{"[x](/a?b=1&copy=2)", "<p><a href=\"/a?b=1&amp;copy=2\">x</a></p>\n"},
		{"[x](/a?b=1&amp=2)", "<p><a href=\"/a?b=1&amp;amp=2\">x</a></p>\n"},
		{"![x](/a?b=1&copy=2)", "<p><img src=\"/a?b=1&amp;copy=2\" alt=\"x\" /></p>\n"},
		// only a complete entity name is decoded
		{"[x](/a&notit;)", "<p><a href=\"/a&amp;notit;\">x</a></p>\n"},
		{"[x](/a&nosuchentity;)", "<p><a href=\"/a&amp;nosuchentity;\">x</a></p>\n"},
		// semicolon-terminated references are decoded
		{"[x](/caf&eacute;)", "<p><a href=\"/café\">x</a></p>\n"},
		{"[x](/caf&#233;)", "<p><a href=\"/café\">x</a></p>\n"},
		{"[x](/caf&#xE9;)", "<p><a href=\"/café\">x</a></p>\n"},
		{"[x](/a&semi;b)", "<p><a href=\"/a;b\">x</a></p>\n"},
		{"![x](/a&#32;b.png)", "<p><img src=\"/a b.png\" alt=\"x\" /></p>\n"},
		// decoding happens exactly once
		{"[x](/a&amp;lt;)", "<p><a href=\"/a&amp;lt;\">x</a></p>\n"},
		{"![x](/a&amp;#64;)", "<p><img src=\"/a&amp;#64;\" alt=\"x\" /></p>\n"},
		// decoded text can't break out of the attribute
		{"[x](/a&quot;&lt;b&gt;)", "<p><a href=\"/a&quot;&lt;b&gt;\">x</a></p>\n"},
	}
	for _, tc := range tests {
		if got := renderHTML(tc.input, html.FlagsNone, false, nil); got != tc.want {
			t.Errorf("input %q\n got %q\nwant %q", tc.input, got, tc.want)
		}
	}
}

func TestLinkRelAttrsUseDecodedDestination(t *testing.T) {
	flags := html.NofollowLinks
	got := renderHTML("[x](/&#47;evil.com)", flags, false, nil)
	want := "<p><a href=\"//evil.com\" rel=\"nofollow\">x</a></p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

package html

import (
	"bytes"
	"testing"
)

func TestEscapeHTMLCallouts(t *testing.T) {
	buf := &bytes.Buffer{}
	code := []byte(`println("hello")
more code //<<4>>
bliep bliep
`)
	out := `println(&quot;hello&quot;)
more code <span class="callout">4</span>
bliep bliep
`
	opts := RendererOptions{}
	opts.Comments = [][]byte{[]byte("//")}

	r := NewRenderer(opts)
	r.EscapeHTMLCallouts(buf, code)

	if buf.String() != out {
		t.Error("callout code block not correctly parsed")
	}
}

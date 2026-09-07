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

func TestEscapeHTMLCalloutsMultipleComments(t *testing.T) {
	code := []byte(`a //<<1>>
b #<<2>>
c <<9>>
`)
	out := `a <span class="callout">1</span>
b <span class="callout">2</span>
c &lt;&lt;9&gt;&gt;
`
	orders := [][][]byte{
		{[]byte("//"), []byte("#")},
		{[]byte("#"), []byte("//")},
	}
	for _, comments := range orders {
		buf := &bytes.Buffer{}
		opts := RendererOptions{}
		opts.Comments = comments
		NewRenderer(opts).EscapeHTMLCallouts(buf, code)
		if buf.String() != out {
			t.Errorf("comments=%q\ngot:\n%s\nwant:\n%s", comments, buf.String(), out)
		}
	}
}

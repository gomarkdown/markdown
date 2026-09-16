package parser

import (
	"math/rand"
	"strings"
	"testing"
)

// TestCodeSpanClosersMatchReference walks random inputs the way Inline
// does, one byte at a time, and also in random order to defeat the cache,
// checking that the memoised closer table agrees with codeSpanEnd.
func TestCodeSpanClosersMatchReference(t *testing.T) {
	alphabet := []byte("```` a\n")
	rng := rand.New(rand.NewSource(1))
	check := func(p *Parser, data []byte, off int) {
		if data[off] != '`' {
			return
		}
		nb, closers := p.codeSpanClosers(data, off)
		if wantNb := skipChar(data, off, '`') - off; nb != wantNb {
			t.Fatalf("%q at %d: got run length %d, want %d", data, off, nb, wantNb)
		}
		got := closers[nb]
		want := codeSpanEnd(data[off:])
		if want != 0 {
			want += off
		}
		if got != want {
			t.Fatalf("%q at %d: got closer %d, want %d", data, off, got, want)
		}
	}
	for n := 0; n < 100000; n++ {
		data := make([]byte, 1+rng.Intn(24))
		for i := range data {
			data[i] = alphabet[rng.Intn(len(alphabet))]
		}
		p := New()
		for off := range data {
			check(p, data, off)
		}
		for i := 0; i < len(data); i++ {
			check(p, data, rng.Intn(len(data)))
		}
	}
}

func BenchmarkInlineBacktickRun(b *testing.B) {
	// A run of unmatched backticks followed by text: the inline loop
	// retries the run from each of its bytes.
	input := []byte(strings.Repeat("`", 4000) + strings.Repeat("a", 4000) + "\n")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := New()
		p.Parse(input)
	}
}

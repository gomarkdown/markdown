package parser

import "testing"

func TestBracketTableLookup(t *testing.T) {
	tbl := bracketTable{data: []byte("[a[b]c]")}
	closeAt, nested, ok := tbl.lookup(0)
	if !ok || closeAt != 6 || !nested {
		t.Fatalf("open 0: close=%d nested=%v ok=%v", closeAt, nested, ok)
	}
	closeAt, nested, ok = tbl.lookup(2)
	if !ok || closeAt != 4 || nested {
		t.Fatalf("open 2: close=%d nested=%v ok=%v", closeAt, nested, ok)
	}

	tbl = bracketTable{data: []byte("[[[")}
	if _, _, ok := tbl.lookup(0); ok {
		t.Fatal("unmatched '[' should not report a close")
	}

	tbl = bracketTable{data: []byte(`[foo \[bar\]]`)}
	closeAt, nested, ok = tbl.lookup(0)
	if !ok || closeAt != 12 || nested {
		t.Fatalf("escaped inner: close=%d nested=%v ok=%v", closeAt, nested, ok)
	}
}

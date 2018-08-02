package parser

import "testing"

func TestLinesUntilEmpty(t *testing.T) {
	data := []byte(`Figure: foo bar bar foo
foo bar

first text after empty line`)

	l := New().linesUntilEmpty(data)
	if l != 33 {
		t.Errorf("want %d, got %d", 33, l)
	}

	data = []byte(`Figure: foo bar bar foo
foo bar
`)
	l = New().linesUntilEmpty(data)
	if l != 32 {
		t.Errorf("want %d, got %d", 33, l)
	}
}

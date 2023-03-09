package parser

import "testing"

func TestLinesUntilEmpty(t *testing.T) {
	data := `Figure: foo bar bar foo
foo bar

first text after empty line`

	l := LinesUntilEmpty([]byte(data))
	if l != 33 {
		t.Errorf("want %d, got %d", 33, l)
	}

	data = `Figure: foo bar bar foo
foo bar
`
	l = LinesUntilEmpty([]byte(data))
	if l != 32 {
		t.Errorf("want %d, got %d", 33, l)
	}
}

func TestCaptionID(t *testing.T) {
	data := `Figure: foo bar bar foo
first text {#no-heading} after empty line`

	if id, _ := captionID([]byte(data)); id != "" {
		t.Errorf("want nothing, got %s", id)
	}

	data = `Figure: foo bar bar foo
foo bar {#heading-id}
`
	if id, _ := captionID([]byte(data)); id == "" {
		t.Errorf("want %s, got nothing", "heading-id")
	}
}

package parser

import "testing"

func TestUpdateWd(t *testing.T) {
	p := "/tmp"

	if x := updateWd(p, "/new"); x != "/new" {
		t.Errorf("want %s, got %s", "/new", x)
	}

	if x := updateWd(p, "new"); x != "/tmp/new" {
		t.Errorf("want %s, got %s", "/tmp/new", x)
	}
}

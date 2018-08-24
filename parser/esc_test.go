package parser

import "testing"

func TestIsEscape(t *testing.T) {
	if x := `\a`; !isEscape([]byte(x), 1) {
		t.Errorf("expected escape for %q, got false", x)
	}
	if x := `\\\a`; !isEscape([]byte(x), 3) {
		t.Errorf("expected escape for %q, got false", x)
	}
	if x := `b\\\a`; !isEscape([]byte(x), 4) {
		t.Errorf("expected escape for %q, got false", x)
	}

	if x := `\\a`; isEscape([]byte(x), 2) {
		t.Errorf("expected no escape for %q, got true", x)
	}
	if x := `\\\\a`; isEscape([]byte(x), 4) {
		t.Errorf("expected no escape for %q, got true", x)
	}
}

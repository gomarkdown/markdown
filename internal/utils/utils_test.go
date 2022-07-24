package utils

import "testing"

func TestIsAlnum(t *testing.T) {
	for _, c := range []byte{'4', 'z', 'A', '0', '9'} {
		if !IsAlnum(c) {
			t.Errorf("'%c' not recognized as an alphanumeric symbol", c)
		}
	}

	for _, c := range []byte{'%', '_', '-', '@', '!'} {
		if IsAlnum(c) {
			t.Errorf("'%c' recognized as an alphanumeric symbol", c)
		}
	}
}

func TestIsDigit(t *testing.T) {
	for _, c := range []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'} {
		if !IsDigit(c) {
			t.Errorf("'%c' not recognized as a numeric symbol", c)
		}
	}

	for _, c := range []byte{'%', '_', '-', '@', '!', 'z', 'A', 'x'} {
		if IsDigit(c) {
			t.Errorf("'%c' recognized as a numeric symbol", c)
		}
	}
}

func TestIsLetter(t *testing.T) {
	for _, c := range []byte{'A', 'b', 'C', 'd', 'E', 'f', 'G', 'Z', 'X', 'Y'} {
		if !IsLetter(c) {
			t.Errorf("'%c' not recognized as a letter", c)
		}
	}

	for _, c := range []byte{'%', '_', '-', '@', '!', '0', '1', '3'} {
		if IsLetter(c) {
			t.Errorf("'%c' recognized as a letter", c)
		}
	}
}

func TestIsMailto(t *testing.T) {
	if b := []byte("mailto:doe@example.com"); !IsMailto(b) {
		t.Errorf("'%s' is not mailto:", b)
	}

	if b := []byte("http://www.example.com"); IsMailto(b) {
		t.Errorf("'%s' is not mailto:", b)
	}
}

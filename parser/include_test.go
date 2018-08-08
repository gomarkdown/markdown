package parser

import "testing"

func TestUpdateWd(t *testing.T) {
	p := "/tmp"

	if x := updateWd(p, "/new/foo"); x != "/new" {
		t.Errorf("want %s, got %s", "/new", x)
	}

	if x := updateWd(p, "new/new"); x != "/tmp/new" {
		t.Errorf("want %s, got %s", "/tmp/new", x)
	}
}

func TestIsInclude(t *testing.T) {
	tests := []struct {
		data []byte
		file string
		addr string
		read int
	}{
		{
			[]byte("{{foo}}"),
			"foo", "", 7,
		},
		{
			[]byte("{{foo}}  "),
			"foo", "", 7,
		},
		{
			[]byte("{{foo}}[a]"),
			"foo", "a", 10,
		},
		{
			[]byte("{{foo}}[a  ]  "),
			"foo", "a  ", 12,
		},
		{
			[]byte("{{foo}}a]"),
			"foo", "", 7,
		},
		// fails
		{
			[]byte("{foo}}"),
			"", "", 0,
		},
		{
			[]byte("{foo}"),
			"", "", 0,
		},
		{
			[]byte("{{foo}}[a"),
			"", "", 0,
		},
	}

	p := New()
	for i, test := range tests {
		file, addr, read := p.isInclude(test.data)
		if file != test.file {
			t.Errorf("test %d, want %s, got %s", i, test.file, file)
		}
		if string(addr) != test.addr {
			t.Errorf("test %d, want %s, got %s", i, test.addr, addr)
		}
		if read != test.read {
			t.Errorf("test %d, want %d, got %d", i, test.read, read)
		}
	}
}

func TestIsCodeInclude(t *testing.T) {
	tests := []struct {
		data []byte
		file string
		addr string
		read int
	}{
		{
			[]byte("<{{foo}}"),
			"foo", "", 8,
		},
		{
			[]byte("<{{foo}}  "),
			"foo", "", 8,
		},
	}

	p := New()
	for i, test := range tests {
		file, addr, read := p.isInclude(test.data)
		if file != test.file {
			t.Errorf("test %d, want %s, got %s", i, test.file, file)
		}
		if string(addr) != test.addr {
			t.Errorf("test %d, want %s, got %s", i, test.addr, addr)
		}
		if read != test.read {
			t.Errorf("test %d, want %d, got %d", i, test.read, read)
		}
	}
}

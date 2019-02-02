package parser

import (
	"runtime"
	"testing"
)

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
		{
			[]byte("   {{foo}}"),
			"foo", "", 10,
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
		{
			[]byte("   <{{foo}}  "),
			"foo", "", 11,
		},
	}

	p := New()
	for i, test := range tests {
		file, addr, read := p.isCodeInclude(test.data)
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

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func TestPush(t *testing.T) {
	if isWindows() {
		return
	}
	i := newIncStack()
	if i.Push("/new/foo"); i.stack[0] != "/new" {
		t.Errorf("want %s, got %s", "/new", i.stack[0])
	}

	if i.Push("new/new"); i.stack[1] != "/new/new" {
		t.Errorf("want %s, got %s", "/new/new", i.stack[1])
	}
}

func TestPop(t *testing.T) {
	i := newIncStack()
	if i.Push("/new/foo"); i.stack[0] != "/new" {
		t.Errorf("want %s, got %s", "/new", i.stack[0])
	}
	i.Pop()
	if len(i.stack) != 0 {
		t.Errorf("after pop, want %d, got %d", 0, len(i.stack))
	}
}

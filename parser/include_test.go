package parser

import (
	"runtime"
	"testing"
)

func TestIsInclude(t *testing.T) {
	tests := []struct {
		data string
		file string
		addr string
		read int
	}{
		{
			"{{foo}}",
			"foo", "", 7,
		},
		{
			"{{foo}}  ",
			"foo", "", 7,
		},
		{
			"{{foo}}[a]",
			"foo", "a", 10,
		},
		{
			"{{foo}}[a  ]  ",
			"foo", "a  ", 12,
		},
		{
			"{{foo}}a]",
			"foo", "", 7,
		},
		{
			"   {{foo}}",
			"foo", "", 10,
		},
		// fails
		{
			"{foo}}",
			"", "", 0,
		},
		{
			"{foo}",
			"", "", 0,
		},
		{
			"{{foo}}[a",
			"", "", 0,
		},
	}

	p := New()
	for i, test := range tests {
		file, addr, read := p.isInclude([]byte(test.data))
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

// Include directives at EOF used to panic: caption look-ahead did data[consumed+1:]
// even when consumed == len(data).
func TestIncludeAtEOFNoPanic(t *testing.T) {
	inputs := []string{
		"{{foo}}",
		"{{foo.md}}",
		"<{{foo}}",
		"  {{foo}}",
		"{{foo}}[a]",
		"<{{foo}}[a]",
		"{{a}}\n{{b}}",
		"para\n\n{{foo}}",
		"{{foo}}\n",
		"{{foo}}\nFigure: caption text\n",
		"{{foo}}\nTable: caption text\n",
		"{{foo}}\nQuote: caption text\n",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			p := NewWithExtensions(CommonExtensions | Includes)
			p.Opts = Options{
				ReadIncludeFn: func(from, path string, address []byte) []byte {
					return []byte("included line\n")
				},
			}
			// Must not panic.
			p.Parse([]byte(input))
		})
	}
}

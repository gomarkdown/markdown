package parser

import (
	"path"
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

	for i, test := range tests {
		file, addr, read := isInclude([]byte(test.data))
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

	for i, test := range tests {
		file, addr, read := isCodeInclude(test.data)
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

func TestPush(t *testing.T) {
	i := newIncStack()
	if i.Push("/new/foo"); i.stack[0] != "/new" {
		t.Errorf("want %s, got %s", "/new", i.stack[0])
	}

	if i.Push("new/new"); i.stack[1] != "/new/new" {
		t.Errorf("want %s, got %s", "/new/new", i.stack[1])
	}
}

func TestPushWindowsPath(t *testing.T) {
	i := newIncStack()
	i.Push(`C:\dir\file`)
	if i.stack[0] != "C:/dir" {
		t.Errorf("drive path: want C:/dir, got %s", i.stack[0])
	}
	if got := path.Join(i.Last(), "a.md"); got != "C:/dir/a.md" {
		t.Errorf("relative include: want C:/dir/a.md, got %s", got)
	}
	i.Push("a.md")
	if i.stack[1] != "C:/dir" {
		t.Errorf("relative include dir: want C:/dir, got %s", i.stack[1])
	}

	j := newIncStack()
	j.Push("C:/dir/file")
	if j.stack[0] != "C:/dir" {
		t.Errorf("slash drive path: want C:/dir, got %s", j.stack[0])
	}

	k := newIncStack()
	k.Push("/docs/a.md")
	k.Push(`D:\other\b.md`)
	if k.stack[1] != "D:/other" {
		t.Errorf("drive path joined onto the previous directory: %s", k.stack[1])
	}
	k.Push(`sub\c.md`)
	if k.stack[2] != "D:/other/sub" {
		t.Errorf("backslash relative path: want D:/other/sub, got %s", k.stack[2])
	}

	u := newIncStack()
	u.Push("/docs/a.md")
	u.Push(`\\server\share\file.md`)
	if u.stack[1] != "/server/share" {
		t.Errorf("UNC path: want /server/share, got %s", u.stack[1])
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

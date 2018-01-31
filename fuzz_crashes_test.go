package markdown

import (
	"testing"
	"time"
)

// crashes found with go-fuzz

func TestCrash1(t *testing.T) {
	tests := []string{
		": \n\n0\n00",
		">>>0```\n\n:\n```",
		">0```\n: \n\n0\n```",
		">>>>0```\n\n:\n```",
		"0\n\n:\n00",
		">>0```\n\n:\n```",
		"[0]:<",
		">0\n>\n:\n00",
		": : \n\n\t0\n00",
		"0\n: : \n\n\t0\n00",
		"0\n\n:\n00",
		"0\n\n: [0]:<",
		"[0]:<",
	}
	for _, test := range tests {
		Parse([]byte(test), nil)
	}
}

// TODO: this enters infinite loop
func NoTestInfinite1(t *testing.T) {
	test := "[[[[[[\n\t: ]]]]]]\n\n: " + "\n\n:(()"
	c := make(chan bool, 1)
	go func() {
		Parse([]byte(test), nil)
		c <- true
	}()
	select {
	case <-c:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out")
	}
}

// TODO: this enters infinite loop
/*
program hanged (timeout 10 seconds)

SIGABRT: abort
PC=0x10bf96a m=0 sigcode=0

goroutine 1 [running]:
github.com/gomarkdown/markdown/ast.LastChild(0x120ea60, 0xc42014b170, 0x120eb20, 0xc42005be00)
	ast/node.go:349 +0x2a fp=0xc42049dbd0 sp=0xc42049dba0 pc=0x10bf96a
github.com/gomarkdown/markdown/parser.endsWithBlankLine(0x120ea60, 0xc42014b170, 0x0)
	parser/block.go:1202 +0x6b fp=0xc42049dc00 sp=0xc42049dbd0 pc=0x10fea0b
github.com/gomarkdown/markdown/parser.finalizeList.func3(0xc42049dc88, 0xc42044c000)
	parser/block.go:1226 +0x4f fp=0xc42049dc28 sp=0xc42049dc00 pc=0x1112cbf
github.com/gomarkdown/markdown/parser.finalizeList(0xc42011c3f0)
	parser/block.go:1226 +0x11a fp=0xc42049dcb8 sp=0xc42049dc28 pc=0x10febea
github.com/gomarkdown/markdown/parser.(*Parser).list(0xc42045b200, 0x12e600e, 0x7, 0x1ffff2, 0x16, 0x0)
	parser/block.go:1187 +0x27b fp=0xc42049dd30 sp=0xc42049dcb8 pc=0x10fe90b
github.com/gomarkdown/markdown/parser.(*Parser).paragraph(0xc42045b200, 0x12e6000, 0x15, 0x200000, 0x0)
	parser/block.go:1482 +0xfd4 fp=0xc42049de00 sp=0xc42049dd30 pc=0x11016d4
github.com/gomarkdown/markdown/parser.(*Parser).block(0xc42045b200, 0x12e6000, 0x15, 0x200000)
	parser/block.go:247 +0x6b8 fp=0xc42049de40 sp=0xc42049de00 pc=0x10f64e8
github.com/gomarkdown/markdown/parser.(*Parser).Parse(0xc42045b200, 0x12e6000, 0x15, 0x200000, 0x8, 0x5a71671a)
	parser/parser.go:241 +0x65 fp=0xc42049de78 sp=0xc42049de40 pc=0x110d295
github.com/gomarkdown/markdown.Parse(0x12e6000, 0x15, 0x200000, 0x0, 0x1289f8c0, 0xa23912f)
	markdown.go:46 +0x9a fp=0xc42049deb8 sp=0xc42049de78 pc=0x1118caa
github.com/gomarkdown/markdown.Fuzz(0x12e6000, 0x15, 0x200000, 0x3)
	fuzz.go:7 +0x60 fp=0xc42049def8 sp=0xc42049deb8 pc=0x1118bf0
go-fuzz-dep.Main(0x11606d8)
	/var/folders/v_/ksw1dqvd59v790zk2wqf_t_80000gn/T/go-fuzz-build351448921/goroot/src/go-fuzz-dep/main.go:49 +0xad fp=0xc42049df68 sp=0xc42049def8 pc=0x1065a4d
main.main()
	go.fuzz.main/main.go:10 +0x2d fp=0xc42049df80 sp=0xc42049df68 pc=0x1118dad
runtime.main()
*/
func NoTestInfinite2(t *testing.T) {
	test := ":\x00\x00\x00\x01V\n>* \x00\x80e\n\t* \n\n:\t"

	c := make(chan bool, 1)
	go func() {
		Parse([]byte(test), nil)
		c <- true
	}()
	select {
	case <-c:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out")
	}
}

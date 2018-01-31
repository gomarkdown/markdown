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

func parseWithShortTimeout(t *testing.T, test string) {
	c := make(chan bool, 1)
	go func() {
		Parse([]byte(test), nil)
		c <- true
	}()
	select {
	case <-c:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out parsing %#v\n", test)
	}
}
func TestInfinite1(t *testing.T) {
	test := "[[[[[[\n\t: ]]]]]]\n\n: " + "\n\n:(()"
	parseWithShortTimeout(t, test)
}

func TestInfinite2(t *testing.T) {
	test := ":\x00\x00\x00\x01V\n>* \x00\x80e\n\t* \n\n:\t"
	parseWithShortTimeout(t, test)
}

func TestInfinite3(t *testing.T) {
	test := "\xa2 \n\t: \n: "
	parseWithShortTimeout(t, test)
}

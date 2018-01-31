package markdown

import "testing"

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
		"[[[[[[\n\t: ]]]]]]\n\n: " + "\n\n:(()",
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

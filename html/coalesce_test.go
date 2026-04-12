package html

import (
	"reflect"
	"testing"
)

func TestCoalesceClassAttrs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "no class attrs",
			in:   []string{`id="foo"`},
			want: []string{`id="foo"`},
		},
		{
			name: "single class attr unchanged",
			in:   []string{`class="language-go"`},
			want: []string{`class="language-go"`},
		},
		{
			name: "two class attrs merged",
			in:   []string{`class="language-yml"`, `class="my-class"`},
			want: []string{`class="language-yml my-class"`},
		},
		{
			name: "class attrs with other attrs preserved",
			in:   []string{`class="language-go"`, `id="code1"`, `class="highlight"`},
			want: []string{`class="language-go highlight"`, `id="code1"`},
		},
		{
			name: "empty input",
			in:   nil,
			want: nil,
		},
		{
			name: "no attrs at all",
			in:   []string{},
			want: []string{},
		},
		{
			name: "empty class value stripped",
			in:   []string{`class=""`, `class="my-class"`},
			want: []string{`class="my-class"`},
		},
		{
			name: "multi-word class values preserved",
			in:   []string{`class="foo bar"`, `class="baz"`},
			want: []string{`class="foo bar baz"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := coalesceClassAttrs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("coalesceClassAttrs(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

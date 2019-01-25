package main

import (
	"strconv"
	"testing"
)

func TestRemoveString(t *testing.T) {
	cases := []struct {
		slice, output []string
		str           string
	}{
		{[]string{"abc", "def"}, []string{"abc"}, "def"},
		{[]string{"abc", "def", "ghi"}, []string{"abc", "ghi"}, "def"},
		{[]string{"abc", "def"}, []string{"abc", "def"}, "jklmn"},
	}
	for i, tc := range cases {
		t.Run("case_"+strconv.Itoa(i), func(t *testing.T) {
			out := removeString(tc.slice, tc.str)
			if len(out) != len(tc.output) {
				t.Fatalf("got unequal lengths; expected %v but got %v", tc.output, out)
			}
			for j := range out {
				if out[j] != tc.output[j] {
					t.Errorf("got %q at %v but got %q", out[j], tc.output[j], j)
				}
			}
		})
	}
}

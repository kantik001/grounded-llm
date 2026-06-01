package main

import "testing"

func TestSafeFilename(t *testing.T) {
	cases := []struct {
		name  string
		ok    bool
	}{
		{"article1.txt", true},
		{"my-article_v2.txt", true},
		{"../etc/passwd", false},
		{"article.txt.exe", false},
		{"кириллица.txt", false},
	}
	for _, tc := range cases {
		got := safeFilename.MatchString(tc.name)
		if got != tc.ok {
			t.Errorf("%q: got %v want %v", tc.name, got, tc.ok)
		}
	}
}

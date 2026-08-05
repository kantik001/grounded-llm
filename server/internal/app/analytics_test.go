package app

import "testing"

func TestParseAnalyticsDays(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 7},
		{"7", 7},
		{"30", 30},
		{"0", 7},
		{"-1", 7},
		{"abc", 7},
		{"120", 90},
	}
	for _, tc := range cases {
		if got := parseAnalyticsDays(tc.in); got != tc.want {
			t.Fatalf("parseAnalyticsDays(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestComputeVerifyPassRate(t *testing.T) {
	if got := computeVerifyPassRate(0, 0); got != 0 {
		t.Fatalf("empty: got %v", got)
	}
	if got := computeVerifyPassRate(8, 2); got != 80 {
		t.Fatalf("80%%: got %v", got)
	}
}

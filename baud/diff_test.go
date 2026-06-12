package baud

import (
	"strings"
	"testing"
)

// diffN is the *int literal helper for expected gutter numbers.
func diffN(n int) *int { return &n }

func diffNumEq(got, want *int) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	return *got == *want
}

func TestParseUnifiedNumbersGutters(t *testing.T) {
	src := strings.Join([]string{
		"--- a/main.go",
		"+++ b/main.go",
		"@@ -1,3 +1,4 @@ func main() {",
		" package main",
		"-var x = 1",
		"+var x = 2",
		"+var y = 3",
		" func main() {}",
		"",
	}, "\n")

	got, err := ParseUnified(src)
	if err != nil {
		t.Fatalf("ParseUnified: %v", err)
	}
	want := []DiffLine{
		{Kind: DiffHunk, Text: "@@ -1,3 +1,4 @@ func main() {"},
		{Kind: DiffCtx, OldN: diffN(1), NewN: diffN(1), Text: "package main"},
		{Kind: DiffDel, OldN: diffN(2), Text: "var x = 1"},
		{Kind: DiffAdd, NewN: diffN(2), Text: "var x = 2"},
		{Kind: DiffAdd, NewN: diffN(3), Text: "var y = 3"},
		{Kind: DiffCtx, OldN: diffN(3), NewN: diffN(4), Text: "func main() {}"},
	}
	if len(got) != len(want) {
		t.Fatalf("ParseUnified returned %d lines, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		g := got[i]
		if g.Kind != w.Kind || g.Text != w.Text ||
			!diffNumEq(g.OldN, w.OldN) || !diffNumEq(g.NewN, w.NewN) {
			t.Errorf("line %d = %+v, want %+v", i, g, w)
		}
	}
}

func TestParseUnifiedMultipleHunks(t *testing.T) {
	src := strings.Join([]string{
		"@@ -10,2 +10,2 @@",
		" a",
		"-b",
		"+B",
		"@@ -40 +40 @@", // lengths omitted — still valid
		"-c",
		"+C",
	}, "\n")

	got, err := ParseUnified(src)
	if err != nil {
		t.Fatalf("ParseUnified: %v", err)
	}
	if len(got) != 7 {
		t.Fatalf("got %d lines, want 7", len(got))
	}
	// The second hunk resets both counters from its header.
	if got[5].Kind != DiffDel || !diffNumEq(got[5].OldN, diffN(40)) {
		t.Errorf("second-hunk del = %+v, want old gutter 40", got[5])
	}
	if got[6].Kind != DiffAdd || !diffNumEq(got[6].NewN, diffN(40)) {
		t.Errorf("second-hunk add = %+v, want new gutter 40", got[6])
	}
}

func TestParseUnifiedNoNewlineMarker(t *testing.T) {
	src := "@@ -1 +1 @@\n-old\n+new\n\\ No newline at end of file"
	got, err := ParseUnified(src)
	if err != nil {
		t.Fatalf("ParseUnified: %v", err)
	}
	last := got[len(got)-1]
	if last.Kind != DiffCtx || last.OldN != nil || last.NewN != nil {
		t.Errorf("no-newline marker = %+v, want ctx row with blank gutters", last)
	}
}

func TestParseUnifiedMalformed(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"empty input", ""},
		{"no hunks", "--- a/main.go\n+++ b/main.go"},
		{"garbage before hunk", "hello world\n@@ -1 +1 @@\n a"},
		{"malformed hunk header", "@@ nonsense @@\n a"},
		{"missing new range", "@@ -1,2 @@\n a"},
		{"bare @@", "@@"},
		{"unmarked line in hunk", "@@ -1 +1 @@\n*** what"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// must error, never panic
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ParseUnified panicked: %v", r)
				}
			}()
			if got, err := ParseUnified(c.src); err == nil {
				t.Errorf("ParseUnified(%q) = %+v, want error", c.src, got)
			}
		})
	}
}

func TestPaginationThousands(t *testing.T) {
	cases := map[int]string{
		0:        "0",
		7:        "7",
		999:      "999",
		1000:     "1,000",
		12403:    "12,403",
		1234567:  "1,234,567",
		-1234567: "-1,234,567",
	}
	for n, want := range cases {
		if got := paginationThousands(n); got != want {
			t.Errorf("paginationThousands(%d) = %q, want %q", n, got, want)
		}
	}
}

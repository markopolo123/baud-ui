package baudui

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
)

// expectedBundle reimplements the `just css` recipe from the committed
// sources: cat tokens.css base.css components/*.css (sorted by filename)
// utilities.css — raw concatenation, no separators.
func expectedBundle(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	read := func(p string) {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		buf.Write(b)
	}
	read("assets/css/tokens.css")
	read("assets/css/base.css")
	entries, err := os.ReadDir("assets/css/components")
	if err != nil {
		t.Fatalf("read components dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".css" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		t.Fatal("no component CSS sources found")
	}
	for _, n := range names {
		read("assets/css/components/" + n)
	}
	read("assets/css/utilities.css")
	return buf.Bytes()
}

// CSS() must be byte-identical to the dist/baud.css bundle `just css`
// would produce from the same sources.
func TestCSSMatchesJustCSSBundle(t *testing.T) {
	want := expectedBundle(t)
	got := CSS()
	if !bytes.Equal(got, want) {
		t.Fatalf("CSS() differs from the `just css` concatenation: got %d bytes, want %d bytes", len(got), len(want))
	}
}

// Ordering is pinned: tokens first, utilities last, component files in
// sorted-filename order in between.
func TestCSSLayerOrderPinned(t *testing.T) {
	bundle := CSS()
	tokens, err := os.ReadFile("assets/css/tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	utilities, err := os.ReadFile("assets/css/utilities.css")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(bundle, tokens) {
		t.Error("bundle does not start with tokens.css")
	}
	if !bytes.HasSuffix(bundle, utilities) {
		t.Error("bundle does not end with utilities.css")
	}

	entries, err := os.ReadDir("assets/css/components")
	if err != nil {
		t.Fatal(err)
	}
	prev := -1
	prevName := ""
	for _, e := range entries { // os.ReadDir returns sorted-by-filename
		if e.IsDir() || filepath.Ext(e.Name()) != ".css" {
			continue
		}
		src, err := os.ReadFile("assets/css/components/" + e.Name())
		if err != nil {
			t.Fatal(err)
		}
		idx := bytes.Index(bundle, src)
		if idx < 0 {
			t.Fatalf("component %s missing from bundle", e.Name())
		}
		if idx <= prev {
			t.Errorf("component %s (offset %d) out of order after %s (offset %d)", e.Name(), idx, prevName, prev)
		}
		prev, prevName = idx, e.Name()
	}
}

// CSS() is cached: repeated calls return the same bytes.
func TestCSSStable(t *testing.T) {
	a, b := CSS(), CSS()
	if !bytes.Equal(a, b) {
		t.Fatal("CSS() not stable across calls")
	}
	if len(a) > 0 && &a[0] != &b[0] {
		t.Fatal("CSS() rebuilt the bundle instead of returning the cached one")
	}
}

// HS() serves the committed behaviors file verbatim, and the LAST behavior
// declared must be ParseHealth — the e2e readiness sentinel that must stay
// at the end of the file.
func TestHSBehaviorsFile(t *testing.T) {
	want, err := os.ReadFile("assets/baud._hs")
	if err != nil {
		t.Fatal(err)
	}
	got := HS()
	if !bytes.Equal(got, want) {
		t.Fatal("HS() differs from assets/baud._hs")
	}
	re := regexp.MustCompile(`(?m)^behavior\s+(\w+)`)
	matches := re.FindAllSubmatch(got, -1)
	if len(matches) == 0 {
		t.Fatal("no behavior declarations found in HS()")
	}
	if last := string(matches[len(matches)-1][1]); last != "ParseHealth" {
		t.Fatalf("last behavior is %q, want ParseHealth (must stay last)", last)
	}
}

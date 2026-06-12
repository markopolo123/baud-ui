//go:build e2e

package e2e

// Generic test helpers shared by every e2e file: server/browser bootstrap and
// computed-style resolution. Component test files define ONLY
// component-prefixed helpers (see docs/WAYS_OF_WORKING.md review checklist).

import (
	"fmt"
	"math"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"

	"github.com/markopolo123/baud-ui/demo"
)

// Theme-resolved accent pairs asserted across suites (assets/css/tokens.css):
//
//	t-gruvbox --accent #fabd2f → rgb(250, 189, 47), --on-accent #1d2021 → rgb(29, 32, 33)
//	t-mocha   --accent #89b4fa → rgb(137, 180, 250), --on-accent #11111b → rgb(17, 17, 27)
const (
	gruvAccent   = "rgb(250, 189, 47)"
	gruvOnAccent = "rgb(29, 32, 33)"
	mochaAccent  = "rgb(137, 180, 250)"
	mochaOnAcc   = "rgb(17, 17, 27)"
)

// startDemo starts the real demo handler on a random port.
func startDemo(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(demo.NewMux())
	t.Cleanup(srv.Close)
	return srv
}

// startBrowser launches headless chromium and returns an open page.
func startBrowser(t *testing.T) playwright.Page {
	t.Helper()
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright: %v (run `just install-browsers` first)", err)
	}
	t.Cleanup(func() { pw.Stop() })
	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("launch chromium: %v", err)
	}
	t.Cleanup(func() { browser.Close() })
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	return page
}

// computedStyle resolves one computed-style property on the first element
// matching the locator.
func computedStyle(t *testing.T, l playwright.Locator, prop string) string {
	t.Helper()
	v, err := l.Evaluate(`(el, prop) => getComputedStyle(el)[prop]`, prop)
	if err != nil {
		t.Fatalf("computed style %q: %v", prop, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed style %q: got %T(%v), want string", prop, v, v)
	}
	return s
}

// computedStyleSel returns the computed style property of the first element
// matching selector.
func computedStyleSel(t *testing.T, page playwright.Page, selector, prop string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => { const el = document.querySelector(%q); return el ? getComputedStyle(el)[%q] : "MISSING"; }`,
		selector, prop))
	if err != nil {
		t.Fatalf("computed %s of %q: %v", prop, selector, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed %s of %q: non-string %v", prop, selector, v)
	}
	if s == "MISSING" {
		t.Fatalf("no element matches %q", selector)
	}
	return s
}

// openSheet starts the demo + browser and opens the component sheet.
func openSheet(t *testing.T) playwright.Page {
	t.Helper()
	srv := startDemo(t)
	page := startBrowser(t)
	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}
	return page
}

// style returns one computed-style property of the first element matching
// the selector.
func style(t *testing.T, page playwright.Page, selector, prop string) string {
	t.Helper()
	v, err := page.Evaluate(
		`args => getComputedStyle(document.querySelector(args.sel))[args.prop]`,
		map[string]any{"sel": selector, "prop": prop},
	)
	if err != nil {
		t.Fatalf("computed %s of %q: %v", prop, selector, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed %s of %q: got %T", prop, selector, v)
	}
	return s
}

// wrapStyle resolves the .input-wrap enclosing the given input and returns
// one of its computed-style properties (the border/glow live on the wrap).
func wrapStyle(t *testing.T, page playwright.Page, inputSel, prop string) string {
	t.Helper()
	v, err := page.Evaluate(
		`args => getComputedStyle(document.querySelector(args.sel).closest('.input-wrap'))[args.prop]`,
		map[string]any{"sel": inputSel, "prop": prop},
	)
	if err != nil {
		t.Fatalf("computed wrap %s of %q: %v", prop, inputSel, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed wrap %s of %q: got %T", prop, inputSel, v)
	}
	return s
}

// computed returns one computed-style property of the first element a
// locator-resolved selector matches.
func computed(t *testing.T, page playwright.Page, sel, prop string) string {
	t.Helper()
	v, err := page.Locator(sel).Evaluate("(el, prop) => getComputedStyle(el)[prop]", prop)
	if err != nil {
		t.Fatalf("computed %s of %s: %v", prop, sel, err)
	}
	s, _ := v.(string)
	return s
}

// assertStyle asserts one computed-style property of a locator.
func assertStyle(t *testing.T, l playwright.Locator, prop, want, what string) {
	t.Helper()
	if got := computedStyle(t, l, prop); got != want {
		t.Errorf("%s: %s = %q, want %q", what, prop, got, want)
	}
}

// isChecked reports the checked state of the first element matching sel.
func isChecked(t *testing.T, page playwright.Page, sel string) bool {
	t.Helper()
	v, err := page.Locator(sel).IsChecked()
	if err != nil {
		t.Fatalf("IsChecked(%s): %v", sel, err)
	}
	return v
}

// activeID returns document.activeElement's id ("" when nothing focused).
func activeID(t *testing.T, page playwright.Page) string {
	t.Helper()
	v, err := page.Evaluate("() => document.activeElement && document.activeElement.id")
	if err != nil {
		t.Fatalf("activeElement: %v", err)
	}
	s, _ := v.(string)
	return s
}

// parseColor parses Chromium computed-colour serializations:
// "rgb(R, G, B)", "rgba(R, G, B, A)" and "color(srgb r g b / a)".
// Channels are returned on the 0–255 scale, alpha 0–1.
func parseColor(s string) (r, g, b, a float64, err error) {
	a = 1
	switch {
	case strings.HasPrefix(s, "rgb"):
		body := s[strings.IndexByte(s, '(')+1 : len(s)-1]
		parts := strings.Split(body, ",")
		if len(parts) != 3 && len(parts) != 4 {
			return 0, 0, 0, 0, fmt.Errorf("unexpected rgb form %q", s)
		}
		var vals [4]float64
		vals[3] = 1
		for i, p := range parts {
			if vals[i], err = strconv.ParseFloat(strings.TrimSpace(p), 64); err != nil {
				return 0, 0, 0, 0, fmt.Errorf("parse %q: %w", s, err)
			}
		}
		return vals[0], vals[1], vals[2], vals[3], nil
	case strings.HasPrefix(s, "color(srgb "):
		body := strings.TrimSuffix(strings.TrimPrefix(s, "color(srgb "), ")")
		body = strings.ReplaceAll(body, "/", " ")
		fields := strings.Fields(body)
		if len(fields) != 3 && len(fields) != 4 {
			return 0, 0, 0, 0, fmt.Errorf("unexpected color() form %q", s)
		}
		var vals [4]float64
		vals[3] = 1
		for i, f := range fields {
			if vals[i], err = strconv.ParseFloat(f, 64); err != nil {
				return 0, 0, 0, 0, fmt.Errorf("parse %q: %w", s, err)
			}
		}
		return vals[0] * 255, vals[1] * 255, vals[2] * 255, vals[3], nil
	}
	return 0, 0, 0, 0, fmt.Errorf("unrecognised colour %q", s)
}

// assertToneAlpha asserts a computed colour is the given tone's rgb channels
// at partial alpha — i.e. a color-mix derivation: neither transparent nor
// the solid tone colour.
func assertToneAlpha(t *testing.T, got, rgb, what string) {
	t.Helper()
	wantParts := strings.Split(rgb, ",")
	var want [3]float64
	for i, p := range wantParts {
		want[i], _ = strconv.ParseFloat(strings.TrimSpace(p), 64)
	}
	r, g, b, a, err := parseColor(got)
	if err != nil {
		t.Errorf("%s: %v", what, err)
		return
	}
	const tol = 1.0
	if math.Abs(r-want[0]) > tol || math.Abs(g-want[1]) > tol || math.Abs(b-want[2]) > tol {
		t.Errorf("%s = %q, want tone channels rgb(%s)", what, got, rgb)
	}
	if a <= 0 || a >= 1 {
		t.Errorf("%s = %q, want partial alpha (0 < a < 1): neither transparent nor solid", what, got)
	}
}

// assertOpaque asserts a computed colour is exactly the given rgb channels at
// full alpha (±1 per channel, as assertToneAlpha).
func assertOpaque(t *testing.T, got, rgb, what string) {
	t.Helper()
	wantParts := strings.Split(rgb, ",")
	var want [3]float64
	for i, p := range wantParts {
		want[i], _ = strconv.ParseFloat(strings.TrimSpace(p), 64)
	}
	r, g, b, a, err := parseColor(got)
	if err != nil {
		t.Errorf("%s: %v", what, err)
		return
	}
	const tol = 1.0
	if math.Abs(r-want[0]) > tol || math.Abs(g-want[1]) > tol || math.Abs(b-want[2]) > tol {
		t.Errorf("%s = %q, want rgb(%s)", what, got, rgb)
	}
	if a < 1 {
		t.Errorf("%s = %q, want fully opaque", what, got)
	}
}

// assertTransparent asserts a computed colour has zero alpha.
func assertTransparent(t *testing.T, got, what string) {
	t.Helper()
	_, _, _, a, err := parseColor(got)
	if err != nil {
		t.Errorf("%s: %v", what, err)
		return
	}
	if a != 0 {
		t.Errorf("%s = %q, want transparent (alpha 0)", what, got)
	}
}

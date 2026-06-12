//go:build e2e

package e2e

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// computedStyle returns the computed style property of the first element
// matching selector.
func computedStyle(t *testing.T, page playwright.Page, selector, prop string) string {
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

func TestBadgeComputedStyles(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	// --- t-gruvbox (default root classes) --------------------------------

	// Solid badges fill with the raw tone token.
	if got := computedStyle(t, page, ".badge.bd-solid.tone-err", "backgroundColor"); got != "rgb(251, 73, 52)" {
		t.Errorf("gruvbox solid err background = %q, want rgb(251, 73, 52)", got)
	}
	if got := computedStyle(t, page, ".badge.bd-solid.tone-ok", "backgroundColor"); got != "rgb(184, 187, 38)" {
		t.Errorf("gruvbox solid ok background = %q, want rgb(184, 187, 38)", got)
	}

	// Outline: transparent background + tone-derived (partial-alpha) border.
	if got := computedStyle(t, page, ".badge.bd-outline.tone-err", "backgroundColor"); got != "rgba(0, 0, 0, 0)" {
		t.Errorf("outline err background = %q, want transparent", got)
	}
	assertToneAlpha(t, computedStyle(t, page, ".badge.bd-outline.tone-err", "borderTopColor"),
		"251, 73, 52", "outline err border")

	// Tint: the 15% color-mix — tone channels at partial alpha, so neither
	// transparent nor the solid colour.
	tintBG := computedStyle(t, page, ".badge.bd-tint.tone-err", "backgroundColor")
	if tintBG == "rgba(0, 0, 0, 0)" || tintBG == "rgb(251, 73, 52)" {
		t.Errorf("tint err background = %q, want a tint (neither transparent nor solid)", tintBG)
	}
	assertToneAlpha(t, tintBG, "251, 73, 52", "tint err background")
	assertToneAlpha(t, computedStyle(t, page, ".badge.bd-tint.tone-err", "borderTopColor"),
		"251, 73, 52", "tint err border")

	// In-badge dot is the 5px square (currentColor, no radius).
	if got := computedStyle(t, page, ".badge .badge-dot", "width"); got != "5px" {
		t.Errorf("badge-dot width = %q, want 5px", got)
	}
	if got := computedStyle(t, page, ".badge .badge-dot", "borderRadius"); got != "0px" {
		t.Errorf("badge-dot border-radius = %q, want 0px (square)", got)
	}

	// Status dot: 7px, the one sanctioned circle, tone colour via currentColor.
	if got := computedStyle(t, page, ".dot.tone-err", "width"); got != "7px" {
		t.Errorf("dot width = %q, want 7px", got)
	}
	if got := computedStyle(t, page, ".dot.tone-err", "borderRadius"); got != "50%" {
		t.Errorf("dot border-radius = %q, want 50%%", got)
	}
	if got := computedStyle(t, page, ".dot.tone-err", "backgroundColor"); got != "rgb(251, 73, 52)" {
		t.Errorf("gruvbox err dot background = %q, want rgb(251, 73, 52)", got)
	}

	// --- pulse: animated only under prefers-reduced-motion: no-preference -

	if got := computedStyle(t, page, ".dot.pulse", "animationName"); got != "baud-pulse" {
		t.Errorf("pulse dot animation-name = %q, want baud-pulse", got)
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}); err != nil {
		t.Fatalf("emulate reduced motion: %v", err)
	}
	if got := computedStyle(t, page, ".dot.pulse", "animationName"); got != "none" {
		t.Errorf("pulse dot animates under reduced motion: animation-name = %q, want none", got)
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionNoPreference,
	}); err != nil {
		t.Fatalf("reset reduced motion: %v", err)
	}

	// --- token flow proof: root-class swap to t-mocha ---------------------

	if _, err := page.Evaluate(
		`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if got := computedStyle(t, page, ".badge.bd-solid.tone-err", "backgroundColor"); got != "rgb(243, 139, 168)" {
		t.Errorf("mocha solid err background = %q, want rgb(243, 139, 168)", got)
	}
	assertToneAlpha(t, computedStyle(t, page, ".badge.bd-tint.tone-err", "backgroundColor"),
		"243, 139, 168", "mocha tint err background")
}

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

func TestBadgeComputedStyles(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	// --- t-gruvbox (default root classes) --------------------------------
	// Full options matrix: every Tone × Variant cell (6 × 3 = 18), asserted
	// against the raw t-gruvbox tokens from assets/css/tokens.css.
	tones := []struct {
		name    string
		rgb     string // tone channels — text colour, outline border, tint base
		solidBG string // solid background (neutral solids fill with --fg-faint)
		solidFG string // solid text (accent uses --on-accent, the rest --bg-panel)
	}{
		{"ok", "184, 187, 38", "184, 187, 38", "40, 40, 40"},        // --ok #b8bb26
		{"warn", "250, 189, 47", "250, 189, 47", "40, 40, 40"},      // --warn #fabd2f
		{"err", "251, 73, 52", "251, 73, 52", "40, 40, 40"},         // --err #fb4934
		{"info", "131, 165, 152", "131, 165, 152", "40, 40, 40"},    // --info #83a598
		{"accent", "250, 189, 47", "250, 189, 47", "29, 32, 33"},    // --accent #fabd2f / --on-accent #1d2021
		{"neutral", "168, 153, 132", "124, 111, 100", "40, 40, 40"}, // --fg-muted #a89984 / --fg-faint #7c6f64
	}
	for _, tone := range tones {
		for _, variant := range []string{"solid", "outline", "tint"} {
			t.Run("gruvbox/"+variant+"/"+tone.name, func(t *testing.T) {
				sel := fmt.Sprintf(".badge.bd-%s.tone-%s", variant, tone.name)
				switch variant {
				case "solid":
					// Solid fills with the raw tone token; text flips to the
					// readable panel/on-accent colour.
					assertOpaque(t, computedStyleSel(t, page, sel, "backgroundColor"), tone.solidBG, sel+" background")
					assertOpaque(t, computedStyleSel(t, page, sel, "color"), tone.solidFG, sel+" text")
				case "outline":
					// Outline: transparent background + tone-derived
					// (partial-alpha) border.
					assertTransparent(t, computedStyleSel(t, page, sel, "backgroundColor"), sel+" background")
					assertToneAlpha(t, computedStyleSel(t, page, sel, "borderTopColor"), tone.rgb, sel+" border")
				case "tint":
					if tone.name == "neutral" {
						// Neutral tint is the opaque raised-surface pair, not
						// a color-mix derivation.
						assertOpaque(t, computedStyleSel(t, page, sel, "backgroundColor"), "50, 48, 47", sel+" background") // --bg-raised #32302f
						assertOpaque(t, computedStyleSel(t, page, sel, "borderTopColor"), "80, 73, 69", sel+" border")      // --border-strong #504945
					} else {
						// Tint: the 15%/38% color-mix — tone channels at
						// partial alpha, neither transparent nor solid.
						assertToneAlpha(t, computedStyleSel(t, page, sel, "backgroundColor"), tone.rgb, sel+" background")
						assertToneAlpha(t, computedStyleSel(t, page, sel, "borderTopColor"), tone.rgb, sel+" border")
					}
				}
			})
		}
	}

	// In-badge dot is the 5px square (currentColor, no radius).
	if got := computedStyleSel(t, page, ".badge .badge-dot", "width"); got != "5px" {
		t.Errorf("badge-dot width = %q, want 5px", got)
	}
	if got := computedStyleSel(t, page, ".badge .badge-dot", "borderRadius"); got != "0px" {
		t.Errorf("badge-dot border-radius = %q, want 0px (square)", got)
	}

	// Status dot: 7px, the one sanctioned circle, tone colour via currentColor.
	if got := computedStyleSel(t, page, ".dot.tone-err", "width"); got != "7px" {
		t.Errorf("dot width = %q, want 7px", got)
	}
	if got := computedStyleSel(t, page, ".dot.tone-err", "borderRadius"); got != "50%" {
		t.Errorf("dot border-radius = %q, want 50%%", got)
	}
	if got := computedStyleSel(t, page, ".dot.tone-err", "backgroundColor"); got != "rgb(251, 73, 52)" {
		t.Errorf("gruvbox err dot background = %q, want rgb(251, 73, 52)", got)
	}

	// --- pulse: animated only under prefers-reduced-motion: no-preference -

	if got := computedStyleSel(t, page, ".dot.pulse", "animationName"); got != "baud-pulse" {
		t.Errorf("pulse dot animation-name = %q, want baud-pulse", got)
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}); err != nil {
		t.Fatalf("emulate reduced motion: %v", err)
	}
	if got := computedStyleSel(t, page, ".dot.pulse", "animationName"); got != "none" {
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
	if got := computedStyleSel(t, page, ".badge.bd-solid.tone-err", "backgroundColor"); got != "rgb(243, 139, 168)" {
		t.Errorf("mocha solid err background = %q, want rgb(243, 139, 168)", got)
	}
	assertToneAlpha(t, computedStyleSel(t, page, ".badge.bd-tint.tone-err", "backgroundColor"),
		"243, 139, 168", "mocha tint err background")

	// …and again to t-sollight: same root-class-swap-only mechanism.
	if _, err := page.Evaluate(
		`() => document.body.classList.replace("t-mocha", "t-sollight")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if got := computedStyleSel(t, page, ".badge.bd-solid.tone-err", "backgroundColor"); got != "rgb(220, 50, 47)" {
		t.Errorf("sollight solid err background = %q, want rgb(220, 50, 47)", got)
	}
}

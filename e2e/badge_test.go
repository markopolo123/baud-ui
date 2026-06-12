//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

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

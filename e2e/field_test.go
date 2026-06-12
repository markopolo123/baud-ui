//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

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

func TestFieldInput(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	const (
		gruvAccent = "rgb(250, 189, 47)"  // t-gruvbox --accent
		gruvErr    = "rgb(251, 73, 52)"   // t-gruvbox --err
		gruvFaint  = "rgb(124, 111, 100)" // t-gruvbox --fg-faint
		mochaAcc   = "rgb(137, 180, 250)" // t-mocha --accent
	)

	// Label is uppercase-styled at --fs-sm via .field-label.
	if tt := style(t, page, `label[for="fi-plain"]`, "textTransform"); tt != "uppercase" {
		t.Errorf("field label text-transform = %q, want uppercase", tt)
	}

	// Unfocused input border is --border-strong, not accent.
	if bc := wrapStyle(t, page, "#fi-plain", "borderTopColor"); bc == gruvAccent {
		t.Errorf("unfocused input border already accent (%s)", bc)
	}

	// Click into the plain input: accent border + visible glow ring.
	if err := page.Locator("#fi-plain").Click(); err != nil {
		t.Fatalf("click #fi-plain: %v", err)
	}
	if bc := wrapStyle(t, page, "#fi-plain", "borderTopColor"); bc != gruvAccent {
		t.Errorf("focused input border = %s, want %s (--accent)", bc, gruvAccent)
	}
	if bs := wrapStyle(t, page, "#fi-plain", "boxShadow"); bs == "none" {
		t.Errorf("focused input has no glow ring (box-shadow none)")
	}

	// Tab moves focus to the next field's input; styling follows focus.
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	active, err := page.Evaluate(`() => document.activeElement.id`)
	if err != nil {
		t.Fatalf("activeElement: %v", err)
	}
	if active != "fi-prefix" {
		t.Fatalf("after Tab, active element = %v, want fi-prefix", active)
	}
	if bc := wrapStyle(t, page, "#fi-prefix", "borderTopColor"); bc != gruvAccent {
		t.Errorf("tabbed-into input border = %s, want %s (--accent)", bc, gruvAccent)
	}
	if bs := wrapStyle(t, page, "#fi-prefix", "boxShadow"); bs == "none" {
		t.Errorf("tabbed-into input has no glow ring (box-shadow none)")
	}

	// Error input: --err border; hint line is --err too.
	if bc := wrapStyle(t, page, "#fi-error", "borderTopColor"); bc != gruvErr {
		t.Errorf("error input border = %s, want %s (--err)", bc, gruvErr)
	}
	if c := style(t, page, ".field-hint.err", "color"); c != gruvErr {
		t.Errorf("error hint color = %s, want %s (--err)", c, gruvErr)
	}

	// Affixes render in --fg-faint.
	if c := style(t, page, "#fi-both", "color"); c == gruvFaint {
		t.Errorf("input text itself should not be faint")
	}
	affix, err := page.Evaluate(
		`() => getComputedStyle(document.querySelector('#fi-both').closest('.input-wrap').querySelector('.affix')).color`,
	)
	if err != nil {
		t.Fatalf("affix color: %v", err)
	}
	if affix != gruvFaint {
		t.Errorf("affix color = %v, want %s (--fg-faint)", affix, gruvFaint)
	}

	// Theme swap is a root-class change only: focus accent becomes mocha blue.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if err := page.Locator("#fi-plain").Click(); err != nil {
		t.Fatalf("refocus #fi-plain: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#fi-plain').closest('.input-wrap')).borderTopColor === 'rgb(137, 180, 250)'`, nil,
	); err != nil {
		bc := wrapStyle(t, page, "#fi-plain", "borderTopColor")
		t.Fatalf("t-mocha focus accent not applied: border stayed %s, want %s: %v", bc, mochaAcc, err)
	}
}

//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved colours asserted below (assets/css/tokens.css):
//
//	t-gruvbox --sel rgba(250,189,47,.14), --fg-faint #7c6f64
//	t-mocha   --sel rgba(137,180,250,.13)
const (
	palGruvSelBG  = "rgba(250, 189, 47, 0.14)"
	palMochaSelBG = "rgba(137, 180, 250, 0.13)"
	palGruvFaint  = "rgb(124, 111, 100)"
)

// paletteOpenSheet opens the sheet and waits for body[data-hs-ok] — the
// ParseHealth sentinel proves the whole behaviors file (Palette included)
// parsed and installed before the test sends keys.
func paletteOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	page := openSheet(t)
	if err := page.Locator("body[data-hs-ok]").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("hyperscript never booted (body[data-hs-ok] missing): %v", err)
	}
	return page
}

// paletteAttr reads an attribute of the first match ("" when absent).
func paletteAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => { const el = document.querySelector(%q); return el ? (el.getAttribute(%q) ?? "") : "NO ELEMENT"; }`,
		sel, name))
	if err != nil {
		t.Fatalf("attribute %s of %q: %v", name, sel, err)
	}
	s, _ := v.(string)
	if s == "NO ELEMENT" {
		t.Fatalf("no element matches %q", sel)
	}
	return s
}

// paletteSummon focuses the #pal-anchor opener and presses Ctrl-K, then
// waits until the overlay is open with the input focused and the first
// row highlighted (the open handshake the other helpers depend on).
func paletteSummon(t *testing.T, page playwright.Page) {
	t.Helper()
	if err := page.Locator("#pal-anchor").Focus(); err != nil {
		t.Fatalf("focus #pal-anchor: %v", err)
	}
	if err := page.Keyboard().Press("Control+k"); err != nil {
		t.Fatalf("press Ctrl-K: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pal-live').classList.contains('open')
			&& document.activeElement && document.activeElement.id === 'pal-live-input'
			&& document.querySelector('#pal-live .palette-item.hl') !== null`, nil,
	); err != nil {
		t.Fatalf("Ctrl-K never opened the palette with the input focused: %v", err)
	}
}

// paletteRowCount counts the live palette's option rows.
func paletteRowCount(t *testing.T, page playwright.Page) int {
	t.Helper()
	n, err := page.Locator("#pal-live .palette-item").Count()
	if err != nil {
		t.Fatalf("count palette rows: %v", err)
	}
	return n
}

// TestPaletteOpenTrapDismissRestore: a real Ctrl-K keydown opens the
// overlay (PaletteKey → baud:palette → Palette), focus lands trapped in
// the query input (Tab is swallowed), and both Esc and a backdrop click
// close the overlay and restore focus to the opener.
func TestPaletteOpenTrapDismissRestore(t *testing.T) {
	page := paletteOpenSheet(t)

	if got := computedStyleSel(t, page, "#pal-live", "display"); got != "none" {
		t.Fatalf("overlay display before Ctrl-K = %v, want none", got)
	}
	paletteSummon(t, page)
	if got := computedStyleSel(t, page, "#pal-live", "display"); got != "flex" {
		t.Errorf("open overlay display = %v, want flex", got)
	}
	if got := activeID(t, page); got != "pal-live-input" {
		t.Errorf("focus after open = %q, want pal-live-input", got)
	}
	if got := paletteAttr(t, page, "#pal-live-input", "aria-activedescendant"); got != "pal-live-cmd-0" {
		t.Errorf("aria-activedescendant on open = %q, want pal-live-cmd-0", got)
	}

	// focus is trapped: Tab does not leave the input.
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	if got := activeID(t, page); got != "pal-live-input" {
		t.Errorf("focus after Tab = %q, want pal-live-input (trapped)", got)
	}

	// Esc closes and restores focus to the opener.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#pal-live').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("Esc never closed the palette: %v", err)
	}
	if got := computedStyleSel(t, page, "#pal-live", "display"); got != "none" {
		t.Errorf("overlay display after Esc = %v, want none", got)
	}
	if got := activeID(t, page); got != "pal-anchor" {
		t.Errorf("focus after Esc = %q, want pal-anchor (restored)", got)
	}

	// backdrop click closes too (the overlay corner is outside the panel).
	paletteSummon(t, page)
	if err := page.Locator("#pal-live").Click(playwright.LocatorClickOptions{
		Position: &playwright.Position{X: 8, Y: 8},
	}); err != nil {
		t.Fatalf("backdrop click: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#pal-live').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("backdrop click never closed the palette: %v", err)
	}
	if got := activeID(t, page); got != "pal-anchor" {
		t.Errorf("focus after backdrop click = %q, want pal-anchor (restored)", got)
	}
}

// TestPaletteServerFilterSwap: typing fires the debounced keyup hx-get
// and the PaletteResults fragment swaps in server-filtered (one "deploy"
// hit, re-highlighted row 0); garbage input swaps in the ∅ empty state;
// reopening clears the query and refetches the unfiltered list.
func TestPaletteServerFilterSwap(t *testing.T) {
	page := paletteOpenSheet(t)
	paletteSummon(t, page)

	if n := paletteRowCount(t, page); n != 6 {
		t.Fatalf("seeded row count = %d, want 6", n)
	}
	// PressSequentially, not Keyboard().Type: Type inserts text without
	// key events, so the keyup trigger would never fire.
	if err := page.Locator("#pal-live-input").PressSequentially("deploy"); err != nil {
		t.Fatalf("type deploy: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#pal-live .palette-item').length === 1`, nil,
	); err != nil {
		t.Fatalf("hx-get swap never landed (still %d rows): %v", paletteRowCount(t, page), err)
	}
	label, err := page.Locator("#pal-live .palette-item .pi-label").TextContent()
	if err != nil || label != "deploy canary build" {
		t.Errorf("filtered row label = %q (err=%v), want deploy canary build", label, err)
	}
	cat, err := page.Locator("#pal-live .palette-item .pi-cat").TextContent()
	if err != nil || cat != "fleet" {
		t.Errorf("filtered row category = %q (err=%v), want fleet", cat, err)
	}
	// the swapped-in list is re-highlighted with ARIA back in sync.
	if got := paletteAttr(t, page, "#pal-live-input", "aria-activedescendant"); got != "pal-live-cmd-0" {
		t.Errorf("aria-activedescendant after swap = %q, want pal-live-cmd-0", got)
	}
	if n, _ := page.Locator("#pal-live .palette-item.hl").Count(); n != 1 {
		t.Errorf("highlighted rows after swap = %d, want 1", n)
	}

	// garbage-safe: a query matching nothing swaps in the ∅ empty state.
	if err := page.Locator("#pal-live-input").PressSequentially(`%$#"zz`); err != nil {
		t.Fatalf("type garbage: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pal-live .palette-empty') !== null
			&& document.querySelectorAll('#pal-live .palette-item').length === 0`, nil,
	); err != nil {
		t.Fatalf("empty state never swapped in: %v", err)
	}
	empty, err := page.Locator("#pal-live .palette-empty").TextContent()
	if err != nil || strings.TrimSpace(empty) != "∅ no matches" {
		t.Errorf("empty state text = %q (err=%v), want ∅ no matches", empty, err)
	}
	if got := paletteAttr(t, page, "#pal-live-input", "aria-activedescendant"); got != "" {
		t.Errorf("aria-activedescendant on empty list = %q, want removed", got)
	}

	// reopening clears the stale filter: the replayed keyup refetches all.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	paletteSummon(t, page)
	if v, _ := page.Locator("#pal-live-input").InputValue(); v != "" {
		t.Errorf("query after reopen = %q, want empty", v)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#pal-live .palette-item').length === 6`, nil,
	); err != nil {
		t.Fatalf("reopen never refetched the unfiltered list (%d rows): %v", paletteRowCount(t, page), err)
	}
}

// TestPaletteKeyboardActivate: ↑↓ move the --sel + accent-inset highlight
// and aria-activedescendant over the row list, ↵ on an action row runs
// its visible hyperscript and closes the palette, ↵ on an anchor row
// navigates.
func TestPaletteKeyboardActivate(t *testing.T) {
	page := paletteOpenSheet(t)
	paletteSummon(t, page)

	if err := page.Keyboard().Press("ArrowDown"); err != nil {
		t.Fatalf("press ArrowDown: %v", err)
	}
	if got := paletteAttr(t, page, "#pal-live-input", "aria-activedescendant"); got != "pal-live-cmd-1" {
		t.Errorf("aria-activedescendant after ↓ = %q, want pal-live-cmd-1", got)
	}
	if got := computedStyleSel(t, page, "#pal-live-cmd-1", "backgroundColor"); got != palGruvSelBG {
		t.Errorf("highlighted row background = %v, want %v (--sel)", got, palGruvSelBG)
	}
	shadow := computedStyleSel(t, page, "#pal-live-cmd-1", "boxShadow")
	if !strings.Contains(shadow, "inset") || !strings.Contains(shadow, gruvAccent) || !strings.Contains(shadow, "2px") {
		t.Errorf("highlighted row box-shadow = %q, want 2px inset %s", shadow, gruvAccent)
	}
	if n, _ := page.Locator("#pal-live-cmd-0.hl").Count(); n != 0 {
		t.Error("row 0 kept .hl after moving away")
	}
	if got := paletteAttr(t, page, "#pal-live-cmd-1", "aria-selected"); got != "true" {
		t.Errorf("highlighted row aria-selected = %q, want true", got)
	}

	// ↑ back, then ↓↓ to the action row (pal-live-cmd-2, deploy canary).
	for _, key := range []string{"ArrowUp", "ArrowDown", "ArrowDown"} {
		if err := page.Keyboard().Press(key); err != nil {
			t.Fatalf("press %s: %v", key, err)
		}
	}
	if got := paletteAttr(t, page, "#pal-live-input", "aria-activedescendant"); got != "pal-live-cmd-2" {
		t.Fatalf("aria-activedescendant after ↑↓↓ = %q, want pal-live-cmd-2", got)
	}

	// ↵ runs the row's hyperscript action — visibly — and closes.
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pal-action-out').textContent === 'canary deployed'`, nil,
	); err != nil {
		out, _ := page.Locator("#pal-action-out").TextContent()
		t.Fatalf("action never ran (#pal-action-out = %q): %v", out, err)
	}
	if got := computedStyleSel(t, page, "#pal-live", "display"); got != "none" {
		t.Errorf("overlay display after action ↵ = %v, want none", got)
	}
	if got := activeID(t, page); got != "pal-anchor" {
		t.Errorf("focus after action ↵ = %q, want pal-anchor (restored)", got)
	}

	// ↵ on the initial highlight (row 0, an anchor) navigates to "/".
	base := strings.TrimSuffix(page.URL(), "/sheet")
	paletteSummon(t, page)
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter on anchor row: %v", err)
	}
	if err := page.WaitForURL(base + "/"); err != nil {
		t.Fatalf("anchor row never navigated (url = %s): %v", page.URL(), err)
	}
}

// TestPaletteStaticStylingThemes: the statically-open sheet palette pins
// the design CSS — accent panel border, accent › prompt, faint UPPERCASE
// 64px category column, --sel + accent inset highlight, zero radius —
// and a t-mocha root-class swap re-resolves every token.
func TestPaletteStaticStylingThemes(t *testing.T) {
	page := paletteOpenSheet(t)
	panel := "#pal-static .palette"
	hl := "#pal-static .palette-item.hl"

	if got := computedStyleSel(t, page, panel, "borderTopColor"); got != gruvAccent {
		t.Errorf("palette border = %v, want %v (--accent)", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, panel, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("palette border-radius = %v, want 0px", got)
	}
	if got := computedStyleSel(t, page, panel, "width"); got != "560px" {
		t.Errorf("palette width = %v, want 560px", got)
	}
	if got := computedStyleSel(t, page, "#pal-static .prompt", "color"); got != gruvAccent {
		t.Errorf("prompt color = %v, want %v (--accent)", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, hl, "backgroundColor"); got != palGruvSelBG {
		t.Errorf("hl row background = %v, want %v (--sel)", got, palGruvSelBG)
	}
	shadow := computedStyleSel(t, page, hl, "boxShadow")
	if !strings.Contains(shadow, "inset") || !strings.Contains(shadow, gruvAccent) || !strings.Contains(shadow, "2px") {
		t.Errorf("hl row box-shadow = %q, want 2px inset %s", shadow, gruvAccent)
	}
	cat := "#pal-static .palette-item .pi-cat"
	if got := computedStyleSel(t, page, cat, "color"); got != palGruvFaint {
		t.Errorf("category color = %v, want %v (--fg-faint)", got, palGruvFaint)
	}
	if got := computedStyleSel(t, page, cat, "textTransform"); got != "uppercase" {
		t.Errorf("category text-transform = %v, want uppercase", got)
	}
	if got := computedStyleSel(t, page, cat, "width"); got != "64px" {
		t.Errorf("category column width = %v, want 64px", got)
	}

	// theme swap is a root-class change only: every token re-resolves.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.body).backgroundColor === 'rgb(17, 17, 27)'`, nil,
	); err != nil {
		t.Fatalf("t-mocha never applied: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "borderTopColor"); got != mochaAccent {
		t.Errorf("mocha palette border = %v, want %v (--accent)", got, mochaAccent)
	}
	if got := computedStyleSel(t, page, hl, "backgroundColor"); got != palMochaSelBG {
		t.Errorf("mocha hl row background = %v, want %v (--sel)", got, palMochaSelBG)
	}
	mochaShadow := computedStyleSel(t, page, hl, "boxShadow")
	if !strings.Contains(mochaShadow, mochaAccent) {
		t.Errorf("mocha hl inset = %q, want %s", mochaShadow, mochaAccent)
	}
}

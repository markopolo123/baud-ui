//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved --fg-faint (assets/css/tokens.css) for the chip key span:
//
//	t-gruvbox #7c6f64 → rgb(124, 111, 100); t-mocha #6c7086 → rgb(108, 112, 134)
const (
	taginputGruvFaint  = "rgb(124, 111, 100)"
	taginputMochaFaint = "rgb(108, 112, 134)"
)

// Accent channels for the chip-tint color-mix assertions (assertToneAlpha).
const (
	taginputGruvAccentRGB  = "250, 189, 47"  // #fabd2f
	taginputMochaAccentRGB = "137, 180, 250" // #89b4fa
)

// taginputOpenSheet opens /sheet and waits for hyperscript to boot. The
// Panes behavior setting the grid template is the boot signal — hyperscript
// initializes the whole document in one pass, so TagInput is installed too.
func taginputOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	srv := startDemo(t)
	page := startBrowser(t)
	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector('[data-panes]');
			return el && getComputedStyle(el).gridTemplateColumns !== 'none';
		}`, nil,
	); err != nil {
		t.Fatalf("hyperscript never booted (Panes grid template unset): %v", err)
	}
	return page
}

// taginputCount counts elements matching a selector.
func taginputCount(t *testing.T, page playwright.Page, sel string) int {
	t.Helper()
	n, err := page.Locator(sel).Count()
	if err != nil {
		t.Fatalf("count %q: %v", sel, err)
	}
	return n
}

// taginputWait polls a JS predicate (hyperscript handlers run async after
// the triggering DOM event, so state assertions poll instead of racing).
func taginputWait(t *testing.T, page playwright.Page, what, expr string) {
	t.Helper()
	if _, err := page.WaitForFunction(expr, nil, playwright.PageWaitForFunctionOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

// taginputMenuVisible reports whether the suggestion menu of the tags root
// holding the given input id is actually painted (display not none).
func taginputMenuVisible(t *testing.T, page playwright.Page, inputID string) bool {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => {
			const m = document.querySelector('.tags:has(#%s) .tags-menu');
			return m ? getComputedStyle(m).display !== 'none' : false;
		}`, inputID))
	if err != nil {
		t.Fatalf("menu visibility of #%s: %v", inputID, err)
	}
	b, _ := v.(bool)
	return b
}

// TestTagInputEnterAddsChip: typing text and pressing Enter makes a chip
// AND its hidden form input (asserted via DOM), clears the text input, and
// an empty input ignores Enter. The wrap paints the accent focus ring.
func TestTagInputEnterAddsChip(t *testing.T) {
	page := taginputOpenSheet(t)
	root := `.tags:has(#ti-empty)`

	if n := taginputCount(t, page, root+` .tag-chip`); n != 0 {
		t.Fatalf("empty tag input starts with %d chips, want 0", n)
	}
	if err := page.Locator("#ti-empty").Click(); err != nil {
		t.Fatalf("click tags input: %v", err)
	}

	// focus ring per Input conventions: accent border on the wrap
	if got := computedStyleSel(t, page, root+` .tags-wrap`, "borderTopColor"); got != gruvAccent {
		t.Errorf("focused tags-wrap border = %v, want %v (--accent)", got, gruvAccent)
	}

	if err := page.Keyboard().Type("team=infra"); err != nil {
		t.Fatalf("type tag text: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}

	// the chip AND its hidden form input exist — this is the form contract
	taginputWait(t, page, "Enter never produced the chip's hidden input",
		`() => !!document.querySelector('.tags:has(#ti-empty) .tag-chip input[type=hidden][name="filters"][value="team=infra"]')`)
	if n := taginputCount(t, page, root+` .tag-chip`); n != 1 {
		t.Errorf("after Enter: %d chips, want 1", n)
	}

	// chip splits key=value: faint key span + value span
	if got, err := page.Locator(root + ` .tag-chip .tag-k`).TextContent(); err != nil || got != "team=" {
		t.Errorf("chip key span = %q (err=%v), want team=", got, err)
	}
	if got, err := page.Locator(root + ` .tag-chip .tag-v`).TextContent(); err != nil || got != "infra" {
		t.Errorf("chip value span = %q (err=%v), want infra", got, err)
	}

	// input cleared, and Enter on the now-empty input adds nothing
	if v, err := page.Locator("#ti-empty").InputValue(); err != nil || v != "" {
		t.Errorf("input value after Enter = %q (err=%v), want empty", v, err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter on empty input: %v", err)
	}
	if n := taginputCount(t, page, root+` .tag-chip`); n != 1 {
		t.Errorf("Enter on empty input changed chip count to %d, want 1", n)
	}

	// duplicate guard: typing the already-added tag and pressing Enter must
	// not make a second chip or hidden input. The add handler clears the
	// text input before its dup check, so the cleared value proves it ran.
	if err := page.Keyboard().Type("team=infra"); err != nil {
		t.Fatalf("type duplicate tag text: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter on duplicate: %v", err)
	}
	taginputWait(t, page, "duplicate Enter never reached the add handler (input not cleared)",
		`() => document.querySelector('#ti-empty').value === ''`)
	if n := taginputCount(t, page, root+` .tag-chip`); n != 1 {
		t.Errorf("duplicate Enter made a second chip: %d chips, want 1", n)
	}
	if n := taginputCount(t, page, root+` input[type=hidden][name="filters"][value="team=infra"]`); n != 1 {
		t.Errorf("duplicate Enter duplicated the hidden input: %d, want 1", n)
	}
}

// TestTagInputBackspacePopsLastChip: Backspace in an EMPTY input removes the
// last chip and its hidden input; with text in the input it edits text only.
func TestTagInputBackspacePopsLastChip(t *testing.T) {
	page := taginputOpenSheet(t)
	root := `.tags:has(#ti-labels)`

	if n := taginputCount(t, page, root+` .tag-chip`); n != 2 {
		t.Fatalf("labels tag input starts with %d chips, want 2", n)
	}
	if err := page.Locator("#ti-labels").Click(); err != nil {
		t.Fatalf("click tags input: %v", err)
	}

	// guard: Backspace while the input holds text must NOT pop a chip
	if err := page.Keyboard().Type("x"); err != nil {
		t.Fatalf("type guard char: %v", err)
	}
	if err := page.Keyboard().Press("Backspace"); err != nil {
		t.Fatalf("press Backspace (text present): %v", err)
	}
	if n := taginputCount(t, page, root+` .tag-chip`); n != 2 {
		t.Fatalf("Backspace with text present popped a chip: %d chips, want 2", n)
	}

	// now empty: Backspace pops the LAST chip (region=eu-west), env=prod stays
	if err := page.Keyboard().Press("Backspace"); err != nil {
		t.Fatalf("press Backspace (empty): %v", err)
	}
	taginputWait(t, page, "Backspace never removed the last chip",
		`() => document.querySelectorAll('.tags:has(#ti-labels) .tag-chip').length === 1`)
	if n := taginputCount(t, page, root+` input[type=hidden][value="region=eu-west"]`); n != 0 {
		t.Error("popped chip's hidden input still present — form would submit a removed tag")
	}
	if n := taginputCount(t, page, root+` input[type=hidden][name="labels"][value="env=prod"]`); n != 1 {
		t.Errorf("surviving chip's hidden input count = %d, want 1", n)
	}

	// pop the remaining chip; a further Backspace on no chips is a no-op
	if err := page.Keyboard().Press("Backspace"); err != nil {
		t.Fatalf("press Backspace (last chip): %v", err)
	}
	taginputWait(t, page, "Backspace never emptied the chip row",
		`() => document.querySelectorAll('.tags:has(#ti-labels) .tag-chip').length === 0`)
	if err := page.Keyboard().Press("Backspace"); err != nil {
		t.Fatalf("press Backspace (no chips): %v", err)
	}
	if n := taginputCount(t, page, root+` input[type=hidden]`); n != 0 {
		t.Errorf("hidden inputs after emptying = %d, want 0", n)
	}
}

// TestTagInputRemoveButton: clicking a chip's ✕ removes the chip and its
// hidden input.
func TestTagInputRemoveButton(t *testing.T) {
	page := taginputOpenSheet(t)
	root := `.tags:has(#ti-bare)`

	if n := taginputCount(t, page, root+` input[type=hidden][name="tags"][value="canary"]`); n != 1 {
		t.Fatalf("bare tag input hidden-input count = %d, want 1", n)
	}
	if err := page.Locator(root + ` .tag-chip .x-btn`).Click(); err != nil {
		t.Fatalf("click chip remove button: %v", err)
	}
	taginputWait(t, page, "✕ never removed the chip",
		`() => document.querySelectorAll('.tags:has(#ti-bare) .tag-chip').length === 0`)
	if n := taginputCount(t, page, root+` input[type=hidden]`); n != 0 {
		t.Error("removed chip's hidden input still present — form would submit a removed tag")
	}
}

// TestTagInputSuggestionsFilter: the menu opens on focus, dismisses on Esc
// and outside click (MenuDismiss), filters as you type, hides already-added
// tags, and clicking a suggestion makes a chip with its hidden input.
func TestTagInputSuggestionsFilter(t *testing.T) {
	page := taginputOpenSheet(t)
	root := `.tags:has(#ti-empty)`

	if taginputMenuVisible(t, page, "ti-empty") {
		t.Fatal("suggestion menu visible before any interaction")
	}
	if err := page.Locator("#ti-empty").Click(); err != nil {
		t.Fatalf("click tags input: %v", err)
	}
	taginputWait(t, page, "menu never opened on focus",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-empty) .tags-menu')).display !== 'none'`)

	// Esc dismisses (MenuDismiss)
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	taginputWait(t, page, "Escape never dismissed the menu",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-empty) .tags-menu')).display === 'none'`)

	// outside click dismisses too
	if err := page.Locator("#ti-empty").Click(); err != nil {
		t.Fatalf("re-click tags input: %v", err)
	}
	taginputWait(t, page, "menu never re-opened",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-empty) .tags-menu')).display !== 'none'`)
	if err := page.Mouse().Click(5, 5); err != nil {
		t.Fatalf("click outside: %v", err)
	}
	taginputWait(t, page, "outside click never dismissed the menu",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-empty) .tags-menu')).display === 'none'`)

	// typing filters: "err" leaves status=err, hides status=ok
	if err := page.Locator("#ti-empty").Click(); err != nil {
		t.Fatalf("focus tags input: %v", err)
	}
	if err := page.Keyboard().Type("err"); err != nil {
		t.Fatalf("type filter text: %v", err)
	}
	taginputWait(t, page, "filtering never hid the non-matching suggestion",
		`() => document.querySelector('.tags:has(#ti-empty) .menu-item[data-tag="status=ok"]').hidden === true`)
	taginputWait(t, page, "matching suggestion got hidden by the filter",
		`() => document.querySelector('.tags:has(#ti-empty) .menu-item[data-tag="status=err"]').hidden === false`)

	// clicking the suggestion adds the chip + hidden input, then the added
	// tag disappears from the menu (dup filter) and the other returns
	// (query was cleared by the add)
	if err := page.Locator(root + ` .menu-item[data-tag="status=err"]`).Click(); err != nil {
		t.Fatalf("click suggestion: %v", err)
	}
	taginputWait(t, page, "suggestion click never produced the chip's hidden input",
		`() => !!document.querySelector('.tags:has(#ti-empty) .tag-chip input[type=hidden][name="filters"][value="status=err"]')`)
	taginputWait(t, page, "added tag still offered in the menu",
		`() => document.querySelector('.tags:has(#ti-empty) .menu-item[data-tag="status=err"]').hidden === true`)
	taginputWait(t, page, "cleared query never restored the other suggestion",
		`() => document.querySelector('.tags:has(#ti-empty) .menu-item[data-tag="status=ok"]').hidden === false`)

	// adding the last remaining suggestion empties the menu — it stops
	// painting even while open
	if err := page.Locator(root + ` .menu-item[data-tag="status=ok"]`).Click(); err != nil {
		t.Fatalf("click last suggestion: %v", err)
	}
	taginputWait(t, page, "exhausted menu still painting",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-empty) .tags-menu')).display === 'none'`)
	if n := taginputCount(t, page, root+` .tag-chip`); n != 2 {
		t.Errorf("chips after adding both suggestions = %d, want 2", n)
	}
}

// TestTagInputChipTintThemeFlow: the chip background/border are the accent
// color-mix tint (accent channels at 0<alpha<1), the key span is --fg-faint,
// and both re-resolve after a t-mocha root-class swap.
func TestTagInputChipTintThemeFlow(t *testing.T) {
	page := taginputOpenSheet(t)
	chip := `.tags:has(#ti-labels) .tag-chip`

	assertToneAlpha(t, computedStyleSel(t, page, chip, "backgroundColor"),
		taginputGruvAccentRGB, "gruvbox chip background")
	assertToneAlpha(t, computedStyleSel(t, page, chip, "borderTopColor"),
		taginputGruvAccentRGB, "gruvbox chip border")
	if got := computedStyleSel(t, page, chip+` .tag-k`, "color"); got != taginputGruvFaint {
		t.Errorf("gruvbox chip key span color = %v, want %v (--fg-faint)", got, taginputGruvFaint)
	}

	// token flow: theme is a root-class swap via the tweaks panel
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	taginputWait(t, page, "t-mocha --fg-faint never reached the chip key span",
		`() => getComputedStyle(document.querySelector('.tags:has(#ti-labels) .tag-chip .tag-k')).color === '`+taginputMochaFaint+`'`)
	assertToneAlpha(t, computedStyleSel(t, page, chip, "backgroundColor"),
		taginputMochaAccentRGB, "mocha chip background")
	assertToneAlpha(t, computedStyleSel(t, page, chip, "borderTopColor"),
		taginputMochaAccentRGB, "mocha chip border")
}

//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved colours asserted below (from assets/css/tokens.css):
//
//	t-gruvbox --accent #fabd2f → rgb(250, 189, 47), --on-accent #1d2021 → rgb(29, 32, 33)
//	t-mocha   --accent #89b4fa → rgb(137, 180, 250), --on-accent #11111b → rgb(17, 17, 27)
const (
	gruvAccent   = "rgb(250, 189, 47)"
	gruvOnAccent = "rgb(29, 32, 33)"
	mochaAccent  = "rgb(137, 180, 250)"
	mochaOnAcc   = "rgb(17, 17, 27)"
)

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

func isChecked(t *testing.T, page playwright.Page, sel string) bool {
	t.Helper()
	v, err := page.Locator(sel).IsChecked()
	if err != nil {
		t.Fatalf("IsChecked(%s): %v", sel, err)
	}
	return v
}

func computed(t *testing.T, page playwright.Page, sel, prop string) string {
	t.Helper()
	v, err := page.Locator(sel).Evaluate("(el, prop) => getComputedStyle(el)[prop]", prop)
	if err != nil {
		t.Fatalf("computed %s of %s: %v", prop, sel, err)
	}
	s, _ := v.(string)
	return s
}

func activeID(t *testing.T, page playwright.Page) string {
	t.Helper()
	v, err := page.Evaluate("() => document.activeElement && document.activeElement.id")
	if err != nil {
		t.Fatalf("activeElement: %v", err)
	}
	s, _ := v.(string)
	return s
}

// TestCheckboxClickAndGlyph: clicking the wrapping label toggles the real
// input and the glyph turns accent (t-gruvbox), then mocha accent after a
// root-class theme swap.
func TestCheckboxClickAndGlyph(t *testing.T) {
	page := openSheet(t)
	box := `label.cbx:has(#chk-telemetry) .cbx-box`

	if isChecked(t, page, "#chk-telemetry") {
		t.Fatal("telemetry checkbox should start unchecked")
	}
	if got := computed(t, page, box, "color"); got == gruvAccent {
		t.Errorf("unchecked glyph already accent: %v", got)
	}

	if err := page.Locator(`label.cbx:has(#chk-telemetry)`).Click(); err != nil {
		t.Fatalf("click checkbox label: %v", err)
	}
	if !isChecked(t, page, "#chk-telemetry") {
		t.Fatal("label click did not check the real input")
	}
	if got := computed(t, page, box, "color"); got != gruvAccent {
		t.Errorf("checked glyph color = %v, want %v (t-gruvbox --accent)", got, gruvAccent)
	}

	// theme swap is a root-class swap via the tweaks panel
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('label.cbx:has(#chk-telemetry) .cbx-box')).color === '`+mochaAccent+`'`, nil,
	); err != nil {
		got := computed(t, page, box, "color")
		t.Errorf("after t-mocha swap glyph color = %v, want %v: %v", got, mochaAccent, err)
	}

	// disabled checkbox ignores label clicks (Force skips playwright's
	// enabled-target wait — the click really lands, the input ignores it)
	if err := page.Locator(`label.cbx:has(#chk-legacy)`).Click(playwright.LocatorClickOptions{
		Force: playwright.Bool(true),
	}); err != nil {
		t.Fatalf("click disabled checkbox label: %v", err)
	}
	if isChecked(t, page, "#chk-legacy") {
		t.Error("disabled checkbox toggled on label click")
	}
}

// TestCheckboxKeyboard: Tab reaches the next checkbox (disabled ones are
// skipped from tab order by the browser), the focus ring lands on the glyph,
// and Space toggles.
func TestCheckboxKeyboard(t *testing.T) {
	page := openSheet(t)

	if err := page.Locator("#chk-telemetry").Focus(); err != nil {
		t.Fatalf("focus #chk-telemetry: %v", err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	if got := activeID(t, page); got != "chk-restart" {
		t.Fatalf("Tab landed on %q, want chk-restart", got)
	}

	// focus-visible ring relocates onto the glyph of the focused control
	ring := computed(t, page, `label.cbx:has(#chk-restart) .cbx-box`, "outlineColor")
	if ring != gruvAccent {
		t.Errorf("focused glyph outline color = %v, want %v", ring, gruvAccent)
	}

	if !isChecked(t, page, "#chk-restart") {
		t.Fatal("auto-restart should start checked")
	}
	if err := page.Keyboard().Press("Space"); err != nil {
		t.Fatalf("press Space: %v", err)
	}
	if isChecked(t, page, "#chk-restart") {
		t.Error("Space did not toggle the focused checkbox off")
	}
}

// TestRadioArrows: arrow keys move the selection across the named group —
// native radio behaviour, no client logic shipped.
func TestRadioArrows(t *testing.T) {
	page := openSheet(t)

	if !isChecked(t, page, "#rad-eu") {
		t.Fatal("eu-west should start selected")
	}
	if err := page.Locator("#rad-eu").Focus(); err != nil {
		t.Fatalf("focus #rad-eu: %v", err)
	}
	if err := page.Keyboard().Press("ArrowDown"); err != nil {
		t.Fatalf("press ArrowDown: %v", err)
	}
	if !isChecked(t, page, "#rad-us") {
		t.Error("ArrowDown did not move selection to us-east")
	}
	if err := page.Keyboard().Press("ArrowRight"); err != nil {
		t.Fatalf("press ArrowRight: %v", err)
	}
	if !isChecked(t, page, "#rad-ap") {
		t.Error("ArrowRight did not move selection to ap-south")
	}
	if isChecked(t, page, "#rad-eu") {
		t.Error("radio group has more than one checked member")
	}

	// selected radio glyph is accent
	if got := computed(t, page, `label.cbx:has(#rad-ap) .cbx-box`, "color"); got != gruvAccent {
		t.Errorf("selected radio glyph color = %v, want %v", got, gruvAccent)
	}
}

// TestToggleSegments: the selected segment fills accent with on-accent text,
// arrow keys move the selection (radio group underneath), and the fill
// follows a theme swap.
func TestToggleSegments(t *testing.T) {
	page := openSheet(t)
	selected := `.tg-opt:has(.tg-input:checked)`

	if got := computed(t, page, selected, "backgroundColor"); got != gruvAccent {
		t.Errorf("selected segment background = %v, want %v", got, gruvAccent)
	}
	if got := computed(t, page, selected, "color"); got != gruvOnAccent {
		t.Errorf("selected segment text color = %v, want %v (--on-accent)", got, gruvOnAccent)
	}
	txt, err := page.Locator(selected).TextContent()
	if err != nil || strings.TrimSpace(txt) != "table" {
		t.Errorf("selected segment text = %q (err=%v), want table", txt, err)
	}

	// keyboard: ArrowRight moves the underlying radio selection
	if err := page.Locator(`.tg input[value="table"]`).Focus(); err != nil {
		t.Fatalf("focus toggle radio: %v", err)
	}
	if err := page.Keyboard().Press("ArrowRight"); err != nil {
		t.Fatalf("press ArrowRight: %v", err)
	}
	if !isChecked(t, page, `.tg input[value="json"]`) {
		t.Fatal("ArrowRight did not select the json segment")
	}
	if txt, _ := page.Locator(selected).TextContent(); strings.TrimSpace(txt) != "json" {
		t.Errorf("accent fill did not follow selection, selected segment text = %q", txt)
	}

	// clicking a segment selects it too
	if err := page.Locator(`.tg-opt:has(input[value="raw"])`).Click(); err != nil {
		t.Fatalf("click raw segment: %v", err)
	}
	if !isChecked(t, page, `.tg input[value="raw"]`) {
		t.Error("clicking a segment did not check its radio")
	}

	// mocha swap: accent fill + on-accent text re-resolve from root class only
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('.tg-opt:has(.tg-input:checked)')).backgroundColor === '`+mochaAccent+`'`, nil,
	); err != nil {
		got := computed(t, page, selected, "backgroundColor")
		t.Errorf("after t-mocha swap segment background = %v, want %v: %v", got, mochaAccent, err)
	}
	if got := computed(t, page, selected, "color"); got != mochaOnAcc {
		t.Errorf("after t-mocha swap segment text = %v, want %v", got, mochaOnAcc)
	}
}

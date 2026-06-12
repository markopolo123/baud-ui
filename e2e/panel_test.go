//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// panelEvalNum evaluates a page expression that returns a number.
// (panel-prefixed per the shared-helper rule.)
func panelEvalNum(t *testing.T, page playwright.Page, expr string) float64 {
	t.Helper()
	v, err := page.Evaluate(expr)
	if err != nil {
		t.Fatalf("evaluate %s: %v", expr, err)
	}
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	default:
		t.Fatalf("evaluate %s: got %T(%v), want number", expr, v, v)
		return 0
	}
}

// panelSwapClass swaps one root mode class on <body> — the only sanctioned
// theme/density/border switching mechanism.
func panelSwapClass(t *testing.T, page playwright.Page, from, to string) {
	t.Helper()
	if _, err := page.Evaluate(
		`([from, to]) => document.body.classList.replace(from, to)`,
		[]string{from, to}); err != nil {
		t.Fatalf("swap root class %s -> %s: %v", from, to, err)
	}
}

// TestPanelStatusbarToolbar asserts the structure trio's computed visuals
// in t-gruvbox + d-dense + b-line, the b-shade / b-ascii border-mode swaps,
// a density swap, real body scrolling, the statusbar mode/spring cells in
// the shell footer slot, and the t-mocha token-flow proof.
func TestPanelStatusbarToolbar(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	// --- panel header: height --rh (24px at d-dense), hairline bottom ----
	hd := page.Locator("#panel-plain > .panel-hd")
	if got := computedStyle(t, hd, "height"); got != "24px" {
		t.Errorf("panel header height = %q, want 24px (--rh at d-dense)", got)
	}
	if got := computedStyle(t, hd, "borderBottomWidth"); got != "1px" {
		t.Errorf("panel header border-bottom-width = %q, want 1px", got)
	}
	if got := computedStyleSel(t, page, "#panel-plain", "backgroundColor"); got != "rgb(40, 40, 40)" {
		t.Errorf("panel background = %q, want rgb(40, 40, 40) (--bg-panel)", got)
	}

	// --- title: --fg-muted, UPPERCASE, --fs-sm --------------------------
	title := page.Locator("#panel-plain .panel-title")
	if got := computedStyle(t, title, "color"); got != "rgb(168, 153, 132)" {
		t.Errorf("panel title color = %q, want rgb(168, 153, 132) (--fg-muted)", got)
	}
	if got := computedStyle(t, title, "textTransform"); got != "uppercase" {
		t.Errorf("panel title text-transform = %q, want uppercase", got)
	}
	if got := computedStyle(t, title, "fontSize"); got != "10.5px" {
		t.Errorf("panel title font-size = %q, want 10.5px (--fs-sm at d-dense)", got)
	}

	// --- body actually scrolls ------------------------------------------
	if got := computedStyleSel(t, page, "#panel-log > .panel-bd", "overflowY"); got != "auto" {
		t.Errorf("panel body overflow-y = %q, want auto", got)
	}
	overflows := panelEvalNum(t, page,
		`() => { const el = document.querySelector('#panel-log > .panel-bd'); return el.scrollHeight - el.clientHeight; }`)
	if overflows <= 0 {
		t.Errorf("panel body scrollHeight - clientHeight = %v, want > 0 (content must overflow)", overflows)
	}
	scrolled := panelEvalNum(t, page,
		`() => { const el = document.querySelector('#panel-log > .panel-bd'); el.scrollTop = 99999; return el.scrollTop; }`)
	if scrolled <= 0 {
		t.Errorf("panel body scrollTop after scrolling = %v, want > 0 (body must scroll)", scrolled)
	}

	// --- statusbar in the shell footer slot ------------------------------
	// Mode cell: accent bg, on-accent bold text (vim style).
	mode := page.Locator("[data-statusbar] .statusbar > .sb-mode")
	if got := computedStyle(t, mode, "backgroundColor"); got != "rgb(250, 189, 47)" {
		t.Errorf("mode cell background = %q, want rgb(250, 189, 47) (--accent)", got)
	}
	if got := computedStyle(t, mode, "color"); got != "rgb(29, 32, 33)" {
		t.Errorf("mode cell color = %q, want rgb(29, 32, 33) (--on-accent)", got)
	}
	if got := computedStyle(t, mode, "fontWeight"); got != "700" {
		t.Errorf("mode cell font-weight = %q, want 700", got)
	}
	// Hairline separators between cells; none after the last cell.
	if got := computedStyleSel(t, page, "[data-statusbar] .statusbar > .sb-cell:first-child", "borderRightWidth"); got != "1px" {
		t.Errorf("statusbar cell separator width = %q, want 1px hairline", got)
	}
	if got := computedStyleSel(t, page, "[data-statusbar] .statusbar > .sb-cell:last-child", "borderRightWidth"); got != "0px" {
		t.Errorf("last statusbar cell separator width = %q, want 0px", got)
	}
	// The spring cell flexes: wider than every sibling combined.
	springW := panelEvalNum(t, page,
		`() => document.querySelector('[data-statusbar] .statusbar > .sb-spring').offsetWidth`)
	siblingsW := panelEvalNum(t, page,
		`() => [...document.querySelectorAll('[data-statusbar] .statusbar > .sb-cell:not(.sb-spring)')]
			.reduce((sum, el) => sum + el.offsetWidth, 0)`)
	if springW <= siblingsW {
		t.Errorf("spring cell offsetWidth = %v, want > %v (sum of fixed siblings)", springW, siblingsW)
	}

	// --- toolbar composition gap -----------------------------------------
	if got := computedStyleSel(t, page, "#toolbar-demo .toolbar", "display"); got != "flex" {
		t.Errorf("toolbar display = %q, want flex", got)
	}
	if got := computedStyleSel(t, page, "#toolbar-demo .toolbar", "gap"); got != "6px" {
		t.Errorf("toolbar gap = %q, want 6px (--gap at d-dense)", got)
	}

	// --- b-shade: borders go transparent, raised header fill -------------
	panelSwapClass(t, page, "b-line", "b-shade")
	if got := computedStyleSel(t, page, "#panel-plain", "borderTopColor"); got != "rgba(0, 0, 0, 0)" {
		t.Errorf("b-shade panel border = %q, want transparent", got)
	}
	if got := computedStyle(t, hd, "borderBottomColor"); got != "rgba(0, 0, 0, 0)" {
		t.Errorf("b-shade panel header border = %q, want transparent", got)
	}
	if got := computedStyle(t, hd, "backgroundColor"); got != "rgb(50, 48, 47)" {
		t.Errorf("b-shade panel header background = %q, want rgb(50, 48, 47) (--bg-raised separation)", got)
	}

	// --- b-ascii: dashed border + ┌─ … ─┐ title wrap ----------------------
	panelSwapClass(t, page, "b-shade", "b-ascii")
	if got := computedStyleSel(t, page, "#panel-plain", "borderTopStyle"); got != "dashed" {
		t.Errorf("b-ascii panel border-style = %q, want dashed", got)
	}
	before, err := page.Evaluate(
		`() => getComputedStyle(document.querySelector('#panel-plain .panel-title'), '::before').content`)
	if err != nil {
		t.Fatalf("panel-title ::before content: %v", err)
	}
	if s, _ := before.(string); !strings.Contains(s, "┌─") {
		t.Errorf("b-ascii title ::before content = %v, want the ┌─ glyph wrap", before)
	}
	panelSwapClass(t, page, "b-ascii", "b-line")
	none, err := page.Evaluate(
		`() => getComputedStyle(document.querySelector('#panel-plain .panel-title'), '::before').content`)
	if err != nil {
		t.Fatalf("panel-title ::before content: %v", err)
	}
	if s, _ := none.(string); strings.Contains(s, "┌─") {
		t.Errorf("b-line title ::before content = %v, want no glyph wrap outside b-ascii", none)
	}

	// --- density swap: header height follows --rh -------------------------
	panelSwapClass(t, page, "d-dense", "d-cozy")
	if got := computedStyle(t, hd, "height"); got != "30px" {
		t.Errorf("panel header height at d-cozy = %q, want 30px (--rh)", got)
	}
	panelSwapClass(t, page, "d-cozy", "d-dense")

	// --- token-flow proof: t-mocha root-class swap -------------------------
	panelSwapClass(t, page, "t-gruvbox", "t-mocha")
	if got := computedStyle(t, mode, "backgroundColor"); got != "rgb(137, 180, 250)" {
		t.Errorf("mocha mode cell background = %q, want rgb(137, 180, 250) (--accent)", got)
	}
	if got := computedStyle(t, mode, "color"); got != "rgb(17, 17, 27)" {
		t.Errorf("mocha mode cell color = %q, want rgb(17, 17, 27) (--on-accent)", got)
	}
	if got := computedStyle(t, title, "color"); got != "rgb(166, 173, 200)" {
		t.Errorf("mocha panel title color = %q, want rgb(166, 173, 200) (--fg-muted)", got)
	}
}

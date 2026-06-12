//go:build e2e

package e2e

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// dataExtrasOpenSheet starts the demo + browser, opens the component sheet
// and waits for the page to settle (the Panes behavior applying its grid
// template is the boot signal — htmx loads before hyperscript, so by then
// the pager's hx-get wiring is live too).
func dataExtrasOpenSheet(t *testing.T) (playwright.Page, string) {
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
			return el && getComputedStyle(el).gridTemplateColumns !== 'none' && !!window.htmx;
		}`, nil,
	); err != nil {
		t.Fatalf("page never settled (Panes template/htmx missing): %v", err)
	}
	return page, srv.URL
}

func dataExtrasText(t *testing.T, page playwright.Page, sel string) string {
	t.Helper()
	txt, err := page.Locator(sel).TextContent()
	if err != nil {
		t.Fatalf("text of %s: %v", sel, err)
	}
	return txt
}

// TestPaginationNextSwapsRegion: clicking "next ›" fires the hx-get round
// trip and the server fragment REPLACES the region — the list shows page 2
// and the range text re-renders (both asserted); page-1 rows are gone.
func TestPaginationNextSwapsRegion(t *testing.T) {
	page, _ := dataExtrasOpenSheet(t)

	if got := dataExtrasText(t, page, "#pager-demo .pager-info"); got != "rows 1–5 of 12,403" {
		t.Fatalf("initial range text = %q, want %q", got, "rows 1–5 of 12,403")
	}
	if list := dataExtrasText(t, page, "#pager-list"); !strings.Contains(list, "row 0001") {
		t.Fatalf("initial list = %q, want page-1 rows", list)
	}

	if err := page.Locator(`#pager-demo .pager-btn[aria-label="next page"]`).Click(); err != nil {
		t.Fatalf("click next: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pager-list').textContent.includes('row 0006')`, nil,
	); err != nil {
		t.Fatalf("page-2 rows never swapped in: %v", err)
	}

	list := dataExtrasText(t, page, "#pager-list")
	if strings.Contains(list, "row 0001") {
		t.Errorf("page-1 rows still present after next — append instead of swap: %q", list)
	}
	if got := dataExtrasText(t, page, "#pager-demo .pager-info"); got != "rows 6–10 of 12,403" {
		t.Errorf("range text after next = %q, want %q", got, "rows 6–10 of 12,403")
	}
	if got := dataExtrasText(t, page, "#pager-demo .pager-pos"); got != "2/2481" {
		t.Errorf("position after next = %q, want 2/2481", got)
	}
	// Off the first page, the lower bound comes alive.
	dis, err := page.Locator(`#pager-demo .pager-btn[aria-label="previous page"]`).IsDisabled()
	if err != nil {
		t.Fatalf("prev disabled state: %v", err)
	}
	if dis {
		t.Error("prev still disabled on page 2")
	}
}

// TestPaginationDisabledPrev: on page 1 the prev/first buttons carry a
// real disabled attribute and a (forced) click sends nothing — proven
// deterministically by following with an enabled next-click and counting
// exactly one /demo/pagination request.
func TestPaginationDisabledPrev(t *testing.T) {
	page, _ := dataExtrasOpenSheet(t)

	var reqs int32
	page.OnRequest(func(r playwright.Request) {
		if strings.Contains(r.URL(), "/demo/pagination") {
			atomic.AddInt32(&reqs, 1)
		}
	})

	for _, label := range []string{"previous page", "first page"} {
		sel := `#pager-demo .pager-btn[aria-label="` + label + `"]`
		dis, err := page.Locator(sel).IsDisabled()
		if err != nil {
			t.Fatalf("%s disabled state: %v", label, err)
		}
		if !dis {
			t.Errorf("%s button not disabled on page 1", label)
		}
		// Native .click() on the element — a disabled button must swallow it.
		if _, err := page.Evaluate(`sel => document.querySelector(sel).click()`, sel); err != nil {
			t.Fatalf("dispatch click on %s: %v", label, err)
		}
	}

	// Now an enabled click; once its swap lands, any disabled-click request
	// would already have been counted.
	if err := page.Locator(`#pager-demo .pager-btn[aria-label="next page"]`).Click(); err != nil {
		t.Fatalf("click next: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pager-list').textContent.includes('row 0006')`, nil,
	); err != nil {
		t.Fatalf("next never swapped: %v", err)
	}
	if n := atomic.LoadInt32(&reqs); n != 1 {
		t.Errorf("pagination requests = %d, want exactly 1 (disabled clicks must send none)", n)
	}
}

// TestPaginationLoadMoreAppends: "load more ↓" hx-gets the next page with
// hx-swap=beforeend — page-2 rows are appended and page-1 rows survive.
func TestPaginationLoadMoreAppends(t *testing.T) {
	page, _ := dataExtrasOpenSheet(t)

	before, err := page.Locator("#pager-more-list > div").Count()
	if err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if before != 5 {
		t.Fatalf("initial load-more list has %d rows, want 5", before)
	}

	if err := page.Locator("#pager-more .pager-btn.more").Click(); err != nil {
		t.Fatalf("click load more: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#pager-more-list').textContent.includes('row 0006')`, nil,
	); err != nil {
		t.Fatalf("appended rows never arrived: %v", err)
	}

	list := dataExtrasText(t, page, "#pager-more-list")
	if !strings.Contains(list, "row 0001") {
		t.Errorf("page-1 rows gone after load more — replaced instead of appended: %q", list)
	}
	after, err := page.Locator("#pager-more-list > div").Count()
	if err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if after != 10 {
		t.Errorf("load-more list has %d rows, want 10 (5 + 5 appended)", after)
	}
}

// TestDataExtrasComputedStyles: the pager's accent load-more button, the
// diff row tints (10% ok/err, 8% info via color-mix — partial alpha), the
// --fg-faint gutters, zero radius, and the t-mocha root-class-swap proof.
func TestDataExtrasComputedStyles(t *testing.T) {
	page, _ := dataExtrasOpenSheet(t)

	// --- pager (t-gruvbox) -------------------------------------------------
	if got := computedStyleSel(t, page, "#pager-more .pager-btn.more", "color"); got != gruvAccent {
		t.Errorf("load-more text = %q, want %s (--accent)", got, gruvAccent)
	}
	// 40% accent color-mix border.
	assertToneAlpha(t, computedStyleSel(t, page, "#pager-more .pager-btn.more", "borderTopColor"),
		"250, 189, 47", "load-more border")
	if got := computedStyleSel(t, page, "#pager-demo .pager-info", "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("pager info text = %q, want rgb(124, 111, 100) (--fg-faint)", got)
	}
	if got := computedStyleSel(t, page, "#pager-demo .pager", "borderTopLeftRadius"); got != "0px" {
		t.Errorf("pager border-radius = %q, want 0px", got)
	}
	if got := computedStyleSel(t, page, "#pager-demo .pager", "fontVariantNumeric"); got != "tabular-nums" {
		t.Errorf("pager font-variant-numeric = %q, want tabular-nums", got)
	}

	// --- diff rows (t-gruvbox) ---------------------------------------------
	addBG := computedStyleSel(t, page, "#diff-demo .diff-line.add", "backgroundColor")
	assertToneAlpha(t, addBG, "184, 187, 38", "diff add row background") // --ok at 10%
	if got := computedStyleSel(t, page, "#diff-demo .diff-line.add .diff-sign", "color"); got != "rgb(184, 187, 38)" {
		t.Errorf("add sign colour = %q, want rgb(184, 187, 38) (--ok)", got)
	}
	delBG := computedStyleSel(t, page, "#diff-demo .diff-line.del", "backgroundColor")
	assertToneAlpha(t, delBG, "251, 73, 52", "diff del row background") // --err at 10%
	if got := computedStyleSel(t, page, "#diff-demo .diff-line.del .diff-text", "color"); got != "rgb(251, 73, 52)" {
		t.Errorf("del text colour = %q, want rgb(251, 73, 52) (--err)", got)
	}
	hunkBG := computedStyleSel(t, page, "#diff-demo .diff-line.hunk", "backgroundColor")
	assertToneAlpha(t, hunkBG, "131, 165, 152", "diff hunk row background") // --info at 8%
	if got := computedStyleSel(t, page, "#diff-demo .diff-line.hunk", "color"); got != "rgb(131, 165, 152)" {
		t.Errorf("hunk text colour = %q, want rgb(131, 165, 152) (--info)", got)
	}
	// Gutters keep --fg-faint even inside tinted rows; context rows plain.
	if got := computedStyleSel(t, page, "#diff-demo .diff-line.del .diff-no", "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("del-row gutter = %q, want rgb(124, 111, 100) (--fg-faint)", got)
	}
	assertTransparent(t, computedStyleSel(t, page, "#diff-demo .diff-line:not(.add):not(.del):not(.hunk)", "backgroundColor"),
		"context row background")

	// --- token flow proof: root-class swap to t-mocha -----------------------
	if _, err := page.Evaluate(
		`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	mochaAdd := computedStyleSel(t, page, "#diff-demo .diff-line.add", "backgroundColor")
	assertToneAlpha(t, mochaAdd, "166, 227, 161", "mocha diff add row background") // mocha --ok #a6e3a1
	if mochaAdd == addBG {
		t.Errorf("add row background did not change on theme swap (still %q)", mochaAdd)
	}
	if got := computedStyleSel(t, page, "#diff-demo .diff-line.del .diff-no", "color"); got != "rgb(108, 112, 134)" {
		t.Errorf("mocha gutter = %q, want rgb(108, 112, 134) (mocha --fg-faint)", got)
	}
	if got := computedStyleSel(t, page, "#pager-more .pager-btn.more", "color"); got != mochaAccent {
		t.Errorf("mocha load-more text = %q, want %s (mocha --accent)", got, mochaAccent)
	}
}

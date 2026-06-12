//go:build e2e

package e2e

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// datatableOpenSheet starts the demo + browser, opens the component sheet
// and gates on element-level readiness: the fleet table's body rows are
// attached and hyperscript has booted (Panes grid applied — needed before
// driving the tweaks panel). Never timing-based.
func datatableOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	srv := startDemo(t)
	page := startBrowser(t)
	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}
	if err := page.Locator("#dt-fleet-body tr").First().WaitFor(); err != nil {
		t.Fatalf("fleet table rows never attached: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector('[data-panes]');
			return el && getComputedStyle(el).gridTemplateColumns !== 'none';
		}`, nil,
	); err != nil {
		t.Fatalf("hyperscript never booted (Panes template missing): %v", err)
	}
	return page
}

// datatableFirstHost reads the host cell of the fleet table's first row —
// the sort-order probe (td 1 is the row mark, td 2 the host column).
func datatableFirstHost(t *testing.T, page playwright.Page) string {
	t.Helper()
	txt, err := page.Locator("#dt-fleet-body tr:first-child td:nth-child(2)").TextContent()
	if err != nil {
		t.Fatalf("read first host cell: %v", err)
	}
	return strings.TrimSpace(txt)
}

// datatableWaitFirstHost polls until the first row's host cell equals want
// — the htmx tbody swap landing is the readiness signal after a sort.
func datatableWaitFirstHost(t *testing.T, page playwright.Page, want, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => document.querySelector("#dt-fleet-body tr:first-child td:nth-child(2)")?.textContent.trim() === %q`,
		want), nil); err != nil {
		t.Fatalf("%s: first host never became %q (now %q): %v", what, want, datatableFirstHost(t, page), err)
	}
}

// datatableWaitHeadReady gates on htmx having initialized the (possibly
// OOB-swapped) cpu header: the swap inserts the new thead immediately but
// htmx binds its hx-get listeners only at settle (~20ms later) — DOM
// attributes are visible before the th is interactive, so keyboard/clicks
// must wait for the internal-data marker, not the attributes.
// COUPLING: peeks htmx's PRIVATE htmx-internal-data property, valid for the
// pinned htmx 2.0.4 (baud/baud.go) — revisit on any htmx upgrade.
func datatableWaitHeadReady(t *testing.T, page playwright.Page, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(
		`() => {
			const th = document.querySelector('#dt-fleet thead th[data-key=cpu]');
			return th && !!th['htmx-internal-data'];
		}`, nil,
	); err != nil {
		t.Fatalf("%s: swapped thead never htmx-initialized: %v", what, err)
	}
}

func datatableAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Locator(sel).GetAttribute(name)
	if err != nil {
		t.Fatalf("attribute %s of %s: %v", name, sel, err)
	}
	return v
}

// datatableRowHeight measures a rendered row box via getBoundingClientRect
// — computed "height" on collapsed-border table cells reports the content
// box, so the rect is the honest --row check.
func datatableRowHeight(t *testing.T, page playwright.Page, sel string) float64 {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => document.querySelector(%q).getBoundingClientRect().height`, sel))
	if err != nil {
		t.Fatalf("rect height of %s: %v", sel, err)
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	}
	t.Fatalf("rect height of %s: non-numeric %v", sel, v)
	return 0
}

// TestDataTableHeaderAndCells: the header row is sticky with the
// --bg-raised fill and UPPERCASE --fs-sm type, rows are --row tall
// (22px at the default d-dense), numeric cells right-align and the table
// keeps tabular numerals and zero border-radius.
func TestDataTableHeaderAndCells(t *testing.T) {
	page := datatableOpenSheet(t)

	th := "#dt-fleet thead th[data-key=host]"
	if got := computedStyleSel(t, page, th, "position"); got != "sticky" {
		t.Errorf("header position = %q, want sticky", got)
	}
	if got := computedStyleSel(t, page, th, "backgroundColor"); got != "rgb(50, 48, 47)" {
		t.Errorf("header background = %q, want rgb(50, 48, 47) (--bg-raised)", got)
	}
	if got := computedStyleSel(t, page, th, "textTransform"); got != "uppercase" {
		t.Errorf("header text-transform = %q, want uppercase", got)
	}
	if got := computedStyleSel(t, page, th, "fontSize"); got != "10.5px" {
		t.Errorf("header font-size = %q, want 10.5px (--fs-sm at d-dense)", got)
	}
	if got := computedStyleSel(t, page, th, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("header border-radius = %q, want 0px", got)
	}

	// Row height = --row (22px at d-dense); collapsed hairlines allow ±1.5.
	if got := datatableRowHeight(t, page, "#dt-fleet-body tr:first-child"); math.Abs(got-22) > 1.5 {
		t.Errorf("row height = %v, want ~22 (--row at d-dense)", got)
	}

	if got := computedStyleSel(t, page, "#dt-fleet", "fontVariantNumeric"); got != "tabular-nums" {
		t.Errorf("table font-variant-numeric = %q, want tabular-nums", got)
	}
	// cpu (numeric) right-aligns; host (text) stays left.
	if got := computedStyleSel(t, page, "#dt-fleet-body tr:first-child td:nth-child(4)", "textAlign"); got != "right" {
		t.Errorf("numeric cell text-align = %q, want right", got)
	}
	// td has no explicit text-align — the UA default "start" (left in LTR).
	if got := computedStyleSel(t, page, "#dt-fleet-body tr:first-child td:nth-child(2)", "textAlign"); got != "left" && got != "start" {
		t.Errorf("text cell text-align = %q, want left/start", got)
	}
}

// TestDataTableHoverAndSelection: a resting row is transparent and fills
// --bg-hover under a real mouse hover; the server-selected row carries the
// --sel wash, the 2px accent inset bar and the accent ▌ row mark.
func TestDataTableHoverAndSelection(t *testing.T) {
	page := datatableOpenSheet(t)

	// host-asc puts unselected auth-svc first; search-idx (h3) is selected.
	row := "#dt-fleet-body tr:first-child"
	cell := row + " td:nth-child(2)"
	assertTransparent(t, computedStyleSel(t, page, cell, "backgroundColor"), "resting row background")

	if err := page.Locator(row).Hover(); err != nil {
		t.Fatalf("hover row: %v", err)
	}
	if got := computedStyleSel(t, page, cell, "backgroundColor"); got != "rgb(60, 56, 54)" {
		t.Errorf("hovered row background = %q, want rgb(60, 56, 54) (--bg-hover)", got)
	}
	if err := page.Mouse().Move(0, 0); err != nil {
		t.Fatalf("park mouse: %v", err)
	}

	sel := "#dt-fleet-body tr.sel"
	if got, err := page.Locator(sel + " td:nth-child(2)").TextContent(); err != nil || strings.TrimSpace(got) != "search-idx" {
		t.Fatalf("selected row host = %q (err %v), want search-idx", got, err)
	}
	// --sel is a translucent accent wash: accent channels at partial alpha.
	assertToneAlpha(t, computedStyleSel(t, page, sel+" td:nth-child(2)", "backgroundColor"),
		"250, 189, 47", "selected row background")
	// 2px accent inset bar on the first cell.
	shadow := computedStyleSel(t, page, sel+" td:nth-child(1)", "boxShadow")
	if !strings.Contains(shadow, "inset") || !strings.Contains(shadow, gruvAccent) || !strings.Contains(shadow, "2px") {
		t.Errorf("selected row inset bar = %q, want 2px inset %s", shadow, gruvAccent)
	}
	// ▌ row mark in accent.
	mark, err := page.Locator(sel + " td.row-mark").TextContent()
	if err != nil || strings.TrimSpace(mark) != "▌" {
		t.Errorf("selected row mark = %q (err %v), want ▌", mark, err)
	}
	if got := computedStyleSel(t, page, sel+" td.row-mark", "color"); got != gruvAccent {
		t.Errorf("selected row mark colour = %q, want %s (--accent)", got, gruvAccent)
	}
	// Unselected rows show no mark.
	first, err := page.Locator(row + " td.row-mark").TextContent()
	if err != nil || strings.TrimSpace(first) != "" {
		t.Errorf("unselected row mark = %q (err %v), want empty", first, err)
	}
}

// TestDataTableSortRoundTrip: clicking a sortable th fires the hx-get, the
// swapped tbody is re-ordered (first row flips between the asc and desc
// extremes), and the out-of-band thead moves the accent ▲/▼ indicator and
// flips the active column's URL. Enter on the focused th sorts too.
func TestDataTableSortRoundTrip(t *testing.T) {
	page := datatableOpenSheet(t)

	// Initial state: host asc — auth-svc first, indicator ▲ on host.
	if got := datatableFirstHost(t, page); got != "auth-svc" {
		t.Fatalf("initial first host = %q, want auth-svc (host asc)", got)
	}
	if got := datatableAttr(t, page, "#dt-fleet thead th.sorted", "data-key"); got != "host" {
		t.Fatalf("initial sorted column = %q, want host", got)
	}
	arrow := "#dt-fleet thead th.sorted .sort-arrow"
	if got, _ := page.Locator(arrow).TextContent(); strings.TrimSpace(got) != "▲" {
		t.Errorf("initial indicator = %q, want ▲", got)
	}
	if got := computedStyleSel(t, page, "#dt-fleet thead th.sorted", "color"); got != gruvAccent {
		t.Errorf("sorted header colour = %q, want %s (--accent)", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, arrow, "color"); got != gruvAccent {
		t.Errorf("indicator colour = %q, want %s (--accent)", got, gruvAccent)
	}

	// Click CPU%: asc round-trip — lowest cpu (notif-fan 12.9) first.
	cpu := "#dt-fleet thead th[data-key=cpu]"
	if err := page.Locator(cpu).Click(); err != nil {
		t.Fatalf("click cpu th: %v", err)
	}
	datatableWaitFirstHost(t, page, "notif-fan", "cpu asc")
	// The OOB thead landed: indicator moved to cpu, URL flipped to desc.
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#dt-fleet thead th[data-key=cpu]').classList.contains('sorted')`, nil,
	); err != nil {
		t.Fatalf("OOB thead never moved the sorted marker to cpu: %v", err)
	}
	datatableWaitHeadReady(t, page, "cpu asc")
	if got := datatableAttr(t, page, cpu, "hx-get"); got != "/demo/datatable?sort=cpu&dir=desc" {
		t.Errorf("active column hx-get = %q, want the desc flip", got)
	}
	if got := datatableAttr(t, page, cpu, "aria-sort"); got != "ascending" {
		t.Errorf("aria-sort = %q, want ascending", got)
	}
	if got, _ := page.Locator(arrow).TextContent(); strings.TrimSpace(got) != "▲" {
		t.Errorf("asc indicator = %q, want ▲", got)
	}

	// Click again: desc — highest cpu (ingest-gw 96.7) first, ▼ indicator.
	if err := page.Locator(cpu).Click(); err != nil {
		t.Fatalf("click cpu th (desc): %v", err)
	}
	datatableWaitFirstHost(t, page, "ingest-gw", "cpu desc")
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#dt-fleet thead th[data-key=cpu]')?.getAttribute('aria-sort') === 'descending'`, nil,
	); err != nil {
		t.Fatalf("OOB thead never flipped to descending: %v", err)
	}
	datatableWaitHeadReady(t, page, "cpu desc")
	if got, _ := page.Locator(arrow).TextContent(); strings.TrimSpace(got) != "▼" {
		t.Errorf("desc indicator = %q, want ▼", got)
	}
	if got := datatableAttr(t, page, cpu, "hx-get"); got != "/demo/datatable?sort=cpu&dir=asc" {
		t.Errorf("active column hx-get = %q, want the asc flip back", got)
	}

	// Keyboard: the th is focusable and Enter fires the same round-trip.
	if err := page.Locator(cpu).Focus(); err != nil {
		t.Fatalf("focus cpu th: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	datatableWaitFirstHost(t, page, "notif-fan", "cpu asc via Enter")
}

// TestDataTableSortRejectsBadParams: the handler 400s unknown columns and
// directions instead of guessing.
func TestDataTableSortRejectsBadParams(t *testing.T) {
	srv := startDemo(t)
	for _, q := range []string{"sort=nope&dir=asc", "sort=cpu&dir=sideways", "dir=asc", "sort=cpu"} {
		res, err := srv.Client().Get(srv.URL + "/demo/datatable?" + q)
		if err != nil {
			t.Fatalf("GET ?%s: %v", q, err)
		}
		res.Body.Close()
		if res.StatusCode != 400 {
			t.Errorf("GET ?%s = %d, want 400", q, res.StatusCode)
		}
	}
	res, err := srv.Client().Get(srv.URL + "/demo/datatable?sort=cpu&dir=asc")
	if err != nil {
		t.Fatalf("GET valid sort: %v", err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Errorf("valid sort = %d, want 200", res.StatusCode)
	}
}

// TestDataTableTonesAndVariants: threshold cells resolve the tone tokens
// (cpu 96.7 ⇒ --err), zebra washes even rows with translucent --bg-raised,
// and lines draws hairline column rules.
func TestDataTableTonesAndVariants(t *testing.T) {
	page := datatableOpenSheet(t)

	if got := computedStyle(t, page.Locator("#dt-fleet-body td.tone-err").First(), "color"); got != "rgb(251, 73, 52)" {
		t.Errorf("err threshold cell colour = %q, want rgb(251, 73, 52) (--err)", got)
	}
	if got := computedStyle(t, page.Locator("#dt-fleet-body td.tone-warn").First(), "color"); got != "rgb(250, 189, 47)" {
		t.Errorf("warn threshold cell colour = %q, want rgb(250, 189, 47) (--warn)", got)
	}
	if got := computedStyle(t, page.Locator("#dt-fleet-body td.tone-ok").First(), "color"); got != "rgb(184, 187, 38)" {
		t.Errorf("ok status cell colour = %q, want rgb(184, 187, 38) (--ok)", got)
	}

	// Zebra: even rows get the half-strength --bg-raised wash, odd stay
	// transparent.
	assertToneAlpha(t, computedStyleSel(t, page, "#dt-zebra-body tr:nth-child(2) td:nth-child(2)", "backgroundColor"),
		"50, 48, 47", "zebra even-row background")
	assertTransparent(t, computedStyleSel(t, page, "#dt-zebra-body tr:nth-child(1) td:nth-child(2)", "backgroundColor"),
		"zebra odd-row background")

	// Lines: hairline column rules between cells, none after the last.
	if got := computedStyleSel(t, page, "#dt-lines-body tr:first-child td:nth-child(2)", "borderRightWidth"); got != "1px" {
		t.Errorf("lines column rule width = %q, want 1px", got)
	}
	assertToneAlpha(t, computedStyleSel(t, page, "#dt-lines-body tr:first-child td:nth-child(2)", "borderRightColor"),
		"60, 56, 54", "lines column rule colour")
	if got := computedStyleSel(t, page, "#dt-lines-body tr:first-child td:last-child", "borderRightWidth"); got != "0px" {
		t.Errorf("lines last-column rule width = %q, want 0px", got)
	}
	// The plain fleet table draws no column rules.
	if got := computedStyleSel(t, page, "#dt-fleet-body tr:first-child td:nth-child(2)", "borderRightWidth"); got != "0px" {
		t.Errorf("plain table column rule width = %q, want 0px", got)
	}
}

// TestDataTableMochaSwap: a t-mocha root-class swap re-resolves the sorted
// header, indicator and selection wash from tokens alone — no DOM change.
func TestDataTableMochaSwap(t *testing.T) {
	page := datatableOpenSheet(t)

	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#dt-fleet thead th.sorted')).color === '`+mochaAccent+`'`, nil,
	); err != nil {
		got := computedStyleSel(t, page, "#dt-fleet thead th.sorted", "color")
		t.Fatalf("after t-mocha swap sorted header = %q, want %s: %v", got, mochaAccent, err)
	}
	assertToneAlpha(t, computedStyleSel(t, page, "#dt-fleet-body tr.sel td:nth-child(2)", "backgroundColor"),
		"137, 180, 250", "selected row background under t-mocha")
	if got := computedStyleSel(t, page, "#dt-fleet thead th[data-key=host]", "backgroundColor"); got != "rgb(30, 30, 46)" {
		t.Errorf("header background under t-mocha = %q, want rgb(30, 30, 46) (--bg-raised)", got)
	}
}

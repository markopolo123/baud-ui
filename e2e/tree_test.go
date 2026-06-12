//go:build e2e

package e2e

import (
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Selectors into the sheet's #tree-fleet fixture (demo/sheet_tree.templ —
// structure documented load-bearing there): root li order is prod,
// staging, docs (state-independent — [open] flips during the tests),
// edge is the only lazy ([hx-get]) branch, ingest-gw the only selected
// row.
const (
	treeFleet      = "#tree-fleet"
	treeStaging    = treeFleet + " > li:nth-child(2) > details.tree-branch"
	treeEdge       = treeFleet + " details.tree-branch[hx-get]"
	treeSel        = treeFleet + " .tree-row.sel"
	treeMochaFaint = "rgb(108, 112, 134)"        // t-mocha --fg-faint #6c7086
	treeGruvFaint  = "rgb(124, 111, 100)"        // t-gruvbox --fg-faint #7c6f64
	treeGruvSelBG  = "rgba(250, 189, 47, 0.14)"  // t-gruvbox --sel
	treeMochaSelBG = "rgba(137, 180, 250, 0.13)" // t-mocha --sel
)

// treeOpenSheet starts the demo + browser, opens the component sheet and
// waits for hyperscript boot (the Panes behavior applying its grid
// template) so the inline toggle handlers are live before interacting.
func treeOpenSheet(t *testing.T) (playwright.Page, string) {
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
		t.Fatalf("hyperscript never booted (Panes template missing): %v", err)
	}
	return page, srv.URL
}

// treeAttr reads an attribute via page.Evaluate — locator reads would
// wait for visibility, which rows inside closed branches never reach.
func treeAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Evaluate(
		`([sel, name]) => { const el = document.querySelector(sel); if (!el) return "MISSING-EL"; const a = el.getAttribute(name); return a === null ? "MISSING-ATTR" : a; }`,
		[]string{sel, name})
	if err != nil {
		t.Fatalf("attribute %s of %s: %v", name, sel, err)
	}
	s, _ := v.(string)
	if s == "MISSING-EL" {
		t.Fatalf("no element matches %q", sel)
	}
	return s
}

// treeDisclosure resolves the ::before content of a branch summary's
// disclosure slot — the ▸/▾ flip is pure CSS on the details open state.
func treeDisclosure(t *testing.T, page playwright.Page, detailsSel string) string {
	t.Helper()
	v, err := page.Evaluate(
		`sel => { const el = document.querySelector(sel + ' > summary .tree-disclosure'); return el ? getComputedStyle(el, '::before').content : "MISSING"; }`,
		detailsSel)
	if err != nil {
		t.Fatalf("disclosure content of %s: %v", detailsSel, err)
	}
	s, _ := v.(string)
	if s == "MISSING" {
		t.Fatalf("no disclosure under %q", detailsSel)
	}
	return strings.Trim(s, `"`)
}

// treeGroupRows counts the li rows currently inside a branch's group.
func treeGroupRows(t *testing.T, page playwright.Page, detailsSel string) int {
	t.Helper()
	v, err := page.Evaluate(
		`sel => document.querySelectorAll(sel + ' > ul > li').length`, detailsSel)
	if err != nil {
		t.Fatalf("group rows of %s: %v", detailsSel, err)
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	t.Fatalf("group rows of %s: non-number %T(%v)", detailsSel, v, v)
	return 0
}

// treePx parses a computed "12px" style value.
func treePx(t *testing.T, v string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(strings.TrimSuffix(v, "px"), 64)
	if err != nil {
		t.Fatalf("non-px computed value %q: %v", v, err)
	}
	return f
}

// treeCountRequests counts same-origin requests whose URL contains frag
// from this point on.
func treeCountRequests(page playwright.Page, base, frag string) *int32 {
	var n int32
	page.OnRequest(func(r playwright.Request) {
		if strings.HasPrefix(r.URL(), base) && strings.Contains(r.URL(), frag) {
			atomic.AddInt32(&n, 1)
		}
	})
	return &n
}

// TestTreeGlyphAndSelectedStyles: branch glyphs resolve --fg-faint mono
// with pre whitespace; the selected row paints --sel + the 2px accent
// inset bar + accent glyph + aria-current; zero radius. A t-mocha
// root-class swap re-resolves everything from tokens alone.
func TestTreeGlyphAndSelectedStyles(t *testing.T) {
	page, _ := treeOpenSheet(t)

	glyph := treeEdge + " > summary .tree-glyph"
	if got := computedStyleSel(t, page, glyph, "color"); got != treeGruvFaint {
		t.Errorf("branch glyph color = %q, want %s (--fg-faint)", got, treeGruvFaint)
	}
	if got := computedStyleSel(t, page, glyph, "whiteSpace"); got != "pre" {
		t.Errorf("glyph white-space = %q, want pre (box-drawing alignment)", got)
	}

	if got := computedStyleSel(t, page, treeSel, "backgroundColor"); got != treeGruvSelBG {
		t.Errorf("selected row background = %q, want %s (--sel)", got, treeGruvSelBG)
	}
	shadow := computedStyleSel(t, page, treeSel, "boxShadow")
	if !strings.Contains(shadow, "inset") || !strings.Contains(shadow, gruvAccent) || !strings.Contains(shadow, "2px") {
		t.Errorf("selected row box-shadow = %q, want 2px inset %s accent bar", shadow, gruvAccent)
	}
	if got := computedStyleSel(t, page, treeSel+" .tree-glyph", "color"); got != gruvAccent {
		t.Errorf("selected glyph color = %q, want %s (--accent)", got, gruvAccent)
	}
	if got := treeAttr(t, page, treeSel, "aria-current"); got != "true" {
		t.Errorf("selected row aria-current = %q, want true", got)
	}
	if got := computedStyleSel(t, page, treeSel, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("tree row border-radius = %q, want 0px", got)
	}
	// Meta renders in the faint small type (--fs-sm < the row's --fs).
	if got := computedStyleSel(t, page, treeSel+" .tree-meta", "color"); got != treeGruvFaint {
		t.Errorf("meta color = %q, want %s (--fg-faint)", got, treeGruvFaint)
	}
	metaFS := treePx(t, computedStyleSel(t, page, treeSel+" .tree-meta", "fontSize"))
	rowFS := treePx(t, computedStyleSel(t, page, treeSel+" .tree-label", "fontSize"))
	if metaFS >= rowFS {
		t.Errorf("meta font-size %vpx not smaller than row %vpx (--fs-sm vs --fs)", metaFS, rowFS)
	}

	// Token flow: swap to t-mocha by root class — glyph, selected fill and
	// inset bar must re-resolve with no DOM change.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => getComputedStyle(document.querySelector(sel)).color === '`+treeMochaFaint+`'`,
		glyph,
	); err != nil {
		got := computedStyleSel(t, page, glyph, "color")
		t.Errorf("after t-mocha swap glyph = %q, want %s: %v", got, treeMochaFaint, err)
	}
	if got := computedStyleSel(t, page, treeSel, "backgroundColor"); got != treeMochaSelBG {
		t.Errorf("after t-mocha swap selected background = %q, want %s", got, treeMochaSelBG)
	}
	if got := computedStyleSel(t, page, treeSel+" .tree-glyph", "color"); got != mochaAccent {
		t.Errorf("after t-mocha swap selected glyph = %q, want %s", got, mochaAccent)
	}
}

// TestTreeExpandCollapse: clicking a loaded branch's summary toggles the
// native details, the hyperscript handler mirrors aria-expanded, the
// CSS disclosure flips ▸→▾, children stay in the DOM when collapsed —
// and zero server requests are made.
func TestTreeExpandCollapse(t *testing.T) {
	page, base := treeOpenSheet(t)
	reqs := treeCountRequests(page, base, "/demo/")

	summary := treeStaging + " > summary"
	if got := treeAttr(t, page, summary, "aria-expanded"); got != "false" {
		t.Fatalf("collapsed branch aria-expanded = %q, want false", got)
	}
	if got := treeDisclosure(t, page, treeStaging); got != "▸ " {
		t.Fatalf("collapsed disclosure = %q, want \"▸ \"", got)
	}
	if n := treeGroupRows(t, page, treeStaging); n != 2 {
		t.Fatalf("staging group rows = %d, want 2 pre-rendered", n)
	}

	if err := page.Locator(summary).Click(); err != nil {
		t.Fatalf("click staging summary: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => { const d = document.querySelector(sel); return d.open && d.querySelector('summary').getAttribute('aria-expanded') === 'true'; }`,
		treeStaging,
	); err != nil {
		t.Fatalf("expand never opened + synced aria-expanded: %v", err)
	}
	if got := treeDisclosure(t, page, treeStaging); got != "▾ " {
		t.Errorf("expanded disclosure = %q, want \"▾ \"", got)
	}
	smoke := treeStaging + " > ul > li.tree-row"
	visible, err := page.Locator(smoke).First().IsVisible()
	if err != nil || !visible {
		t.Errorf("expanded child row not visible (err=%v)", err)
	}

	if err := page.Locator(summary).Click(); err != nil {
		t.Fatalf("click staging summary again: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => { const d = document.querySelector(sel); return !d.open && d.querySelector('summary').getAttribute('aria-expanded') === 'false'; }`,
		treeStaging,
	); err != nil {
		t.Fatalf("collapse never closed + synced aria-expanded: %v", err)
	}
	if n := treeGroupRows(t, page, treeStaging); n != 2 {
		t.Errorf("collapsed group rows = %d, want 2 kept in DOM", n)
	}
	if n := atomic.LoadInt32(reqs); n != 0 {
		t.Errorf("local expand/collapse made %d server request(s), want 0", n)
	}
}

// TestTreeKeyboard: the summary is natively focusable; Enter and Space
// toggle the branch, with aria-expanded following each time.
func TestTreeKeyboard(t *testing.T) {
	page, _ := treeOpenSheet(t)

	summary := treeStaging + " > summary"
	if err := page.Locator(summary).Focus(); err != nil {
		t.Fatalf("focus staging summary: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => { const d = document.querySelector(sel); return d.open && d.querySelector('summary').getAttribute('aria-expanded') === 'true'; }`,
		treeStaging,
	); err != nil {
		t.Fatalf("Enter never expanded the branch: %v", err)
	}
	if err := page.Keyboard().Press("Space"); err != nil {
		t.Fatalf("press Space: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => { const d = document.querySelector(sel); return !d.open && d.querySelector('summary').getAttribute('aria-expanded') === 'false'; }`,
		treeStaging,
	); err != nil {
		t.Fatalf("Space never collapsed the branch: %v", err)
	}
}

// TestTreeLazyChildren: a lazy branch starts collapsed and empty; the
// first expand fires exactly one hx-get whose TreeChildren fragment
// fills the group with prefix-continued glyph rows; collapsing and
// re-expanding does NOT refetch (toggle-once) nor duplicate rows.
func TestTreeLazyChildren(t *testing.T) {
	page, base := treeOpenSheet(t)
	reqs := treeCountRequests(page, base, "/demo/tree")

	if n := treeGroupRows(t, page, treeEdge); n != 0 {
		t.Fatalf("lazy group rows = %d, want 0 before first expand", n)
	}
	summary := treeEdge + " > summary"
	if err := page.Locator(summary).Click(); err != nil {
		t.Fatalf("click edge summary: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => document.querySelectorAll(sel + ' > ul > li.tree-row').length === 3`,
		treeEdge,
	); err != nil {
		t.Fatalf("lazy children never swapped in (rows = %d): %v", treeGroupRows(t, page, treeEdge), err)
	}
	if got := treeAttr(t, page, summary, "aria-expanded"); got != "true" {
		t.Errorf("expanded lazy branch aria-expanded = %q, want true", got)
	}
	// Server-rendered glyph continuity: edge is prod's last child, so its
	// children continue under a three-space prefix.
	firstGlyph, err := page.Locator(treeEdge + " > ul > li.tree-row .tree-glyph").First().TextContent()
	if err != nil {
		t.Fatalf("read fetched glyph: %v", err)
	}
	if !strings.HasPrefix(firstGlyph, "   ├─") {
		t.Errorf("fetched row glyph = %q, want prefix \"   ├─\"", firstGlyph)
	}
	if txt, _ := page.Locator(treeEdge + " > ul").TextContent(); !strings.Contains(txt, "edge-cache-1") {
		t.Errorf("fetched group = %q, want edge-cache-1 row", txt)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Fatalf("first expand made %d /demo/tree request(s), want 1", n)
	}

	// Collapse, expand again: toggle-once must not refetch or duplicate.
	if err := page.Locator(summary).Click(); err != nil {
		t.Fatalf("collapse edge: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => !document.querySelector(sel).open`, treeEdge,
	); err != nil {
		t.Fatalf("edge never collapsed: %v", err)
	}
	if err := page.Locator(summary).Click(); err != nil {
		t.Fatalf("re-expand edge: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => { const d = document.querySelector(sel); return d.open && d.querySelector('summary').getAttribute('aria-expanded') === 'true'; }`,
		treeEdge,
	); err != nil {
		t.Fatalf("edge never re-expanded: %v", err)
	}
	if n := treeGroupRows(t, page, treeEdge); n != 3 {
		t.Errorf("re-expanded group rows = %d, want 3 (no duplication)", n)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("after re-expand %d /demo/tree request(s), want 1 (toggle once)", n)
	}
}

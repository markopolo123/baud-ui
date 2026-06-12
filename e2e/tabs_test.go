//go:build e2e

package e2e

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// tabsOpenSheet starts the demo + browser, opens the component sheet and
// waits for hyperscript to finish processing the page (the Panes behavior
// applying its grid template is the boot signal) so the installed Tabs
// behavior is live before any interaction.
func tabsOpenSheet(t *testing.T) (playwright.Page, string) {
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

func tabsAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Locator(sel).GetAttribute(name)
	if err != nil {
		t.Fatalf("attribute %s of %s: %v", name, sel, err)
	}
	return v
}

// tabsHidden reads the hidden attribute via page.Evaluate — locator-based
// evaluation would wait for visibility, which a hidden pane never reaches.
func tabsHidden(t *testing.T, page playwright.Page, sel string) bool {
	t.Helper()
	v, err := page.Evaluate(
		`sel => { const el = document.querySelector(sel); if (!el) return "MISSING"; return el.hasAttribute('hidden'); }`, sel)
	if err != nil {
		t.Fatalf("hidden of %s: %v", sel, err)
	}
	if v == "MISSING" {
		t.Fatalf("no element matches %q", sel)
	}
	b, _ := v.(bool)
	return b
}

func tabsActiveID(t *testing.T, page playwright.Page) string {
	t.Helper()
	v, err := page.Evaluate(`() => document.activeElement && document.activeElement.id`)
	if err != nil {
		t.Fatalf("activeElement: %v", err)
	}
	s, _ := v.(string)
	return s
}

// TestTabsUnderlineStyles: the active underline tab paints the 2px accent
// underline and full-strength text; inactive tabs stay transparent/faint.
// The count badge renders through the Badge primitive at tab scale. A
// t-mocha root-class swap re-resolves the underline from tokens alone.
func TestTabsUnderlineStyles(t *testing.T) {
	page, _ := tabsOpenSheet(t)

	active := "#tabs-under .tab.is-active"
	inactive := "#tabs-under-tab-1"

	if got := computedStyleSel(t, page, active, "borderBottomWidth"); got != "2px" {
		t.Errorf("active underline width = %q, want 2px", got)
	}
	if got := computedStyleSel(t, page, active, "borderBottomColor"); got != gruvAccent {
		t.Errorf("active underline color = %q, want %s (--accent)", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, active, "color"); got != "rgb(235, 219, 178)" {
		t.Errorf("active tab text = %q, want rgb(235, 219, 178) (--fg)", got)
	}
	assertTransparent(t, computedStyleSel(t, page, inactive, "borderBottomColor"), "inactive underline")
	if got := computedStyleSel(t, page, inactive, "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("inactive tab text = %q, want rgb(124, 111, 100) (--fg-faint)", got)
	}
	if got := computedStyleSel(t, page, active, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("tab border-radius = %q, want 0px", got)
	}

	// Count badge inside the active underline tab: --bg-active fill + --fg
	// text — fails if the .tab .badge compaction rules are removed (the
	// neutral tint Badge default is --bg-raised + --fg-muted).
	badge := active + " .badge"
	if got := computedStyleSel(t, page, badge, "backgroundColor"); got != "rgb(80, 73, 69)" {
		t.Errorf("active tab badge background = %q, want rgb(80, 73, 69) (--bg-active)", got)
	}
	if got := computedStyleSel(t, page, badge, "color"); got != "rgb(235, 219, 178)" {
		t.Errorf("active tab badge text = %q, want rgb(235, 219, 178) (--fg)", got)
	}

	// Token flow: swap the theme root class to t-mocha via the tweaks panel —
	// the active underline must re-resolve to mocha --accent, no DOM change.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#tabs-under .tab.is-active')).borderBottomColor === '`+mochaAccent+`'`, nil,
	); err != nil {
		got := computedStyleSel(t, page, active, "borderBottomColor")
		t.Errorf("after t-mocha swap underline = %q, want %s: %v", got, mochaAccent, err)
	}
}

// TestTabsBoxedStyles: the boxed strip is bordered, its active tab fills
// accent with on-accent text, inactive segments stay transparent/faint.
func TestTabsBoxedStyles(t *testing.T) {
	page, _ := tabsOpenSheet(t)

	if got := computedStyleSel(t, page, "#tabs-boxed", "borderTopColor"); got != "rgb(80, 73, 69)" {
		t.Errorf("boxed strip border = %q, want rgb(80, 73, 69) (--border-strong)", got)
	}
	active := "#tabs-boxed .tab.is-active"
	if got := computedStyleSel(t, page, active, "backgroundColor"); got != gruvAccent {
		t.Errorf("boxed active background = %q, want %s (--accent)", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, active, "color"); got != gruvOnAccent {
		t.Errorf("boxed active text = %q, want %s (--on-accent)", got, gruvOnAccent)
	}
	inactive := "#tabs-boxed-tab-1"
	assertTransparent(t, computedStyleSel(t, page, inactive, "backgroundColor"), "boxed inactive background")
	if got := computedStyleSel(t, page, inactive, "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("boxed inactive text = %q, want rgb(124, 111, 100) (--fg-faint)", got)
	}
}

// TestTabsLocalSwitching: clicking a local-mode tab swaps the pre-rendered
// panes and the ARIA selection purely client-side — zero requests reach the
// demo server.
func TestTabsLocalSwitching(t *testing.T) {
	page, base := tabsOpenSheet(t)

	if tabsHidden(t, page, "#pane-pods") {
		t.Fatal("pods pane should start visible")
	}
	if !tabsHidden(t, page, "#pane-events") {
		t.Fatal("events pane should start hidden")
	}

	// Count same-origin requests from here on — a local swap must make none.
	var reqs int32
	page.OnRequest(func(r playwright.Request) {
		if strings.HasPrefix(r.URL(), base) {
			atomic.AddInt32(&reqs, 1)
		}
	})

	if err := page.Locator("#tabs-under-tab-1").Click(); err != nil {
		t.Fatalf("click events tab: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.getElementById('pane-events').hasAttribute('hidden')`, nil,
	); err != nil {
		t.Fatalf("events pane never became visible: %v", err)
	}
	if !tabsHidden(t, page, "#pane-pods") {
		t.Error("pods pane stayed visible after switching away")
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-1", "aria-selected"); got != "true" {
		t.Errorf("clicked tab aria-selected = %q, want true", got)
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-0", "aria-selected"); got != "false" {
		t.Errorf("previous tab aria-selected = %q, want false", got)
	}
	// is-active (and the accent underline with it) follows the selection.
	if got := computedStyleSel(t, page, "#tabs-under-tab-1", "borderBottomColor"); got != gruvAccent {
		t.Errorf("clicked tab underline = %q, want %s", got, gruvAccent)
	}
	if n := atomic.LoadInt32(&reqs); n != 0 {
		t.Errorf("local tab switch made %d server request(s), want 0", n)
	}
}

// TestTabsHTMXSwitching: clicking an htmx-mode tab issues the hx-get round
// trip and the server-rendered pane replaces the shared tabpanel content.
func TestTabsHTMXSwitching(t *testing.T) {
	page, _ := tabsOpenSheet(t)

	txt, err := page.Locator("#tabs-range-pane").TextContent()
	if err != nil {
		t.Fatalf("read range pane: %v", err)
	}
	if !strings.Contains(txt, "5m") || !strings.Contains(txt, "60 samples") {
		t.Fatalf("initial range pane = %q, want the 5m/60-samples pane", txt)
	}

	if err := page.Locator("#tabs-boxed-tab-1").Click(); err != nil {
		t.Fatalf("click 1h tab: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.getElementById('tabs-range-pane').textContent.includes('720 samples')`, nil,
	); err != nil {
		got, _ := page.Locator("#tabs-range-pane").TextContent()
		t.Fatalf("server pane never swapped in (pane = %q): %v", got, err)
	}
	if got, _ := page.Locator("#tabs-range-pane").TextContent(); !strings.Contains(got, "1h") {
		t.Errorf("swapped pane = %q, want the 1h window", got)
	}
	// The strip's selection state follows the htmx activation too.
	if got := tabsAttr(t, page, "#tabs-boxed-tab-1", "aria-selected"); got != "true" {
		t.Errorf("htmx tab aria-selected = %q, want true", got)
	}
	if got := computedStyleSel(t, page, "#tabs-boxed-tab-1", "backgroundColor"); got != gruvAccent {
		t.Errorf("htmx active tab background = %q, want %s (accent fill)", got, gruvAccent)
	}
	if got := tabsAttr(t, page, "#tabs-boxed-tab-0", "aria-selected"); got != "false" {
		t.Errorf("previous htmx tab aria-selected = %q, want false", got)
	}
}

// TestTabsKeyboard: ←/→ rove focus across the tablist (wrapping, roving
// tabindex) without activating; ↵ and Space activate the focused tab.
func TestTabsKeyboard(t *testing.T) {
	page, _ := tabsOpenSheet(t)

	if err := page.Locator("#tabs-under-tab-0").Focus(); err != nil {
		t.Fatalf("focus first tab: %v", err)
	}
	if err := page.Keyboard().Press("ArrowRight"); err != nil {
		t.Fatalf("press ArrowRight: %v", err)
	}
	if got := tabsActiveID(t, page); got != "tabs-under-tab-1" {
		t.Fatalf("ArrowRight focused %q, want tabs-under-tab-1", got)
	}
	// Focus roves, selection does not (manual activation pattern).
	if got := tabsAttr(t, page, "#tabs-under-tab-0", "aria-selected"); got != "true" {
		t.Errorf("arrow key changed selection: tab-0 aria-selected = %q, want true", got)
	}
	// Roving tabindex follows focus.
	if got := tabsAttr(t, page, "#tabs-under-tab-1", "tabindex"); got != "0" {
		t.Errorf("focused tab tabindex = %q, want 0", got)
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-0", "tabindex"); got != "-1" {
		t.Errorf("unfocused tab tabindex = %q, want -1", got)
	}

	// ← wraps from the first tab to the last.
	if err := page.Keyboard().Press("ArrowLeft"); err != nil {
		t.Fatalf("press ArrowLeft: %v", err)
	}
	if got := tabsActiveID(t, page); got != "tabs-under-tab-0" {
		t.Fatalf("ArrowLeft focused %q, want tabs-under-tab-0", got)
	}
	if err := page.Keyboard().Press("ArrowLeft"); err != nil {
		t.Fatalf("press ArrowLeft: %v", err)
	}
	if got := tabsActiveID(t, page); got != "tabs-under-tab-3" {
		t.Fatalf("ArrowLeft did not wrap: focused %q, want tabs-under-tab-3", got)
	}

	// ↵ activates the focused tab: selection + pane swap.
	if err := page.Keyboard().Press("ArrowLeft"); err != nil { // → tab-2 (logs)
		t.Fatalf("press ArrowLeft: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.getElementById('pane-logs').hasAttribute('hidden')`, nil,
	); err != nil {
		t.Fatalf("Enter never activated the logs tab: %v", err)
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-2", "aria-selected"); got != "true" {
		t.Errorf("Enter-activated tab aria-selected = %q, want true", got)
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-0", "aria-selected"); got != "false" {
		t.Errorf("old tab aria-selected = %q, want false", got)
	}

	// Space activates too.
	if err := page.Keyboard().Press("ArrowRight"); err != nil { // → tab-3 (env)
		t.Fatalf("press ArrowRight: %v", err)
	}
	if err := page.Keyboard().Press("Space"); err != nil {
		t.Fatalf("press Space: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.getElementById('pane-env').hasAttribute('hidden')`, nil,
	); err != nil {
		t.Fatalf("Space never activated the env tab: %v", err)
	}
	if got := tabsAttr(t, page, "#tabs-under-tab-3", "aria-selected"); got != "true" {
		t.Errorf("Space-activated tab aria-selected = %q, want true", got)
	}
	if !tabsHidden(t, page, "#pane-logs") {
		t.Error("logs pane stayed visible after activating env")
	}
}

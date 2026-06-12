//go:build e2e

package e2e

// Overlay (Modal/Drawer) browser coverage, driving the sheet section in
// demo/sheet_overlay.templ (structure documented load-bearing there):
// #ov-open-modal/#ov-open-drawer toggle the static overlays #ov-modal /
// #ov-drawer; #ov-load-modal/#ov-load-drawer hx-get /demo/modal and
// /demo/drawer fragments into body/beforeend (#demo-modal, #demo-drawer).

import (
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved tokens asserted below (assets/css/tokens.css):
const (
	ovGruvStrong  = "rgb(80, 73, 69)" // t-gruvbox --border-strong #504945
	ovMochaStrong = "rgb(69, 71, 90)" // t-mocha --border-strong #45475a
	ovGruvPanel   = "rgb(40, 40, 40)" // t-gruvbox --bg-panel #282828
	ovMochaPanel  = "rgb(24, 24, 37)" // t-mocha --bg-panel #181825
	ovGruvRaised  = "rgb(50, 48, 47)" // t-gruvbox --bg-raised #32302f
)

// ovOpenSheet opens the component sheet and gates on the honest
// behavior-ready signals: ParseHealth marking the body (whole baud._hs
// parsed) and the static overlays' Overlay listeners attached
// (.overlay-ready, cf the .resizable-ready precedent) — key presses or
// clicks before those would be lost.
func ovOpenSheet(t *testing.T) (playwright.Page, string) {
	t.Helper()
	srv := startDemo(t)
	page := startBrowser(t)
	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}
	if err := page.Locator("body[data-hs-ok]").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("body[data-hs-ok] never appeared — baud._hs parse error?: %v", err)
	}
	if err := page.Locator("#ov-modal.overlay-ready").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("Overlay never signalled .overlay-ready: %v", err)
	}
	return page, srv.URL
}

// ovWaitReady waits for an htmx-injected overlay to land AND its Overlay
// listeners to attach.
func ovWaitReady(t *testing.T, page playwright.Page, overlaySel string) {
	t.Helper()
	if err := page.Locator(overlaySel + ".overlay-ready").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("injected overlay %s never became ready: %v", overlaySel, err)
	}
}

// ovWaitFocus waits until document.activeElement matches sel — the only
// honest assertion for trap/restore behaviour.
func ovWaitFocus(t *testing.T, page playwright.Page, sel, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(
		`sel => document.activeElement && document.activeElement.matches(sel)`, sel,
	); err != nil {
		got, _ := page.Evaluate(`() => document.activeElement && (document.activeElement.id || document.activeElement.outerHTML.slice(0, 80))`)
		t.Fatalf("%s: activeElement never matched %q (now on %v): %v", what, sel, got, err)
	}
}

// ovWaitDetached waits for an overlay to leave the DOM (htmx variant close).
func ovWaitDetached(t *testing.T, page playwright.Page, sel string) {
	t.Helper()
	if err := page.Locator(sel).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("overlay %s never left the DOM: %v", sel, err)
	}
}

// ovAttr reads an attribute via Evaluate (works on hidden static overlays).
func ovAttr(t *testing.T, page playwright.Page, sel, name string) string {
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

// ovCountRequests counts same-origin requests whose URL contains frag from
// this point on.
func ovCountRequests(page playwright.Page, base, frag string) *int32 {
	var n int32
	page.OnRequest(func(r playwright.Request) {
		if strings.HasPrefix(r.URL(), base) && strings.Contains(r.URL(), frag) {
			atomic.AddInt32(&n, 1)
		}
	})
	return &n
}

// ovStillAttached settles two animation frames then reports whether sel is
// still in the DOM — a sleepless "nothing closed" probe.
func ovStillAttached(t *testing.T, page playwright.Page, sel string) bool {
	t.Helper()
	v, err := page.Evaluate(
		`sel => new Promise(res => requestAnimationFrame(() => requestAnimationFrame(() => res(document.querySelector(sel) !== null))))`, sel)
	if err != nil {
		t.Fatalf("attached probe %s: %v", sel, err)
	}
	b, _ := v.(bool)
	return b
}

// ovNum coerces a playwright Evaluate number (int for whole values,
// float64 otherwise).
func ovNum(t *testing.T, v interface{}, what string) float64 {
	t.Helper()
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	}
	t.Fatalf("%s: non-number %T(%v)", what, v, v)
	return 0
}

// ovPx parses a computed "12px" value.
func ovPx(t *testing.T, v string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(strings.TrimSuffix(v, "px"), 64)
	if err != nil {
		t.Fatalf("non-px computed value %q: %v", v, err)
	}
	return f
}

// TestOverlayHtmxModalRoundTripTrapAndRestore: the opener hx-gets the
// modal fragment into body/beforeend (exactly one request), focus lands
// inside, Tab is trapped at both edges (proven on document.activeElement),
// and Esc removes the overlay from the DOM and restores opener focus.
func TestOverlayHtmxModalRoundTripTrapAndRestore(t *testing.T) {
	page, base := ovOpenSheet(t)
	reqs := ovCountRequests(page, base, "/demo/modal")

	if err := page.Locator("#ov-load-modal").Click(); err != nil {
		t.Fatalf("click htmx modal opener: %v", err)
	}
	ovWaitReady(t, page, "#demo-modal")
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("opening made %d /demo/modal request(s), want 1", n)
	}

	// ARIA contract on the injected dialog.
	if got := ovAttr(t, page, "#demo-modal .modal", "role"); got != "dialog" {
		t.Errorf("modal role = %q, want dialog", got)
	}
	if got := ovAttr(t, page, "#demo-modal .modal", "aria-modal"); got != "true" {
		t.Errorf("aria-modal = %q, want true", got)
	}
	if got := ovAttr(t, page, "#demo-modal .modal", "aria-labelledby"); got != "demo-modal-title" {
		t.Errorf("aria-labelledby = %q, want demo-modal-title", got)
	}
	title, err := page.Locator("#demo-modal-title").TextContent()
	if err != nil || strings.TrimSpace(title) != "kill pod" {
		t.Errorf("labelling title = %q (err=%v), want \"kill pod\"", title, err)
	}

	// Focus lands inside: the close glyph is the first focusable.
	ovWaitFocus(t, page, "#demo-modal .x-btn", "open focus")

	// Trap: Tab from the LAST focusable (the danger action) wraps to the
	// first; Shift+Tab from the first wraps back to the last.
	if err := page.Locator("#demo-modal .modal-ft .btn-danger").Focus(); err != nil {
		t.Fatalf("focus last focusable: %v", err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	ovWaitFocus(t, page, "#demo-modal .x-btn", "Tab wrap last→first")
	if err := page.Keyboard().Press("Shift+Tab"); err != nil {
		t.Fatalf("press Shift+Tab: %v", err)
	}
	ovWaitFocus(t, page, "#demo-modal .modal-ft .btn-danger", "Shift+Tab wrap first→last")

	// Esc: htmx-injected overlay leaves the DOM, opener regains focus.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	ovWaitDetached(t, page, "#demo-modal")
	ovWaitFocus(t, page, "#ov-load-modal", "Esc focus restore")
}

// TestOverlayBackdropAndCloseControls: a click inside the dialog does NOT
// close; the backdrop click and the [data-overlay-close] cancel button do,
// removing the htmx overlay and restoring opener focus each time.
func TestOverlayBackdropAndCloseControls(t *testing.T) {
	page, _ := ovOpenSheet(t)

	// -- backdrop click on the drawer variant -----------------------------
	if err := page.Locator("#ov-load-drawer").Click(); err != nil {
		t.Fatalf("click htmx drawer opener: %v", err)
	}
	ovWaitReady(t, page, "#demo-drawer")
	if err := page.Locator("#demo-drawer .drawer-bd").Click(); err != nil {
		t.Fatalf("click drawer body: %v", err)
	}
	if !ovStillAttached(t, page, "#demo-drawer") {
		t.Fatalf("clicking inside the drawer closed it")
	}
	// The drawer hugs the right edge: (5,5) hits the bare backdrop.
	if err := page.Locator("#demo-drawer").Click(playwright.LocatorClickOptions{
		Position: &playwright.Position{X: 5, Y: 5},
	}); err != nil {
		t.Fatalf("click backdrop: %v", err)
	}
	ovWaitDetached(t, page, "#demo-drawer")
	ovWaitFocus(t, page, "#ov-load-drawer", "backdrop-close focus restore")

	// -- ✕ and cancel controls on the modal variant ------------------------
	if err := page.Locator("#ov-load-modal").Click(); err != nil {
		t.Fatalf("click htmx modal opener: %v", err)
	}
	ovWaitReady(t, page, "#demo-modal")
	if err := page.Locator("#demo-modal .modal-hd .x-btn").Click(); err != nil {
		t.Fatalf("click ✕: %v", err)
	}
	ovWaitDetached(t, page, "#demo-modal")
	ovWaitFocus(t, page, "#ov-load-modal", "✕-close focus restore")

	if err := page.Locator("#ov-load-modal").Click(); err != nil {
		t.Fatalf("re-open htmx modal: %v", err)
	}
	ovWaitReady(t, page, "#demo-modal")
	if err := page.Locator("#demo-modal .modal-ft [data-overlay-close]").Click(); err != nil {
		t.Fatalf("click cancel: %v", err)
	}
	ovWaitDetached(t, page, "#demo-modal")
	ovWaitFocus(t, page, "#ov-load-modal", "cancel-close focus restore")
}

// TestOverlayStaticToggleOneAtATime: the static overlay is display:none
// until opened by class toggle, only one overlay is open at a time, Esc
// drops .open while keeping the element in the DOM, and a clean
// open/close cycle restores opener focus. The drawer (whose ✕ is its only
// focusable) also proves the single-element Tab cycle.
func TestOverlayStaticToggleOneAtATime(t *testing.T) {
	page, _ := ovOpenSheet(t)

	if got := computedStyleSel(t, page, "#ov-modal", "display"); got != "none" {
		t.Fatalf("closed static overlay display = %q, want none", got)
	}
	if err := page.Locator("#ov-open-modal").Click(); err != nil {
		t.Fatalf("click static modal opener: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#ov-modal').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("static modal never gained .open: %v", err)
	}
	if got := computedStyleSel(t, page, "#ov-modal", "display"); got != "flex" {
		t.Errorf("open static overlay display = %q, want flex", got)
	}
	ovWaitFocus(t, page, "#ov-modal .x-btn", "static open focus")

	// One at a time: opening the drawer (synthetic click — the backdrop
	// covers the real button) force-drops the modal's .open.
	if _, err := page.Evaluate(`() => document.getElementById('ov-open-drawer').click()`); err != nil {
		t.Fatalf("synthetic drawer-opener click: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#ov-drawer').classList.contains('open')
		    && !document.querySelector('#ov-modal').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("one-overlay-at-a-time never enforced: %v", err)
	}

	// Single-focusable trap: Tab cycles the drawer's ✕ onto itself.
	ovWaitFocus(t, page, "#ov-drawer .x-btn", "drawer open focus")
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	ovWaitFocus(t, page, "#ov-drawer .x-btn", "single-element Tab cycle")

	// Esc: static overlay drops .open but stays in the DOM.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#ov-drawer').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("Esc never dropped .open: %v", err)
	}
	if n, err := page.Locator("#ov-drawer").Count(); err != nil || n != 1 {
		t.Fatalf("static drawer left the DOM on close (count=%d, err=%v)", n, err)
	}

	// Clean open/Esc cycle restores opener focus.
	if err := page.Locator("#ov-open-modal").Click(); err != nil {
		t.Fatalf("re-open static modal: %v", err)
	}
	ovWaitFocus(t, page, "#ov-modal .x-btn", "re-open focus")
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#ov-modal').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("Esc never closed the static modal: %v", err)
	}
	ovWaitFocus(t, page, "#ov-open-modal", "static Esc focus restore")
}

// TestOverlayModalStylesAndTheme: the design constants — 540px at 9vh,
// strong border, deep 0 24px 60px --shadow, zero radius, uppercase 700
// title, right-aligned raised footer, translucent backdrop — and the
// t-mocha root-class swap re-resolving border/fill from tokens alone.
func TestOverlayModalStylesAndTheme(t *testing.T) {
	page, _ := ovOpenSheet(t)
	if err := page.Locator("#ov-open-modal").Click(); err != nil {
		t.Fatalf("open static modal: %v", err)
	}
	ovWaitFocus(t, page, "#ov-modal .x-btn", "open focus")

	m := "#ov-modal .modal"
	if got := computedStyleSel(t, page, m, "width"); got != "540px" {
		t.Errorf("modal width = %q, want 540px", got)
	}
	ihv, err := page.Evaluate(`() => innerHeight`)
	if err != nil {
		t.Fatalf("innerHeight: %v", err)
	}
	ih := ovNum(t, ihv, "innerHeight")
	if got := ovPx(t, computedStyleSel(t, page, m, "marginTop")); math.Abs(got-0.09*ih) > 1.5 {
		t.Errorf("modal margin-top = %vpx, want 9vh = %vpx", got, 0.09*ih)
	}
	if got := computedStyleSel(t, page, m, "borderTopWidth"); got != "1px" {
		t.Errorf("modal border width = %q, want 1px", got)
	}
	if got := computedStyleSel(t, page, m, "borderTopColor"); got != ovGruvStrong {
		t.Errorf("modal border color = %q, want %s (--border-strong)", got, ovGruvStrong)
	}
	if got := computedStyleSel(t, page, m, "backgroundColor"); got != ovGruvPanel {
		t.Errorf("modal background = %q, want %s (--bg-panel)", got, ovGruvPanel)
	}
	shadow := computedStyleSel(t, page, m, "boxShadow")
	if !strings.Contains(shadow, "24px") || !strings.Contains(shadow, "60px") {
		t.Errorf("modal box-shadow = %q, want the deep 0 24px 60px --shadow", shadow)
	}
	if got := computedStyleSel(t, page, m, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("modal border-radius = %q, want 0px", got)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal-title", "textTransform"); got != "uppercase" {
		t.Errorf("title text-transform = %q, want uppercase", got)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal-title", "fontWeight"); got != "700" {
		t.Errorf("title font-weight = %q, want 700", got)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal-ft", "justifyContent"); got != "flex-end" {
		t.Errorf("footer justify-content = %q, want flex-end (right-aligned actions)", got)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal-ft", "backgroundColor"); got != ovGruvRaised {
		t.Errorf("footer background = %q, want %s (--bg-raised)", got, ovGruvRaised)
	}
	// Backdrop: --shadow at 60% — translucent black, neither clear nor solid.
	assertToneAlpha(t, computedStyleSel(t, page, "#ov-modal", "backgroundColor"), "0, 0, 0", "modal backdrop")

	// Token flow: t-mocha by root-class swap only.
	if _, err := page.Evaluate(`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if _, err := page.WaitForFunction(
		`sel => getComputedStyle(document.querySelector(sel)).borderTopColor === '`+ovMochaStrong+`'`, m,
	); err != nil {
		got := computedStyleSel(t, page, m, "borderTopColor")
		t.Errorf("after t-mocha swap border = %q, want %s: %v", got, ovMochaStrong, err)
	}
	if got := computedStyleSel(t, page, m, "backgroundColor"); got != ovMochaPanel {
		t.Errorf("after t-mocha swap background = %q, want %s", got, ovMochaPanel)
	}
}

// TestOverlayDrawerSlidesFromRight: the drawer pins to the right edge at
// full height and 420px width, carries the strong left hairline and the
// leftward deep shadow, and enters via the 160ms drawer-in animation.
func TestOverlayDrawerSlidesFromRight(t *testing.T) {
	page, _ := ovOpenSheet(t)
	if err := page.Locator("#ov-open-drawer").Click(); err != nil {
		t.Fatalf("open static drawer: %v", err)
	}
	ovWaitFocus(t, page, "#ov-drawer .x-btn", "open focus")

	d := "#ov-drawer .drawer"
	if got := computedStyleSel(t, page, d, "position"); got != "absolute" {
		t.Errorf("drawer position = %q, want absolute", got)
	}
	if got := computedStyleSel(t, page, d, "width"); got != "420px" {
		t.Errorf("drawer width = %q, want 420px", got)
	}
	// Let the 160ms drawer-in entrance settle before measuring geometry —
	// mid-animation the transform still offsets the rect.
	if _, err := page.WaitForFunction(
		`sel => document.querySelector(sel).getAnimations().every(a => a.playState === 'finished')`, d,
	); err != nil {
		t.Fatalf("drawer entrance animation never finished: %v", err)
	}
	v, err := page.Evaluate(`sel => { const r = document.querySelector(sel).getBoundingClientRect(); return { left: r.left, right: r.right, top: r.top, bottom: r.bottom, iw: innerWidth, ih: innerHeight }; }`, d)
	if err != nil {
		t.Fatalf("drawer rect: %v", err)
	}
	r, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("drawer rect: non-map %T", v)
	}
	num := func(k string) float64 { return ovNum(t, r[k], "drawer rect "+k) }
	if math.Abs(num("right")-num("iw")) > 0.5 || math.Abs(num("left")-(num("iw")-420)) > 0.5 {
		t.Errorf("drawer not pinned to the right edge: left=%v right=%v innerWidth=%v", num("left"), num("right"), num("iw"))
	}
	if num("top") != 0 || math.Abs(num("bottom")-num("ih")) > 0.5 {
		t.Errorf("drawer not full height: top=%v bottom=%v innerHeight=%v", num("top"), num("bottom"), num("ih"))
	}
	if got := computedStyleSel(t, page, d, "borderLeftWidth"); got != "1px" {
		t.Errorf("drawer border-left width = %q, want 1px", got)
	}
	if got := computedStyleSel(t, page, d, "borderLeftColor"); got != ovGruvStrong {
		t.Errorf("drawer border-left color = %q, want %s (--border-strong)", got, ovGruvStrong)
	}
	if shadow := computedStyleSel(t, page, d, "boxShadow"); !strings.Contains(shadow, "-18px") {
		t.Errorf("drawer box-shadow = %q, want the leftward -18px 0 50px --shadow", shadow)
	}
	if got := computedStyleSel(t, page, d, "animationName"); got != "drawer-in" {
		t.Errorf("drawer animation-name = %q, want drawer-in", got)
	}
	if got := computedStyleSel(t, page, d, "animationDuration"); got != "0.16s" {
		t.Errorf("drawer animation-duration = %q, want 0.16s", got)
	}
	// Drawer backdrop is the lighter 45% --shadow mix — still translucent.
	assertToneAlpha(t, computedStyleSel(t, page, "#ov-drawer", "backgroundColor"), "0, 0, 0", "drawer backdrop")
}

// TestOverlayReducedMotionKillsEntrance: the 160ms entrances exist only
// inside @media (prefers-reduced-motion: no-preference).
func TestOverlayReducedMotionKillsEntrance(t *testing.T) {
	page, _ := ovOpenSheet(t)
	if err := page.Locator("#ov-open-modal").Click(); err != nil {
		t.Fatalf("open static modal: %v", err)
	}
	ovWaitFocus(t, page, "#ov-modal .x-btn", "open focus")
	if got := computedStyleSel(t, page, "#ov-modal .modal", "animationName"); got != "modal-in" {
		t.Errorf("modal animation-name = %q, want modal-in", got)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal", "animationDuration"); got != "0.16s" {
		t.Errorf("modal animation-duration = %q, want 0.16s", got)
	}

	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}); err != nil {
		t.Fatalf("emulate reduced motion: %v", err)
	}
	if got := computedStyleSel(t, page, "#ov-modal .modal", "animationName"); got != "none" {
		t.Errorf("modal animates under reduced motion: animation-name = %q, want none", got)
	}
}

//go:build e2e

package e2e

// fleetctl console browser coverage — the demo's golden path over the
// composition in demo/app.templ + demo/fleetctl_api.go: table sort
// round-trip, inspect drawer with the ConfirmInput kill guard and its OOB
// toast, the deploy modal (mouse and ⌘K-palette-driven), topbar tab view
// swaps, the lazy navigator branch, theme/density root-class swaps and
// reduced-motion sanity. Helpers here are fleetctl-prefixed; generic
// bootstrap lives in helpers_test.go.

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved tokens asserted below (assets/css/tokens.css):
const (
	fleetGruvErr  = "rgb(251, 73, 52)"   // t-gruvbox --err #fb4934
	fleetMochaErr = "rgb(243, 139, 168)" // t-mocha --err #f38ba8
)

// fleetctlOpen starts the demo, loads the console at / and gates on the
// ParseHealth sentinel so every hyperscript behavior is live before any
// interaction.
func fleetctlOpen(t *testing.T) (playwright.Page, string) {
	t.Helper()
	srv := startDemo(t)
	page := startBrowser(t)
	if _, err := page.Goto(srv.URL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /: %v", err)
	}
	if err := page.Locator("body[data-hs-ok]").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("body[data-hs-ok] never appeared — baud._hs parse error?: %v", err)
	}
	return page, srv.URL
}

// fleetctlFirstSvc reads the first body row's service cell (the cell
// after the row-mark).
func fleetctlFirstSvc(t *testing.T, page playwright.Page) string {
	t.Helper()
	v, err := page.Evaluate(
		`() => { const td = document.querySelector('#fleet-hosts-body tr:first-child td:nth-child(2)'); return td ? td.textContent.trim() : 'NO ROW'; }`)
	if err != nil {
		t.Fatalf("read first svc cell: %v", err)
	}
	s, _ := v.(string)
	return s
}

// fleetctlWaitFirstSvc waits until the table's first service cell equals
// want — the honest signal a sort swap settled.
func fleetctlWaitFirstSvc(t *testing.T, page playwright.Page, want, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(
		`want => { const td = document.querySelector('#fleet-hosts-body tr:first-child td:nth-child(2)'); return td && td.textContent.trim() === want; }`,
		want,
	); err != nil {
		t.Fatalf("%s: first service cell never became %q (now %q): %v", what, want, fleetctlFirstSvc(t, page), err)
	}
}

// fleetctlWaitHeadReady gates on htmx having initialized the (possibly
// OOB-swapped) cpu header: the swap inserts the new thead immediately but
// htmx binds its hx-get listeners only at settle (~20ms later), so clicks
// must wait for the internal-data marker, not the attributes (the
// datatableWaitHeadReady precedent). COUPLING: peeks htmx's PRIVATE
// htmx-internal-data property, valid for the pinned htmx 2.0.4
// (baud/baud.go) — revisit on any htmx upgrade.
func fleetctlWaitHeadReady(t *testing.T, page playwright.Page, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(
		`() => {
			const th = document.querySelector('#fleet-hosts-head th[data-key=cpu]');
			return th && !!th['htmx-internal-data'];
		}`, nil,
	); err != nil {
		t.Fatalf("%s: thead never htmx-initialized: %v", what, err)
	}
}

// fleetctlWaitOverlayReady waits for an htmx-injected overlay to land AND
// its Overlay listeners to attach (.overlay-ready signal).
func fleetctlWaitOverlayReady(t *testing.T, page playwright.Page, sel string) {
	t.Helper()
	if err := page.Locator(sel + ".overlay-ready").WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateAttached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("injected overlay %s never became ready: %v", sel, err)
	}
}

// fleetctlWaitDetached waits for an element to leave the DOM.
func fleetctlWaitDetached(t *testing.T, page playwright.Page, sel string) {
	t.Helper()
	if err := page.Locator(sel).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("%s never left the DOM: %v", sel, err)
	}
}

// fleetctlWaitFocus waits until document.activeElement matches sel.
func fleetctlWaitFocus(t *testing.T, page playwright.Page, sel, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(
		`sel => document.activeElement && document.activeElement.matches(sel)`, sel,
	); err != nil {
		got, _ := page.Evaluate(`() => document.activeElement && (document.activeElement.id || document.activeElement.tagName)`)
		t.Fatalf("%s: activeElement never matched %q (now on %v): %v", what, sel, got, err)
	}
}

// fleetctlCountRequests counts same-origin requests whose URL contains
// frag from this point on.
func fleetctlCountRequests(page playwright.Page, base, frag string) *int32 {
	var n int32
	page.OnRequest(func(r playwright.Request) {
		if strings.HasPrefix(r.URL(), base) && strings.Contains(r.URL(), frag) {
			atomic.AddInt32(&n, 1)
		}
	})
	return &n
}

// fleetctlAttr reads an attribute of the first match ("" when absent).
func fleetctlAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Evaluate(
		`([sel, name]) => { const el = document.querySelector(sel); if (!el) return 'NO ELEMENT'; return el.getAttribute(name) ?? ''; }`,
		[]string{sel, name})
	if err != nil {
		t.Fatalf("attribute %s of %s: %v", name, sel, err)
	}
	s, _ := v.(string)
	if s == "NO ELEMENT" {
		t.Fatalf("no element matches %q", sel)
	}
	return s
}

// fleetctlKillBtnDisabled reads the drawer kill button's disabled state.
func fleetctlKillBtnDisabled(t *testing.T, page playwright.Page) bool {
	t.Helper()
	v, err := page.Evaluate(`() => document.querySelector('#fleet-kill .btn-danger').disabled`)
	if err != nil {
		t.Fatalf("read kill button disabled: %v", err)
	}
	b, _ := v.(bool)
	return b
}

// TestFleetctlConsoleSortRoundTrip: the console boots with the worst error
// rate on top (err desc — server state), a cpu header click round-trips
// /fleet/hosts and re-sorts asc with the OOB thead tracking aria-sort,
// and a second click flips to desc.
func TestFleetctlConsoleSortRoundTrip(t *testing.T) {
	page, base := fleetctlOpen(t)

	// Mode cell renders the vim-style accent block — the console booted.
	modeBg := computedStyleSel(t, page, "[data-statusbar] .sb-mode", "backgroundColor")
	if modeBg != gruvAccent {
		t.Errorf("statusbar mode cell background = %q, want %s (--accent)", modeBg, gruvAccent)
	}

	if got := fleetctlFirstSvc(t, page); got != "ingest-gw" {
		t.Fatalf("initial first row = %q, want ingest-gw (err desc default)", got)
	}
	if got := fleetctlAttr(t, page, "#fleet-hosts-head th[data-key=err]", "aria-sort"); got != "descending" {
		t.Errorf("initial err aria-sort = %q, want descending", got)
	}

	reqs := fleetctlCountRequests(page, base, "/fleet/hosts?sort=")
	fleetctlWaitHeadReady(t, page, "initial header")
	if err := page.Locator("#fleet-hosts-head th[data-key=cpu]").Click(); err != nil {
		t.Fatalf("click cpu header: %v", err)
	}
	fleetctlWaitFirstSvc(t, page, "flag-svc", "cpu asc")
	if _, err := page.WaitForFunction(
		`() => { const th = document.querySelector('#fleet-hosts-head th[data-key=cpu]'); return th && th.getAttribute('aria-sort') === 'ascending'; }`, nil,
	); err != nil {
		t.Fatalf("OOB thead never marked cpu ascending: %v", err)
	}

	fleetctlWaitHeadReady(t, page, "OOB-swapped header")
	if err := page.Locator("#fleet-hosts-head th[data-key=cpu]").Click(); err != nil {
		t.Fatalf("click cpu header again: %v", err)
	}
	fleetctlWaitFirstSvc(t, page, "ingest-gw", "cpu desc")
	if _, err := page.WaitForFunction(
		`() => { const th = document.querySelector('#fleet-hosts-head th[data-key=cpu]'); return th && th.getAttribute('aria-sort') === 'descending'; }`, nil,
	); err != nil {
		t.Fatalf("OOB thead never flipped cpu descending: %v", err)
	}
	if n := atomic.LoadInt32(reqs); n != 2 {
		t.Errorf("two header clicks made %d sort request(s), want 2", n)
	}

	// The selection is server state: ingest-gw keeps the ▌ mark after swaps.
	mark, err := page.Locator("#fleet-hosts-body tr.sel td.row-mark").TextContent()
	if err != nil || strings.TrimSpace(mark) != "▌" {
		t.Errorf("selected row mark = %q (err=%v), want ▌", mark, err)
	}
}

// TestFleetctlDrawerKillGuardToast — the golden path: inspect opens the
// host drawer (htmx body/beforeend), the DiffViewer shows the config
// change, the ConfirmInput guard holds until the exact host name is typed,
// the kill action round-trips /fleet/kill whose OOB toast lands in
// #toasts, ✕ dismisses it, and Esc closes the drawer restoring focus.
func TestFleetctlDrawerKillGuardToast(t *testing.T) {
	page, base := fleetctlOpen(t)

	if err := page.Locator("#fleet-inspect-btn").Click(); err != nil {
		t.Fatalf("click inspect: %v", err)
	}
	fleetctlWaitOverlayReady(t, page, "#fleet-drawer")

	// Dialog ARIA + title.
	if got := fleetctlAttr(t, page, "#fleet-drawer .drawer", "role"); got != "dialog" {
		t.Errorf("drawer role = %q, want dialog", got)
	}
	if got := fleetctlAttr(t, page, "#fleet-drawer .drawer", "aria-labelledby"); got != "fleet-drawer-title" {
		t.Errorf("aria-labelledby = %q, want fleet-drawer-title", got)
	}
	title, err := page.Locator("#fleet-drawer-title").TextContent()
	if err != nil || strings.TrimSpace(title) != "ingest-gw · v2.14.0" {
		t.Errorf("drawer title = %q (err=%v), want ingest-gw · v2.14.0", title, err)
	}

	// The config-change DiffViewer rendered its add/del rows.
	if n, _ := page.Locator("#fleet-drawer .diff-line.add").Count(); n != 3 {
		t.Errorf("diff add rows = %d, want 3", n)
	}
	if n, _ := page.Locator("#fleet-drawer .diff-line.del").Count(); n != 2 {
		t.Errorf("diff del rows = %d, want 2", n)
	}

	// Guard: disabled until the exact name; mismatch shows the err border.
	if !fleetctlKillBtnDisabled(t, page) {
		t.Fatalf("kill button must start disabled")
	}
	input := page.Locator("#fleet-confirm-input")
	if err := input.Fill("wrong-host"); err != nil {
		t.Fatalf("fill mismatch: %v", err)
	}
	if err := page.Locator("#fleet-kill .input-wrap.err").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("mismatch never flagged the err border: %v", err)
	}
	if !fleetctlKillBtnDisabled(t, page) {
		t.Fatalf("kill button enabled on a mismatch")
	}
	if err := input.Fill("ingest-gw"); err != nil {
		t.Fatalf("fill exact match: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#fleet-kill .btn-danger').disabled`, nil,
	); err != nil {
		t.Fatalf("exact match never enabled the kill button: %v", err)
	}
	if err := page.Locator("#fleet-kill .input-wrap:not(.err)").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("exact match never cleared the err border: %v", err)
	}

	// Fire: exactly one /fleet/kill request, the ok toast lands via OOB.
	reqs := fleetctlCountRequests(page, base, "/fleet/kill")
	if err := page.Locator("#fleet-kill .btn-danger").Click(); err != nil {
		t.Fatalf("click kill: %v", err)
	}
	if err := page.Locator("#toasts .toast.tone-ok").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("OOB toast never landed: %v", err)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("kill made %d request(s), want 1", n)
	}
	toastTitle, err := page.Locator("#toasts .toast .toast-title").TextContent()
	if err != nil || !strings.Contains(toastTitle, "pod killed: ingest-gw") {
		t.Errorf("toast title = %q (err=%v), want pod killed: ingest-gw…", toastTitle, err)
	}

	// ✕ dismisses the toast; Esc closes the drawer and restores focus.
	if err := page.Locator("#toasts .toast .x-btn").Click(); err != nil {
		t.Fatalf("click toast ✕: %v", err)
	}
	fleetctlWaitDetached(t, page, "#toasts .toast")
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	fleetctlWaitDetached(t, page, "#fleet-drawer")
	fleetctlWaitFocus(t, page, "#fleet-inspect-btn", "drawer Esc focus restore")
}

// TestFleetctlDeployModalMouseAndEsc: the topbar deploy button hx-gets the
// confirm modal; Esc removes it and restores opener focus; reopened, the
// confirm action fires /fleet/deploy/run (OOB toast) and a successful
// htmx:afterRequest sends baud:overlayClose to remove the modal.
func TestFleetctlDeployModalMouseAndEsc(t *testing.T) {
	page, base := fleetctlOpen(t)

	if err := page.Locator("#fleet-deploy-btn").Click(); err != nil {
		t.Fatalf("click deploy: %v", err)
	}
	fleetctlWaitOverlayReady(t, page, "#fleet-modal")
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	fleetctlWaitDetached(t, page, "#fleet-modal")
	fleetctlWaitFocus(t, page, "#fleet-deploy-btn", "modal Esc focus restore")

	if err := page.Locator("#fleet-deploy-btn").Click(); err != nil {
		t.Fatalf("re-open deploy modal: %v", err)
	}
	fleetctlWaitOverlayReady(t, page, "#fleet-modal")
	reqs := fleetctlCountRequests(page, base, "/fleet/deploy/run")
	if err := page.Locator("#fleet-deploy-run").Click(); err != nil {
		t.Fatalf("click confirm deploy: %v", err)
	}
	fleetctlWaitDetached(t, page, "#fleet-modal")
	if err := page.Locator("#toasts .toast.tone-ok").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("deploy OOB toast never landed: %v", err)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("confirm made %d /fleet/deploy/run request(s), want 1", n)
	}
	toastTitle, err := page.Locator("#toasts .toast .toast-title").TextContent()
	if err != nil || !strings.Contains(toastTitle, "deploy queued") {
		t.Errorf("toast title = %q (err=%v), want deploy queued…", toastTitle, err)
	}
}

// TestFleetctlPaletteRunsDeployCommand: a real ⌘K opens the console
// palette, typing filters through /fleet/palette (debounced hx-get), ↵ on
// the action row dispatches baud:paletteCmd — the topbar deploy button's
// listener replays its click and the modal arrives. The palette closed
// itself on activation.
func TestFleetctlPaletteRunsDeployCommand(t *testing.T) {
	page, _ := fleetctlOpen(t)

	if err := page.Locator("#fleet-cmd-btn").Focus(); err != nil {
		t.Fatalf("focus cmd button: %v", err)
	}
	if err := page.Keyboard().Press("Control+k"); err != nil {
		t.Fatalf("press Ctrl-K: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#fleet-palette').classList.contains('open')
			&& document.activeElement && document.activeElement.id === 'fleet-palette-input'`, nil,
	); err != nil {
		t.Fatalf("Ctrl-K never opened the console palette: %v", err)
	}
	if n, _ := page.Locator("#fleet-palette .palette-item").Count(); n != 7 {
		t.Fatalf("seeded command count = %d, want 7", n)
	}

	if err := page.Locator("#fleet-palette-input").PressSequentially("deploy fleet"); err != nil {
		t.Fatalf("type deploy fleet: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#fleet-palette .palette-item').length === 1`, nil,
	); err != nil {
		t.Fatalf("server filter never narrowed to the deploy command: %v", err)
	}
	if got := fleetctlAttr(t, page, "#fleet-palette .palette-item", "data-cmd"); got != "fleet-deploy" {
		t.Fatalf("filtered row data-cmd = %q, want fleet-deploy", got)
	}

	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	fleetctlWaitOverlayReady(t, page, "#fleet-modal")
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#fleet-palette').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("palette never closed after activation: %v", err)
	}
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	fleetctlWaitDetached(t, page, "#fleet-modal")
}

// TestFleetctlPaletteRestartToast: ⌘K → "restart ingest workers" must
// have a visible console effect (acceptance-audit finding: no palette
// no-ops): the hidden relay replays its click, /fleet/cmd?op=restart-
// ingest round-trips, and the OOB toast lands.
func TestFleetctlPaletteRestartToast(t *testing.T) {
	page, base := fleetctlOpen(t)

	if err := page.Locator("#fleet-cmd-btn").Click(); err != nil {
		t.Fatalf("click cmd button: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#fleet-palette').classList.contains('open')
			&& document.activeElement && document.activeElement.id === 'fleet-palette-input'`, nil,
	); err != nil {
		t.Fatalf("cmd button never opened the console palette: %v", err)
	}

	if err := page.Locator("#fleet-palette-input").PressSequentially("restart"); err != nil {
		t.Fatalf("type restart: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#fleet-palette .palette-item').length === 1`, nil,
	); err != nil {
		t.Fatalf("server filter never narrowed to the restart command: %v", err)
	}
	if got := fleetctlAttr(t, page, "#fleet-palette .palette-item", "data-cmd"); got != "restart-ingest" {
		t.Fatalf("filtered row data-cmd = %q, want restart-ingest", got)
	}

	reqs := fleetctlCountRequests(page, base, "/fleet/cmd?op=restart-ingest")
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if err := page.Locator("#toasts .toast.tone-ok").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("restart OOB toast never landed: %v", err)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("restart command made %d /fleet/cmd request(s), want 1", n)
	}
	toastTitle, err := page.Locator("#toasts .toast .toast-title").TextContent()
	if err != nil || !strings.Contains(toastTitle, "restart queued: ingest-gw") {
		t.Errorf("toast title = %q (err=%v), want restart queued: ingest-gw…", toastTitle, err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#fleet-palette').classList.contains('open')`, nil,
	); err != nil {
		t.Fatalf("palette never closed after activation: %v", err)
	}
}

// TestFleetctlMetricValueHierarchy: the metrics strip's big numbers carry
// the prototype MetricCell hierarchy — 1.5× base size (18px at d-dense),
// weight 700, tone-coloured (breaching p99 in --err) — and rescale from a
// d-cozy root-class swap alone (13px × 1.5 = 19.5px).
func TestFleetctlMetricValueHierarchy(t *testing.T) {
	page, _ := fleetctlOpen(t)

	if got := computedStyleSel(t, page, "dd.dl-v > .fl-metric", "fontSize"); got != "18px" {
		t.Errorf("d-dense metric value font-size = %q, want 18px (--fs × 1.5)", got)
	}
	if got := computedStyleSel(t, page, "dd.dl-v > .fl-metric", "fontWeight"); got != "700" {
		t.Errorf("metric value font-weight = %q, want 700", got)
	}
	if got := computedStyleSel(t, page, "dd.dl-v > .fl-metric.fl-err", "color"); got != fleetGruvErr {
		t.Errorf("breaching p99 metric color = %q, want %s (--err)", got, fleetGruvErr)
	}

	if _, err := page.Evaluate(`() => document.body.classList.replace('d-dense', 'd-cozy')`); err != nil {
		t.Fatalf("swap density class: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('dd.dl-v > .fl-metric')).fontSize === '19.5px'`, nil,
	); err != nil {
		got := computedStyleSel(t, page, "dd.dl-v > .fl-metric", "fontSize")
		t.Fatalf("d-cozy metric value font-size stayed %q, want 19.5px: %v", got, err)
	}
}

// TestFleetctlTabsNoJS — graceful degradation, scripting disabled (the
// suite's first js-off context): the topbar tabs are real links, so
// clicking incidents navigates to /?tab=incidents and the server renders
// that pane active with correct aria-selected. No htmx, no hyperscript.
func TestFleetctlTabsNoJS(t *testing.T) {
	srv := startDemo(t)
	page := startBrowserNoJS(t)

	if _, err := page.Goto(srv.URL+"/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /: %v", err)
	}
	// Scripting is genuinely off: the ParseHealth sentinel never runs.
	if n, _ := page.Locator("body[data-hs-ok]").Count(); n != 0 {
		t.Fatalf("data-hs-ok present — JavaScript ran in the js-off context")
	}
	if n, _ := page.Locator("#fleet-hosts").Count(); n != 1 {
		t.Fatalf("default fleet pane missing under js-off")
	}

	// Keyboard reachability without scripting: roving tabindex is
	// hyperscript, so anchor tabs must all sit in the native tab order.
	if err := page.Locator("#fleet-tabs-tab-0").Focus(); err != nil {
		t.Fatalf("focus active tab: %v", err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	active, err := page.Evaluate(`() => document.activeElement && document.activeElement.id`)
	if err != nil {
		t.Fatalf("read activeElement: %v", err)
	}
	if active != "fleet-tabs-tab-1" {
		t.Fatalf("js-off Tab from tab 0 landed on %v, want fleet-tabs-tab-1 (inactive tabs must be keyboard-reachable)", active)
	}
	if err := page.Locator("#fleet-tabs-tab-1").Click(); err != nil {
		t.Fatalf("click incidents tab link: %v", err)
	}
	if err := page.WaitForURL(srv.URL + "/?tab=incidents"); err != nil {
		t.Fatalf("incidents tab never navigated to /?tab=incidents: %v", err)
	}
	if err := page.Locator("#fleet-view #fleet-incidents").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("incidents pane never rendered server-side: %v", err)
	}
	if n, _ := page.Locator("#fleet-hosts").Count(); n != 0 {
		t.Errorf("fleet hosts table still present on the incidents view")
	}
	if got := fleetctlAttr(t, page, "#fleet-tabs-tab-1", "aria-selected"); got != "true" {
		t.Errorf("incidents tab aria-selected = %q, want true", got)
	}
	if got := fleetctlAttr(t, page, "#fleet-tabs-tab-0", "aria-selected"); got != "false" {
		t.Errorf("fleet tab aria-selected = %q, want false", got)
	}

	// And back: the fleet tab link restores the default view.
	if err := page.Locator("#fleet-tabs-tab-0").Click(); err != nil {
		t.Fatalf("click fleet tab link: %v", err)
	}
	if err := page.WaitForURL(srv.URL + "/?tab=fleet"); err != nil {
		t.Fatalf("fleet tab never navigated to /?tab=fleet: %v", err)
	}
	if n, _ := page.Locator("#fleet-hosts").Count(); n != 1 {
		t.Errorf("fleet hosts table missing after navigating back")
	}
}

// TestFleetctlTabsSwapViews: the topbar tabs hx-get their pane into the
// shared #fleet-view tabpanel — incidents replaces the fleet table,
// returning to fleet re-renders it (fresh server state).
func TestFleetctlTabsSwapViews(t *testing.T) {
	page, _ := fleetctlOpen(t)

	if err := page.Locator("#fleet-tabs-tab-1").Click(); err != nil {
		t.Fatalf("click incidents tab: %v", err)
	}
	if err := page.Locator("#fleet-view #fleet-incidents").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("incidents pane never swapped in: %v", err)
	}
	fleetctlWaitDetached(t, page, "#fleet-hosts")
	if got := fleetctlAttr(t, page, "#fleet-tabs-tab-1", "aria-selected"); got != "true" {
		t.Errorf("incidents tab aria-selected = %q, want true", got)
	}
	if got := fleetctlAttr(t, page, "#fleet-tabs-tab-0", "aria-selected"); got != "false" {
		t.Errorf("fleet tab aria-selected = %q, want false", got)
	}

	if err := page.Locator("#fleet-tabs-tab-0").Click(); err != nil {
		t.Fatalf("click fleet tab: %v", err)
	}
	if err := page.Locator("#fleet-view #fleet-hosts").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		t.Fatalf("fleet pane never swapped back: %v", err)
	}
	if got := fleetctlFirstSvc(t, page); got != "ingest-gw" {
		t.Errorf("re-rendered table first row = %q, want ingest-gw (default sort)", got)
	}
}

// TestFleetctlTreeLazyBranch: the navigator's edge branch fetches its
// children from /fleet/tree on first expand only.
func TestFleetctlTreeLazyBranch(t *testing.T) {
	page, base := fleetctlOpen(t)
	reqs := fleetctlCountRequests(page, base, "/fleet/tree")

	lazy := "#fleet-tree details[hx-get]"
	if n, _ := page.Locator(lazy + " > ul > li").Count(); n != 0 {
		t.Fatalf("lazy branch has %d pre-rendered children, want 0", n)
	}
	if err := page.Locator(lazy + " > summary").Click(); err != nil {
		t.Fatalf("click lazy branch: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#fleet-tree details[hx-get] > ul > li.tree-row').length === 3`, nil,
	); err != nil {
		t.Fatalf("lazy children never swapped in: %v", err)
	}
	if txt, _ := page.Locator(lazy + " > ul").TextContent(); !strings.Contains(txt, "edge-cache-ams") {
		t.Errorf("lazy group = %q, want edge-cache-ams row", txt)
	}
	if n := atomic.LoadInt32(reqs); n != 1 {
		t.Errorf("first expand made %d /fleet/tree request(s), want 1", n)
	}
}

// TestFleetctlThemeAndDensityRootSwap: the whole console re-tokens from
// root-class swaps alone — t-mocha re-resolves the mode cell accent and
// the log tail's err tone, d-cozy re-resolves the table row height.
func TestFleetctlThemeAndDensityRootSwap(t *testing.T) {
	page, _ := fleetctlOpen(t)

	if got := computedStyleSel(t, page, "[data-statusbar] .sb-mode", "backgroundColor"); got != gruvAccent {
		t.Errorf("gruvbox mode cell background = %q, want %s", got, gruvAccent)
	}
	if got := computedStyleSel(t, page, ".fleet-log-line .fl-lvl.fl-err", "color"); got != fleetGruvErr {
		t.Errorf("gruvbox log err level color = %q, want %s", got, fleetGruvErr)
	}
	if got := computedStyleSel(t, page, "#fleet-hosts-body td", "height"); got != "22px" {
		t.Errorf("d-dense table row height = %q, want 22px (--row)", got)
	}

	if _, err := page.Evaluate(`() => document.body.classList.replace('t-gruvbox', 't-mocha')`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('[data-statusbar] .sb-mode')).backgroundColor === '`+mochaAccent+`'`, nil,
	); err != nil {
		got := computedStyleSel(t, page, "[data-statusbar] .sb-mode", "backgroundColor")
		t.Fatalf("t-mocha mode cell background stayed %q, want %s: %v", got, mochaAccent, err)
	}
	if got := computedStyleSel(t, page, ".fleet-log-line .fl-lvl.fl-err", "color"); got != fleetMochaErr {
		t.Errorf("t-mocha log err level color = %q, want %s", got, fleetMochaErr)
	}

	if _, err := page.Evaluate(`() => document.body.classList.replace('d-dense', 'd-cozy')`); err != nil {
		t.Fatalf("swap density class: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#fleet-hosts-body td')).height === '27px'`, nil,
	); err != nil {
		got := computedStyleSel(t, page, "#fleet-hosts-body td", "height")
		t.Fatalf("d-cozy table row height stayed %q, want 27px: %v", got, err)
	}
}

// TestFleetctlReducedMotion: the health panel's pulse dot animates only
// under prefers-reduced-motion: no-preference.
func TestFleetctlReducedMotion(t *testing.T) {
	page, _ := fleetctlOpen(t)

	if got := computedStyleSel(t, page, "#fleet-m-health .dot.pulse", "animationName"); got != "baud-pulse" {
		t.Errorf("pulse dot animation-name = %q, want baud-pulse", got)
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}); err != nil {
		t.Fatalf("emulate reduced motion: %v", err)
	}
	if got := computedStyleSel(t, page, "#fleet-m-health .dot.pulse", "animationName"); got != "none" {
		t.Errorf("pulse dot animates under reduced motion: animation-name = %q, want none", got)
	}
}

// TestFleetctlEndpointsRejectGarbage: every console endpoint validates its
// parameters — garbage is a 400, the happy path a 200.
func TestFleetctlEndpointsRejectGarbage(t *testing.T) {
	srv := startDemo(t)
	for _, q := range []string{
		"/fleet/hosts?sort=nope&dir=asc",
		"/fleet/hosts?sort=cpu&dir=sideways",
		"/fleet/hosts?sort=cpu",
		"/fleet/tab?view=nope",
		"/fleet/tab",
		"/fleet/host?id=nope",
		"/fleet/host",
		"/fleet/kill?host=nope",
		"/fleet/tree?node=nope",
		"/fleet/cmd?op=nope",
		"/fleet/cmd",
		"/?tab=nope",
	} {
		res, err := srv.Client().Get(srv.URL + q)
		if err != nil {
			t.Fatalf("GET %s: %v", q, err)
		}
		res.Body.Close()
		if res.StatusCode != 400 {
			t.Errorf("GET %s = %d, want 400", q, res.StatusCode)
		}
	}
	for _, q := range []string{
		"/fleet/hosts?sort=cpu&dir=asc",
		"/fleet/tab?view=incidents",
		"/fleet/host?id=ingest-gw",
		"/fleet/kill?host=ingest-gw",
		"/fleet/tree?node=euw1/edge",
		"/fleet/deploy",
		"/fleet/deploy/run",
		"/fleet/palette?q=deploy",
		"/fleet/cmd?op=restart-ingest",
		"/fleet/cmd?op=drain-batch-runner",
		"/?tab=incidents",
	} {
		res, err := srv.Client().Get(srv.URL + q)
		if err != nil {
			t.Fatalf("GET %s: %v", q, err)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Errorf("GET %s = %d, want 200", q, res.StatusCode)
		}
	}
}

//go:build e2e

// Package e2e runs real-browser assertions with playwright-go against the
// demo handler mounted on a random-port test server. Gated behind the e2e
// build tag; run via `just e2e` (after `just install-browsers`).
package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// sheetPanes is the resizable Panes demo on the component sheet
// (demo/sheet_layout.templ): Template "32ch 1fr", Resizable, so the widened
// data-panes spec is "32ch 7px 1fr".
const sheetPanes = `[data-panes-id="sheet-panes"]`

// waitForPanesTemplate polls until the Panes hyperscript behavior has applied
// the grid template, then returns the resolved (computed) column tracks.
func waitForPanesTemplate(t *testing.T, page playwright.Page) []string {
	t.Helper()
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => {
			const el = document.querySelector(%q);
			return el && getComputedStyle(el).gridTemplateColumns !== 'none';
		}`, sheetPanes), nil,
	); err != nil {
		t.Fatalf("Panes behavior never set grid-template-columns: %v", err)
	}
	return strings.Fields(computedStyleSel(t, page, sheetPanes, "gridTemplateColumns"))
}

// pxValue parses a resolved "<n>px" track.
func pxValue(t *testing.T, track string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(strings.TrimSuffix(track, "px"), 64)
	if err != nil || !strings.HasSuffix(track, "px") {
		t.Fatalf("track %q did not resolve to a px length", track)
	}
	return v
}

// TestParseHealthSentinel: the ParseHealth behavior is the LAST definition
// in assets/baud._hs and is installed on <body>; its init marks
// body[data-hs-ok]. A _hyperscript parse error anywhere in the file kills
// every definition after it silently, so the attribute appearing proves
// the whole behaviors file parsed. If this fails, bisect the most recent
// assets/baud._hs edit (watch for reserved words: next, meta, it, result,
// target, …).
func TestParseHealthSentinel(t *testing.T) {
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
		t.Fatalf("body[data-hs-ok] never appeared — assets/baud._hs likely has a parse error "+
			"(ParseHealth, the last behavior in the file, never ran): %v", err)
	}
}

func TestSheetSmoke(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	// Desktop-only hard floor straight from the CSS bundle.
	minWidth, err := page.Evaluate(`() => getComputedStyle(document.body).minWidth`)
	if err != nil {
		t.Fatalf("evaluate minWidth: %v", err)
	}
	if minWidth != "1240px" {
		t.Errorf("body min-width = %v, want 1240px", minWidth)
	}

	// The Panes hyperscript behavior must have applied the grid template
	// (give hyperscript a moment — poll until it is non-none), and the
	// resolved px template must match the data-panes spec "32ch 7px 1fr":
	// right track count, a resolved px value for the 32ch track, 7px gutter.
	tracks := waitForPanesTemplate(t, page)
	spec, err := page.Evaluate(fmt.Sprintf(
		`() => document.querySelector(%q).dataset.panes`, sheetPanes))
	if err != nil {
		t.Fatalf("read data-panes spec: %v", err)
	}
	specTracks := strings.Fields(spec.(string))
	if len(tracks) != len(specTracks) {
		t.Fatalf("resolved template %v has %d tracks, data-panes spec %q has %d",
			tracks, len(tracks), spec, len(specTracks))
	}
	if first := pxValue(t, tracks[0]); first <= 0 {
		t.Errorf("32ch track resolved to %q, want a positive px length", tracks[0])
	}
	if tracks[1] != "7px" {
		t.Errorf("gutter track resolved to %q, want 7px", tracks[1])
	}
	if last := pxValue(t, tracks[2]); last <= 0 {
		t.Errorf("1fr track resolved to %q, want a positive px length", tracks[2])
	}

	// Theme switching is a root-class swap via the tweaks panel: after
	// clicking mocha, body background resolves to t-mocha --bg-app.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.body).backgroundColor === 'rgb(17, 17, 27)'`, nil,
	); err != nil {
		bg, _ := page.Evaluate(`() => getComputedStyle(document.body).backgroundColor`)
		t.Fatalf("t-mocha --bg-app not applied: background stayed %v: %v", bg, err)
	}
}

// TestResizableDragPersists drags the sheet-panes gutter with the mouse,
// asserts the grid template changed accordingly, then reloads and asserts the
// Resizable behavior restored the persisted template from localStorage
// (key baud-panes:sheet-panes).
func TestResizableDragPersists(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}
	before := waitForPanesTemplate(t, page)

	// Panes init done ≠ Resizable's pointerdown listener attached — a press
	// dispatched before attach is lost (events don't queue for future
	// listeners; bit us on slow CI). Wait for the behavior's init signal.
	if err := page.Locator(sheetPanes + ".resizable-ready").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(15000),
	}); err != nil {
		t.Fatalf("Resizable never signalled .resizable-ready: %v", err)
	}

	// Drag the gutter 120px to the right: pointer down on its centre,
	// move in steps (the behavior tracks pointermove), release.
	gutter := page.Locator(sheetPanes + " > .split-gutter")
	box, err := gutter.BoundingBox()
	if err != nil || box == nil {
		t.Fatalf("gutter bounding box: %v", err)
	}
	x, y := box.X+box.Width/2, box.Y+box.Height/2
	const delta = 120
	mouse := page.Mouse()
	if err := mouse.Move(x, y); err != nil {
		t.Fatalf("move to gutter: %v", err)
	}
	if err := mouse.Down(); err != nil {
		t.Fatalf("mouse down: %v", err)
	}
	// The drag handler is interpreted hyperscript: under CI load the whole
	// mouse sequence can finish before its listen loop starts, in which case
	// every move (and the pointerup) is missed and nothing persists. The
	// behavior adds .drag immediately before its listen loop — gate on it,
	// then give the interpreter a beat to enter its first `wait for`.
	if err := page.Locator(sheetPanes + " > .split-gutter.drag").WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("drag handler never signalled .drag after pointerdown: %v", err)
	}
	page.WaitForTimeout(300)
	if err := mouse.Move(x+delta, y, playwright.MouseMoveOptions{Steps: playwright.Int(20)}); err != nil {
		t.Fatalf("drag: %v", err)
	}
	page.WaitForTimeout(300)
	if err := mouse.Up(); err != nil {
		t.Fatalf("mouse up: %v", err)
	}

	// The handler persists asynchronously after pointerup — poll until the
	// template lands under baud-panes:<id>, the proof the drag completed.
	if _, err := page.WaitForFunction(
		`() => !!localStorage.getItem('baud-panes:sheet-panes')`, nil,
	); err != nil {
		t.Fatalf("drag never persisted to localStorage baud-panes:sheet-panes: %v", err)
	}

	// Pointermove events coalesce under CI load, so the final position may
	// lag the mathematical endpoint — the contract is direction + magnitude
	// + conservation, not the exact delta (the persistence round-trip below
	// IS exact).
	after := strings.Fields(computedStyleSel(t, page, sheetPanes, "gridTemplateColumns"))
	if len(after) != len(before) {
		t.Fatalf("template track count changed: before %v, after %v", before, after)
	}
	growth := pxValue(t, after[0]) - pxValue(t, before[0])
	shrink := pxValue(t, before[2]) - pxValue(t, after[2])
	if growth < delta/2 || growth > delta+2 {
		t.Errorf("first track grew %.1fpx after a %dpx drag, want within [%d, %d]", growth, delta, delta/2, delta+2)
	}
	if after[1] != "7px" {
		t.Errorf("gutter track after drag = %q, want 7px", after[1])
	}
	if shrink < delta/2 || shrink > delta+2 {
		t.Errorf("last track shrank %.1fpx after a %dpx drag, want within [%d, %d]", shrink, delta, delta/2, delta+2)
	}
	if diff := growth - shrink; diff < -1 || diff > 1 {
		t.Errorf("drag not conserved: first grew %.1fpx but last shrank %.1fpx", growth, shrink)
	}

	// Capture the persisted template for the restore comparison.
	saved, err := page.Evaluate(`() => localStorage.getItem('baud-panes:sheet-panes')`)
	if err != nil {
		t.Fatalf("read localStorage: %v", err)
	}
	savedTpl, ok := saved.(string)
	if !ok || savedTpl == "" {
		t.Fatalf("localStorage baud-panes:sheet-panes = %v, want the persisted template", saved)
	}

	// Reload: Resizable init must restore the persisted template (Panes
	// installs first, so the persisted value wins over the authored one).
	if _, err := page.Reload(playwright.PageReloadOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("reload: %v", err)
	}
	// Poll until the inline style matches the persisted tracks — proof the
	// restore came from localStorage, not a fresh layout (Panes init sets
	// the authored "32ch 7px 1fr" first; Resizable init then overrides it).
	// Compare numerically: CSSOM rounds "350.4375px" to "350.438px".
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => {
			const el = document.querySelector(%q);
			const saved = localStorage.getItem('baud-panes:sheet-panes');
			if (!el || !saved) return false;
			const got = el.style.gridTemplateColumns.split(/\s+/);
			const want = saved.split(/\s+/);
			return got.length === want.length && got.every((tk, i) =>
				tk.endsWith('px') &&
				Math.abs(parseFloat(tk) - parseFloat(want[i])) < 0.5);
		}`, sheetPanes), nil,
	); err != nil {
		inline, _ := page.Evaluate(fmt.Sprintf(
			`() => document.querySelector(%q).style.gridTemplateColumns`, sheetPanes))
		t.Fatalf("persisted template never restored: inline = %v, persisted %q: %v",
			inline, savedTpl, err)
	}
	restored := strings.Fields(computedStyleSel(t, page, sheetPanes, "gridTemplateColumns"))
	savedTracks := strings.Fields(savedTpl)
	if len(restored) != len(savedTracks) {
		t.Fatalf("restored template %v, want the persisted %q", restored, savedTpl)
	}
	const tol = 0.5 // CSSOM rounding only — the restore itself is exact
	for i := range restored {
		if got, want := pxValue(t, restored[i]), pxValue(t, savedTracks[i]); got < want-tol || got > want+tol {
			t.Errorf("restored track %d = %q, want ~%q (persisted)", i, restored[i], savedTracks[i])
		}
	}
}

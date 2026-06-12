//go:build e2e

package e2e

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Selectors into the sheet's toast fixtures (demo/sheet_toast.templ —
// structure documented load-bearing there): four static Sticky toasts
// (ok, err, warn, info in order) for styling/dismiss assertions, and the
// live trigger buttons whose round-trips OOB-swap into the page-level
// #toasts region (baud/page.templ).
const (
	toastStatic  = "#toast-static-demo"
	toastRegion  = "#toasts"
	toastBtnOK   = "#toast-btn-ok"
	toastBtnWarn = "#toast-btn-warn"
	// toastBtnShort round-trips ?ms=600 — the honest short auto-dismiss.
	toastBtnShort = "#toast-btn-short"

	// Tone tokens (assets/css/tokens.css), resolved per theme.
	toastGruvOK   = "rgb(184, 187, 38)"  // t-gruvbox --ok   #b8bb26
	toastGruvErr  = "rgb(251, 73, 52)"   // t-gruvbox --err  #fb4934
	toastGruvWarn = "rgb(250, 189, 47)"  // t-gruvbox --warn #fabd2f
	toastGruvInfo = "rgb(131, 165, 152)" // t-gruvbox --info #83a598
	toastMochaOK  = "rgb(166, 227, 161)" // t-mocha   --ok   #a6e3a1
	toastMochaErr = "rgb(243, 139, 168)" // t-mocha   --err  #f38ba8
)

// toastOpenSheet opens the component sheet and waits for the ParseHealth
// sentinel — proof the whole behaviors file parsed and hyperscript booted,
// so Toast installs (auto/manual dismiss) are live before interacting.
func toastOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	page := openSheet(t)
	if err := page.Locator("body[data-hs-ok]").WaitFor(); err != nil {
		t.Fatalf("hyperscript never booted (ParseHealth sentinel missing): %v", err)
	}
	return page
}

// toastCount counts the live toasts currently in the OOB region.
func toastCount(t *testing.T, page playwright.Page) int {
	t.Helper()
	n, err := page.Locator(toastRegion + " > .toast").Count()
	if err != nil {
		t.Fatalf("count region toasts: %v", err)
	}
	return n
}

// TestToastOOBSwap: clicking a trigger fires a real htmx round-trip whose
// hx-swap-oob fragment lands the toast inside #toasts; two triggers stack.
func TestToastOOBSwap(t *testing.T) {
	page := toastOpenSheet(t)

	if got := toastCount(t, page); got != 0 {
		t.Fatalf("region not empty before triggering: %d toasts", got)
	}
	if err := page.Locator(toastBtnOK).Click(); err != nil {
		t.Fatalf("click ok trigger: %v", err)
	}
	ok := page.Locator(toastRegion + " > .toast.tone-ok")
	if err := ok.WaitFor(); err != nil {
		t.Fatalf("ok toast never landed in #toasts: %v", err)
	}
	title, err := ok.Locator(".toast-title").TextContent()
	if err != nil {
		t.Fatalf("read toast title: %v", err)
	}
	if title != "Deployed ingest-gw v2.14.1" {
		t.Errorf("OOB toast title = %q, want %q", title, "Deployed ingest-gw v2.14.1")
	}
	body, err := ok.Locator(".toast-body").TextContent()
	if err != nil {
		t.Fatalf("read toast body: %v", err)
	}
	if body != "rollout complete on 12 pods" {
		t.Errorf("OOB toast body = %q, want %q", body, "rollout complete on 12 pods")
	}
	if got, _ := ok.GetAttribute("role"); got != "status" {
		t.Errorf("OOB toast role = %q, want status", got)
	}
	if got, _ := page.Locator(toastRegion).GetAttribute("aria-live"); got != "polite" {
		t.Errorf("#toasts aria-live = %q, want polite", got)
	}

	// Stacking: a second tone coexists with the first.
	if err := page.Locator(toastBtnWarn).Click(); err != nil {
		t.Fatalf("click warn trigger: %v", err)
	}
	if err := page.Locator(toastRegion + " > .toast.tone-warn").WaitFor(); err != nil {
		t.Fatalf("warn toast never landed: %v", err)
	}
	if got := toastCount(t, page); got != 2 {
		t.Errorf("stacked toasts = %d, want 2 coexisting", got)
	}

	// The stack sits fixed bottom-right, above the statusbar:
	// bottom = --rh + 12px = 36px at d-dense.
	if got := computedStyleSel(t, page, toastRegion, "position"); got != "fixed" {
		t.Errorf("#toasts position = %q, want fixed", got)
	}
	if got := computedStyleSel(t, page, toastRegion, "bottom"); got != "36px" {
		t.Errorf("#toasts bottom = %q, want 36px (--rh 24px + 12px)", got)
	}
}

// TestToastAutoDismiss: a 600ms toast removes itself; non-vacuity is the
// coexisting 4000ms toast, which must still be attached the moment the
// short one detaches (interval-gated removal, not a blanket sweep — and
// with the behavior absent the short toast would persist and the detach
// wait below would fail loudly).
func TestToastAutoDismiss(t *testing.T) {
	page := toastOpenSheet(t)

	if err := page.Locator(toastBtnWarn).Click(); err != nil {
		t.Fatalf("click warn (4000ms) trigger: %v", err)
	}
	warn := page.Locator(toastRegion + " > .toast.tone-warn")
	if err := warn.WaitFor(); err != nil {
		t.Fatalf("warn toast never landed: %v", err)
	}
	if err := page.Locator(toastBtnShort).Click(); err != nil {
		t.Fatalf("click short (600ms) trigger: %v", err)
	}
	short := page.Locator(toastRegion + " > .toast.tone-ok")
	if err := short.WaitFor(); err != nil {
		t.Fatalf("short toast never landed: %v", err)
	}
	if got, _ := short.GetAttribute("data-toast-ms"); got != "600" {
		t.Fatalf("short toast data-toast-ms = %q, want 600", got)
	}

	// The toast existed (asserted above) and then removes ITSELF.
	if err := short.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(5000),
	}); err != nil {
		t.Fatalf("600ms toast never auto-dismissed: %v", err)
	}
	visible, err := warn.IsVisible()
	if err != nil {
		t.Fatalf("warn visibility: %v", err)
	}
	if !visible {
		t.Errorf("4000ms toast vanished with the 600ms one — removal is not interval-gated")
	}
}

// TestToastHoverPause: hovering pauses the countdown. Two 600ms toasts;
// the hovered one must outlive the unhovered one (the clock), then
// dismiss after the pointer leaves — no raw sleeps, the unhovered
// toast's removal is the time gate.
func TestToastHoverPause(t *testing.T) {
	page := toastOpenSheet(t)

	for i := 0; i < 2; i++ {
		if err := page.Locator(toastBtnShort).Click(); err != nil {
			t.Fatalf("click short trigger #%d: %v", i, err)
		}
	}
	toasts := page.Locator(toastRegion + " > .toast")
	first := toasts.Nth(0)
	if err := toasts.Nth(1).WaitFor(); err != nil {
		t.Fatalf("second short toast never landed: %v", err)
	}
	if err := first.Hover(); err != nil {
		t.Fatalf("hover first toast: %v", err)
	}
	// The unhovered toast dismisses on schedule…
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#toasts > .toast').length === 1`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
	); err != nil {
		t.Fatalf("unhovered 600ms toast never auto-dismissed: %v", err)
	}
	// …while the hovered one survived its own 600ms interval.
	visible, err := first.IsVisible()
	if err != nil {
		t.Fatalf("hovered toast visibility: %v", err)
	}
	if !visible {
		t.Fatalf("hovered toast dismissed — hover did not pause the countdown")
	}
	// Unhover: the countdown re-arms and the survivor dismisses too.
	if err := page.Mouse().Move(0, 0); err != nil {
		t.Fatalf("move pointer away: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#toasts > .toast').length === 0`, nil,
		playwright.PageWaitForFunctionOptions{Timeout: playwright.Float(5000)},
	); err != nil {
		t.Fatalf("toast never dismissed after unhover: %v", err)
	}
}

// TestToastManualDismiss: ✕ removes exactly its own toast — mouse on one
// static sticky toast, keyboard (Tab-reachable button + Enter) on another.
// Sticky toasts never auto-remove, so any removal here is the ✕ handler;
// with the behavior absent they would persist and the waits fail.
func TestToastManualDismiss(t *testing.T) {
	page := toastOpenSheet(t)

	static := page.Locator(toastStatic + " > .toast")
	if n, err := static.Count(); err != nil || n != 4 {
		t.Fatalf("static sticky toasts = %d (err %v), want 4", n, err)
	}
	errToast := page.Locator(toastStatic + " > .toast.tone-err")
	if got, _ := errToast.GetAttribute("data-toast-ms"); got != "0" {
		t.Fatalf("static err toast data-toast-ms = %q, want 0 (sticky)", got)
	}
	if err := errToast.Locator(".x-btn").Click(); err != nil {
		t.Fatalf("click ✕ on err toast: %v", err)
	}
	if err := errToast.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateDetached,
	}); err != nil {
		t.Fatalf("err toast not removed by ✕: %v", err)
	}
	if n, _ := static.Count(); n != 3 {
		t.Errorf("✕ removed more than its own toast: %d static left, want 3", n)
	}

	// Keyboard: the ✕ is a real focusable <button>; Enter activates it.
	warnX := page.Locator(toastStatic + " > .toast.tone-warn .x-btn")
	if err := warnX.Focus(); err != nil {
		t.Fatalf("focus warn ✕: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if err := page.Locator(toastStatic + " > .toast.tone-warn").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateDetached,
	}); err != nil {
		t.Fatalf("warn toast not removed by keyboard ✕: %v", err)
	}
}

// TestToastToneTokens: per-tone computed styles on the static fixtures —
// the 3px left bar and glyph resolve to the theme tone vars, and a
// root-class swap to t-mocha re-resolves them with no other change.
func TestToastToneTokens(t *testing.T) {
	page := toastOpenSheet(t)

	for _, tc := range []struct{ tone, want string }{
		{"ok", toastGruvOK},
		{"err", toastGruvErr},
		{"warn", toastGruvWarn},
		{"info", toastGruvInfo},
	} {
		sel := fmt.Sprintf("%s > .toast.tone-%s", toastStatic, tc.tone)
		if got := computedStyleSel(t, page, sel, "borderLeftColor"); got != tc.want {
			t.Errorf("gruvbox %s tone bar = %q, want %q", tc.tone, got, tc.want)
		}
		if got := computedStyleSel(t, page, sel+" .toast-glyph", "color"); got != tc.want {
			t.Errorf("gruvbox %s glyph colour = %q, want %q", tc.tone, got, tc.want)
		}
	}
	sel := toastStatic + " > .toast.tone-ok"
	if got := computedStyleSel(t, page, sel, "borderLeftWidth"); got != "3px" {
		t.Errorf("tone bar width = %q, want 3px", got)
	}
	if got := computedStyleSel(t, page, sel, "borderRadius"); got != "0px" {
		t.Errorf("toast border-radius = %q, want 0px", got)
	}

	// Token flow proof: theme switches by root-class swap only.
	if _, err := page.Evaluate(
		`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if got := computedStyleSel(t, page, toastStatic+" > .toast.tone-ok", "borderLeftColor"); got != toastMochaOK {
		t.Errorf("mocha ok tone bar = %q, want %q", got, toastMochaOK)
	}
	if got := computedStyleSel(t, page, toastStatic+" > .toast.tone-err", "borderLeftColor"); got != toastMochaErr {
		t.Errorf("mocha err tone bar = %q, want %q", got, toastMochaErr)
	}
}

// TestToastReducedMotion: the 160ms slide-in runs only under
// prefers-reduced-motion: no-preference.
func TestToastReducedMotion(t *testing.T) {
	page := toastOpenSheet(t)

	sel := toastStatic + " > .toast.tone-ok"
	if got := computedStyleSel(t, page, sel, "animationName"); got != "toast-in" {
		t.Errorf("toast animation-name = %q, want toast-in", got)
	}
	if got := computedStyleSel(t, page, sel, "animationDuration"); got != "0.16s" {
		t.Errorf("toast animation-duration = %q, want 0.16s", got)
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{
		ReducedMotion: playwright.ReducedMotionReduce,
	}); err != nil {
		t.Fatalf("emulate reduced motion: %v", err)
	}
	if got := computedStyleSel(t, page, sel, "animationName"); got != "none" {
		t.Errorf("toast animates under reduced motion: animation-name = %q, want none", got)
	}
}

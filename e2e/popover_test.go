//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved colours asserted below (from assets/css/tokens.css):
//
//	t-gruvbox --bg-panel #282828 → rgb(40, 40, 40), --bg-raised #32302f → rgb(50, 48, 47)
//	          --border-strong #504945 → rgb(80, 73, 69), --fg-faint #7c6f64 → rgb(124, 111, 100)
//	t-mocha   --bg-panel #181825 → rgb(24, 24, 37), --bg-raised #1e1e2e → rgb(30, 30, 46)
const (
	popGruvPanel   = "rgb(40, 40, 40)"
	popGruvStrong  = "rgb(80, 73, 69)"
	popMochaPanel  = "rgb(24, 24, 37)"
	tipGruvRaised  = "rgb(50, 48, 47)"
	tipMochaRaised = "rgb(30, 30, 46)"
	tipGruvFgFaint = "rgb(124, 111, 100)"
)

// popoverOpenSheet opens the sheet and waits for the ParseHealth sentinel
// (body[data-hs-ok]) so every hyperscript behavior — MenuDismiss included —
// is installed before the test interacts.
func popoverOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	page := openSheet(t)
	if err := page.Locator("body[data-hs-ok]").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("body[data-hs-ok] never appeared — hyperscript did not boot: %v", err)
	}
	return page
}

// popoverAttr reads an attribute (fails the test when the element is missing).
func popoverAttr(t *testing.T, page playwright.Page, sel, name string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => { const el = document.querySelector(%q); return el ? (el.getAttribute(%q) ?? "") : "NO ELEMENT"; }`,
		sel, name))
	if err != nil {
		t.Fatalf("attribute %s of %q: %v", name, sel, err)
	}
	s, _ := v.(string)
	if s == "NO ELEMENT" {
		t.Fatalf("no element matches %q", sel)
	}
	return s
}

// tooltipPseudo resolves one computed-style property of a ::before/::after
// pseudo-element — element handles cannot reach pseudo-elements, so this
// goes through getComputedStyle(el, pseudo) directly.
func tooltipPseudo(t *testing.T, page playwright.Page, sel, pseudo, prop string) string {
	t.Helper()
	v, err := page.Evaluate(
		`args => { const el = document.querySelector(args.sel); return el ? getComputedStyle(el, args.pseudo)[args.prop] : "NO ELEMENT"; }`,
		map[string]any{"sel": sel, "pseudo": pseudo, "prop": prop},
	)
	if err != nil {
		t.Fatalf("computed %s of %q%s: %v", prop, sel, pseudo, err)
	}
	s, _ := v.(string)
	if s == "NO ELEMENT" {
		t.Fatalf("no element matches %q", sel)
	}
	return s
}

// tooltipHover drives a real pointer onto the host's center via the raw
// Mouse API. Locator.Hover is unusable on INLINE tooltip hosts: CDP
// getContentQuads includes the (invisible, pointer-events: none) ::after
// tip box for inline elements, so Playwright's hit-target check aims off
// the host and times out — a driver quirk, not a CSS one (the host's own
// client rect hit-tests to itself; buttons hover fine with the same CSS).
// The tooltipWaitShown poll that follows is the actionability gate.
func tooltipHover(t *testing.T, page playwright.Page, sel string) {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(`() => {
		const el = document.querySelector(%q);
		if (!el) return null;
		el.scrollIntoView({block: "center"});
		const r = el.getBoundingClientRect();
		return {x: r.x + r.width / 2, y: r.y + r.height / 2};
	}`, sel))
	if err != nil {
		t.Fatalf("locate %q for hover: %v", sel, err)
	}
	pt, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("no element matches %q (got %v)", sel, v)
	}
	if err := page.Mouse().Move(tooltipNum(t, pt["x"]), tooltipNum(t, pt["y"])); err != nil {
		t.Fatalf("mouse move onto %q: %v", sel, err)
	}
}

// tooltipNum coerces an Evaluate-returned coordinate to float64.
func tooltipNum(t *testing.T, v any) float64 {
	t.Helper()
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	}
	t.Fatalf("non-numeric coordinate %T(%v)", v, v)
	return 0
}

// tooltipWaitShown polls until the host's ::after tip reaches opacity 1 —
// locator-gated (WaitForFunction), never a raw sleep: the 150ms show delay
// rides the CSS transition, so the poll IS the delay assertion partner.
func tooltipWaitShown(t *testing.T, page playwright.Page, sel string) {
	t.Helper()
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => getComputedStyle(document.querySelector(%q), "::after").opacity === "1"`, sel), nil,
	); err != nil {
		t.Fatalf("tooltip ::after on %q never reached opacity 1: %v", sel, err)
	}
}

// TestPopoverOpenDismiss: the trigger click opens the anchored 280px panel
// (--bg-panel fill, strong border, deep shadow); MenuDismiss closes it on
// Esc AND on outside click, with aria-expanded tracking every transition.
// Clicks inside the panel do NOT dismiss it.
func TestPopoverOpenDismiss(t *testing.T) {
	page := popoverOpenSheet(t)
	trigger := "#pop-actions .pop-trigger"
	panel := "#pop-actions .popover"

	if got := computedStyleSel(t, page, panel, "display"); got != "none" {
		t.Fatalf("panel display = %v before opening, want none", got)
	}
	if got := popoverAttr(t, page, trigger, "aria-controls"); got != "pop-actions-panel" {
		t.Errorf("trigger aria-controls = %q, want pop-actions-panel", got)
	}

	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("click trigger: %v", err)
	}
	if got := popoverAttr(t, page, trigger, "aria-expanded"); got != "true" {
		t.Errorf("aria-expanded after open = %q, want true", got)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "flex" {
		t.Errorf("open panel display = %v, want flex", got)
	}
	if got := computedStyleSel(t, page, panel, "width"); got != "280px" {
		t.Errorf("panel width = %v, want 280px (anchored quick-actions panel)", got)
	}
	if got := computedStyleSel(t, page, panel, "position"); got != "absolute" {
		t.Errorf("panel position = %v, want absolute", got)
	}
	if got := computedStyleSel(t, page, panel, "backgroundColor"); got != popGruvPanel {
		t.Errorf("panel background = %v, want %v (--bg-panel)", got, popGruvPanel)
	}
	if got := computedStyleSel(t, page, panel, "borderTopColor"); got != popGruvStrong {
		t.Errorf("panel border = %v, want %v (--border-strong)", got, popGruvStrong)
	}
	if got := computedStyleSel(t, page, panel, "boxShadow"); got == "none" {
		t.Error("panel box-shadow = none, want a --shadow drop shadow")
	}
	if got := computedStyleSel(t, page, panel, "borderTopLeftRadius"); got != "0px" {
		t.Errorf("panel border-radius = %v, want 0px", got)
	}

	// Clicking inside the panel is NOT an outside click — it stays open.
	if err := page.Locator("#pop-actions .popover .menu-item").First().Click(); err != nil {
		t.Fatalf("click inside panel: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "flex" {
		t.Errorf("panel display after inside click = %v, want flex (still open)", got)
	}

	// Esc closes (MenuDismiss) and aria-expanded follows.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "none" {
		t.Errorf("panel display after Esc = %v, want none", got)
	}
	if got := popoverAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after Esc = %q, want false", got)
	}

	// Outside click closes too.
	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("re-open trigger: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "flex" {
		t.Fatalf("panel did not re-open, display = %v", got)
	}
	if err := page.Locator("h2.sheet-h").First().Click(); err != nil {
		t.Fatalf("outside click: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "none" {
		t.Errorf("panel display after outside click = %v, want none", got)
	}
	if got := popoverAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after outside click = %q, want false", got)
	}

	// Trigger click toggles closed again (and back in sync).
	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("open trigger: %v", err)
	}
	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("toggle trigger closed: %v", err)
	}
	if got := computedStyleSel(t, page, panel, "display"); got != "none" {
		t.Errorf("panel display after trigger toggle = %v, want none", got)
	}
	if got := popoverAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after trigger toggle = %q, want false", got)
	}
}

// TestPopoverOneAtATime: opening one popover is an outside click for every
// other open popover — the first closes (with its ARIA following) when the
// second's trigger is clicked.
func TestPopoverOneAtATime(t *testing.T) {
	page := popoverOpenSheet(t)

	if err := page.Locator("#pop-actions .pop-trigger").Click(); err != nil {
		t.Fatalf("open first popover: %v", err)
	}
	if got := computedStyleSel(t, page, "#pop-actions .popover", "display"); got != "flex" {
		t.Fatalf("first popover did not open, display = %v", got)
	}

	if err := page.Locator("#pop-deploy .pop-trigger").Click(); err != nil {
		t.Fatalf("open second popover: %v", err)
	}
	if got := computedStyleSel(t, page, "#pop-deploy .popover", "display"); got != "flex" {
		t.Errorf("second popover display = %v, want flex", got)
	}
	if got := computedStyleSel(t, page, "#pop-actions .popover", "display"); got != "none" {
		t.Errorf("first popover display = %v, want none (one at a time)", got)
	}
	if got := popoverAttr(t, page, "#pop-actions .pop-trigger", "aria-expanded"); got != "false" {
		t.Errorf("first trigger aria-expanded = %q, want false after losing to the second", got)
	}
	if got := popoverAttr(t, page, "#pop-deploy .pop-trigger", "aria-expanded"); got != "true" {
		t.Errorf("second trigger aria-expanded = %q, want true", got)
	}
}

// TestPopoverThemeSwap: the open panel re-resolves --bg-panel under t-mocha
// by root-class swap only.
func TestPopoverThemeSwap(t *testing.T) {
	page := popoverOpenSheet(t)

	if err := page.Locator("#pop-actions .pop-trigger").Click(); err != nil {
		t.Fatalf("open popover: %v", err)
	}
	if got := computedStyleSel(t, page, "#pop-actions .popover", "backgroundColor"); got != popGruvPanel {
		t.Fatalf("gruvbox panel background = %v, want %v", got, popGruvPanel)
	}
	if _, err := page.Evaluate(`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme: %v", err)
	}
	if got := computedStyleSel(t, page, "#pop-actions .popover", "backgroundColor"); got != popMochaPanel {
		t.Errorf("t-mocha panel background = %v, want %v (--bg-panel re-resolved)", got, popMochaPanel)
	}
}

// TestTooltipHover: the pure-CSS tip starts hidden, carries the 150ms show
// delay on its opacity transition, and reaches opacity 1 on hover with the
// data-tip text as its ::after content (--bg-raised fill, strong border)
// and the ::before arrow visible alongside.
func TestTooltipHover(t *testing.T) {
	page := popoverOpenSheet(t)
	host := "#tip-badge"

	if got := tooltipPseudo(t, page, host, "::after", "opacity"); got != "0" {
		t.Fatalf("tip ::after opacity before hover = %v, want 0", got)
	}
	if got := tooltipPseudo(t, page, host, "::before", "opacity"); got != "0" {
		t.Fatalf("tip ::before opacity before hover = %v, want 0", got)
	}
	// The 150ms show delay is the transition delay (motion-gated CSS;
	// headless chromium runs with no-preference, so it must resolve).
	if got := tooltipPseudo(t, page, host, "::after", "transitionDelay"); got != "0.15s" {
		t.Errorf("tip ::after transition-delay = %v, want 0.15s", got)
	}
	if got := tooltipPseudo(t, page, host, "::after", "content"); got != `"works on any element"` {
		t.Errorf("tip ::after content = %v, want the data-tip text", got)
	}

	tooltipHover(t, page, host)
	tooltipWaitShown(t, page, host)
	if got := tooltipPseudo(t, page, host, "::before", "opacity"); got != "1" {
		t.Errorf("tip ::before arrow opacity on hover = %v, want 1", got)
	}
	if got := tooltipPseudo(t, page, host, "::after", "backgroundColor"); got != tipGruvRaised {
		t.Errorf("tip background = %v, want %v (--bg-raised)", got, tipGruvRaised)
	}
	if got := tooltipPseudo(t, page, host, "::after", "borderTopColor"); got != popGruvStrong {
		t.Errorf("tip border = %v, want %v (--border-strong)", got, popGruvStrong)
	}
	if got := tooltipPseudo(t, page, host, "::before", "borderTopColor"); got != popGruvStrong {
		t.Errorf("arrow border-top = %v, want %v (--border-strong)", got, popGruvStrong)
	}
	if got := tooltipPseudo(t, page, host, "::after", "borderTopLeftRadius"); got != "0px" {
		t.Errorf("tip border-radius = %v, want 0px", got)
	}

	// Token flow under t-mocha: the visible tip re-resolves --bg-raised.
	if _, err := page.Evaluate(`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme: %v", err)
	}
	if got := tooltipPseudo(t, page, host, "::after", "backgroundColor"); got != tipMochaRaised {
		t.Errorf("t-mocha tip background = %v, want %v (--bg-raised re-resolved)", got, tipMochaRaised)
	}
}

// TestTooltipMultiline: newlines in data-tip survive into the rendered
// ::after content and white-space: pre keeps them as line breaks — the
// aligned multi-line mono tip contract.
func TestTooltipMultiline(t *testing.T) {
	page := popoverOpenSheet(t)
	host := "#tip-latency"

	if got := tooltipPseudo(t, page, host, "::after", "whiteSpace"); got != "pre" {
		t.Errorf("tip white-space = %v, want pre", got)
	}
	content := tooltipPseudo(t, page, host, "::after", "content")
	for _, line := range []string{"p99 = 312ms", "p95 = 104ms", "p50 =  22ms"} {
		if !strings.Contains(content, line) {
			t.Errorf("tip content %q is missing line %q", content, line)
		}
	}
	// getComputedStyle serializes a newline in a content string as the CSS
	// escape `\a` — its presence proves the line breaks survived.
	if !strings.Contains(content, `\a`) && !strings.Contains(content, "\n") {
		t.Errorf("tip content %q lost its newlines (multi-line tips must stay multi-line)", content)
	}

	tooltipHover(t, page, host)
	tooltipWaitShown(t, page, host)
}

// TestTooltipKeyboardFocus: tips are keyboard-accessible — tabbing onto a
// focusable host (:focus-visible) shows the tip without any pointer.
func TestTooltipKeyboardFocus(t *testing.T) {
	page := popoverOpenSheet(t)

	if err := page.Locator("#tip-key-anchor").Focus(); err != nil {
		t.Fatalf("focus anchor: %v", err)
	}
	if err := page.Keyboard().Press("Tab"); err != nil {
		t.Fatalf("press Tab: %v", err)
	}
	if got := activeID(t, page); got != "tip-key" {
		t.Fatalf("after Tab activeElement = %q, want tip-key", got)
	}
	tooltipWaitShown(t, page, "#tip-key")
	if got := tooltipPseudo(t, page, "#tip-key", "::after", "content"); got != `"shown on focus-visible"` {
		t.Errorf("focus tip content = %v, want the data-tip text", got)
	}
}

// TestTooltipUnderVariant: tip-under marks explained prose — dotted
// --fg-faint underline and a help cursor on the host.
func TestTooltipUnderVariant(t *testing.T) {
	page := popoverOpenSheet(t)

	if got := computedStyleSel(t, page, "#tip-budget", "textDecorationLine"); got != "underline" {
		t.Errorf("tip-under text-decoration-line = %v, want underline", got)
	}
	if got := computedStyleSel(t, page, "#tip-budget", "textDecorationStyle"); got != "dotted" {
		t.Errorf("tip-under text-decoration-style = %v, want dotted", got)
	}
	if got := computedStyleSel(t, page, "#tip-budget", "textDecorationColor"); got != tipGruvFgFaint {
		t.Errorf("tip-under underline colour = %v, want %v (--fg-faint)", got, tipGruvFgFaint)
	}
	if got := computedStyleSel(t, page, "#tip-budget", "cursor"); got != "help" {
		t.Errorf("tip-under cursor = %v, want help", got)
	}
	// the plain Tip host has neither the underline nor the help cursor
	if got := computedStyleSel(t, page, "#tip-badge", "textDecorationLine"); got != "none" {
		t.Errorf("plain tip host text-decoration-line = %v, want none", got)
	}
	if got := computedStyleSel(t, page, "#tip-badge", "cursor"); got == "help" {
		t.Error("plain tip host cursor = help, want the default cursor")
	}
}

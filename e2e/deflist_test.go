//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestDefListBreadcrumbComputedStyles asserts the DefList grid typography
// (faint UPPERCASE keys, tabular-nums values, hairline rules under the
// lines variant only) and the Breadcrumb link/current split, in t-gruvbox
// tokens, then proves the token flow with a t-mocha root-class swap.
func TestDefListBreadcrumbComputedStyles(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	// --- deflist keys: --fg-faint (#7c6f64) + uppercase, --fs-sm ---------
	if got := computedStyleSel(t, page, ".dl > .dl-k", "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("gruvbox dt colour = %q, want --fg-faint rgb(124, 111, 100)", got)
	}
	if got := computedStyleSel(t, page, ".dl > .dl-k", "textTransform"); got != "uppercase" {
		t.Errorf("dt text-transform = %q, want uppercase", got)
	}

	// --- deflist values: tabular numerals ---------------------------------
	if got := computedStyleSel(t, page, ".dl > .dl-v", "fontVariantNumeric"); !strings.Contains(got, "tabular-nums") {
		t.Errorf("dd font-variant-numeric = %q, want tabular-nums", got)
	}

	// --- lines variant draws a hairline; plain deflist draws none ---------
	// Non-vacuous pair: same dd cell, presence/absence flips on .dl-lines.
	if got := computedStyleSel(t, page, ".dl.dl-lines > .dl-v", "borderBottomStyle"); got != "solid" {
		t.Errorf("lined dd border-bottom-style = %q, want solid", got)
	}
	// Hairline is the 55% --border (#3c3836) color-mix: --border channels
	// at partial alpha — neither transparent nor the solid border colour.
	assertToneAlpha(t, computedStyleSel(t, page, ".dl.dl-lines > .dl-v", "borderBottomColor"),
		"60, 56, 54", "lined dd border-bottom-color")
	if got := computedStyleSel(t, page, ".dl:not(.dl-lines) > .dl-v", "borderBottomStyle"); got != "none" {
		t.Errorf("plain dd border-bottom-style = %q, want none", got)
	}

	// --- breadcrumb: non-current crumbs are faint links -------------------
	if got := computedStyleSel(t, page, ".crumbs a.crumb", "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("gruvbox crumb link colour = %q, want --fg-faint rgb(124, 111, 100)", got)
	}
	if got := computedStyleSel(t, page, ".crumbs a.crumb", "textDecorationLine"); got != "none" {
		t.Errorf("crumb link text-decoration = %q, want none at rest", got)
	}

	// --- current crumb: bold full-foreground text, not a link -------------
	if got := computedStyleSel(t, page, ".crumbs .crumb.cur", "fontWeight"); got != "600" {
		t.Errorf("current crumb font-weight = %q, want 600", got)
	}
	if got := computedStyleSel(t, page, ".crumbs .crumb.cur", "color"); got != "rgb(235, 219, 178)" {
		t.Errorf("gruvbox current crumb colour = %q, want --fg rgb(235, 219, 178)", got)
	}
	tag, err := page.Evaluate(`() => document.querySelector(".crumbs .crumb.cur").tagName`)
	if err != nil {
		t.Fatalf("current crumb tagName: %v", err)
	}
	if tag != "SPAN" {
		t.Errorf("current crumb tag = %v, want SPAN (never a link)", tag)
	}

	// --- separators: presentational glyphs in --fg-faint -------------------
	if got := computedStyleSel(t, page, `.crumbs .crumb-sep[aria-hidden="true"]`, "color"); got != "rgb(124, 111, 100)" {
		t.Errorf("gruvbox separator colour = %q, want --fg-faint rgb(124, 111, 100)", got)
	}

	// --- token flow proof: root-class swap to t-mocha ----------------------
	if _, err := page.Evaluate(
		`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if got := computedStyleSel(t, page, ".dl > .dl-k", "color"); got != "rgb(108, 112, 134)" {
		t.Errorf("mocha dt colour = %q, want --fg-faint rgb(108, 112, 134)", got)
	}
	if got := computedStyleSel(t, page, ".crumbs a.crumb", "color"); got != "rgb(108, 112, 134)" {
		t.Errorf("mocha crumb link colour = %q, want --fg-faint rgb(108, 112, 134)", got)
	}
	if got := computedStyleSel(t, page, ".crumbs .crumb.cur", "color"); got != "rgb(205, 214, 244)" {
		t.Errorf("mocha current crumb colour = %q, want --fg rgb(205, 214, 244)", got)
	}
	assertToneAlpha(t, computedStyleSel(t, page, ".dl.dl-lines > .dl-v", "borderBottomColor"),
		"49, 50, 68", "mocha lined dd border-bottom-color") // --border #313244
}

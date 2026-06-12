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
//	t-gruvbox --bg-panel #282828 → rgb(40, 40, 40), --sel rgba(250,189,47,.14)
//	t-mocha   --bg-panel #181825 → rgb(24, 24, 37), --sel rgba(137,180,250,.13)
const (
	gruvPanel = "rgb(40, 40, 40)"
	gruvSel   = "rgba(250, 189, 47, 0.14)"
	mochaSel  = "rgba(137, 180, 250, 0.13)"
)

// selectOpenSheet opens the sheet and waits for hyperscript to have booted
// (the Panes behavior applying its grid template is the readiness proxy)
// so the select behaviors are installed before the test interacts.
func selectOpenSheet(t *testing.T) playwright.Page {
	t.Helper()
	page := openSheet(t)
	if _, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector('[data-panes]');
			return el && getComputedStyle(el).gridTemplateColumns !== 'none';
		}`, nil,
	); err != nil {
		t.Fatalf("hyperscript never booted (Panes template not applied): %v", err)
	}
	return page
}

// selectAttr reads an attribute (empty string when absent).
func selectAttr(t *testing.T, page playwright.Page, sel, name string) string {
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

// selectChevGlyph reads the ::before content of the chevron inside sel.
func selectChevGlyph(t *testing.T, page playwright.Page, sel string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => getComputedStyle(document.querySelector(%q + " .select-chev"), "::before").content`, sel))
	if err != nil {
		t.Fatalf("chevron content of %q: %v", sel, err)
	}
	s, _ := v.(string)
	return strings.Trim(s, `"`)
}

// TestSelectNativeBaseline: the real <select> is styled to match the
// trigger (token-resolved fill/border, --rh height, appearance none) and
// keeps native behaviour: value, option picking, disabled.
func TestSelectNativeBaseline(t *testing.T) {
	page := selectOpenSheet(t)

	if got := computedStyleSel(t, page, "#sel-native", "height"); got != "24px" {
		t.Errorf("native select height = %v, want 24px (--rh at d-dense)", got)
	}
	if got := computedStyleSel(t, page, "#sel-native", "backgroundColor"); got != "rgb(29, 32, 33)" {
		t.Errorf("native select background = %v, want rgb(29, 32, 33) (--bg-input)", got)
	}
	if got := computedStyleSel(t, page, "#sel-native", "borderTopColor"); got != "rgb(80, 73, 69)" {
		t.Errorf("native select border = %v, want rgb(80, 73, 69) (--border-strong)", got)
	}
	if got := computedStyleSel(t, page, "#sel-native", "appearance"); got != "none" {
		t.Errorf("native select appearance = %v, want none (custom chevron overlays)", got)
	}
	if got := computedStyleSel(t, page, "#sel-native", "borderTopLeftRadius"); got != "0px" {
		t.Errorf("native select border-radius = %v, want 0px", got)
	}
	if got := selectChevGlyph(t, page, "span.select:has(#sel-native)"); got != "▼" {
		t.Errorf("native select chevron = %q, want ▼", got)
	}

	v, err := page.Locator("#sel-native").InputValue()
	if err != nil || v != "us-east" {
		t.Errorf("native select value = %q (err=%v), want us-east", v, err)
	}
	if _, err := page.Locator("#sel-native").SelectOption(playwright.SelectOptionValues{
		Values: &[]string{"ap-south"},
	}); err != nil {
		t.Fatalf("select ap-south: %v", err)
	}
	if v, _ := page.Locator("#sel-native").InputValue(); v != "ap-south" {
		t.Errorf("after native pick value = %q, want ap-south", v)
	}

	disabled, err := page.Locator("#sel-native-disabled").IsDisabled()
	if err != nil || !disabled {
		t.Errorf("disabled native select IsDisabled = %v (err=%v), want true", disabled, err)
	}
	if got := computedStyleSel(t, page, "#sel-native-disabled", "opacity"); got != "0.42" {
		t.Errorf("disabled native select opacity = %v, want 0.42", got)
	}
}

// TestSelectMenuOpenDismiss: trigger click opens the absolutely-positioned
// menu (--bg-panel fill, strong border, shadow, flipped ▲ chevron) and the
// MenuDismiss behavior closes it on Esc AND on outside click, with
// aria-expanded tracking every transition.
func TestSelectMenuOpenDismiss(t *testing.T) {
	page := selectOpenSheet(t)
	trigger := "#sel-env .select-trigger"
	menu := "#sel-env .menu"

	if got := computedStyleSel(t, page, menu, "display"); got != "none" {
		t.Fatalf("menu display = %v before opening, want none", got)
	}
	if got := selectChevGlyph(t, page, "#sel-env"); got != "▼" {
		t.Errorf("closed chevron = %q, want ▼", got)
	}

	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("click trigger: %v", err)
	}
	if got := selectAttr(t, page, trigger, "aria-expanded"); got != "true" {
		t.Errorf("aria-expanded after open = %q, want true", got)
	}
	if got := computedStyleSel(t, page, menu, "display"); got != "block" {
		t.Errorf("open menu display = %v, want block", got)
	}
	if got := computedStyleSel(t, page, menu, "backgroundColor"); got != gruvPanel {
		t.Errorf("menu background = %v, want %v (--bg-panel)", got, gruvPanel)
	}
	if got := computedStyleSel(t, page, menu, "boxShadow"); got == "none" {
		t.Error("menu box-shadow = none, want a --shadow drop shadow")
	}
	if got := computedStyleSel(t, page, menu, "position"); got != "absolute" {
		t.Errorf("menu position = %v, want absolute", got)
	}
	if got := computedStyleSel(t, page, menu, "borderTopColor"); got != "rgb(80, 73, 69)" {
		t.Errorf("menu border = %v, want rgb(80, 73, 69) (--border-strong)", got)
	}
	if got := selectChevGlyph(t, page, "#sel-env"); got != "▲" {
		t.Errorf("open chevron = %q, want ▲", got)
	}

	// Esc closes (MenuDismiss) and the ARIA state follows (SelectKeys).
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if got := computedStyleSel(t, page, menu, "display"); got != "none" {
		t.Errorf("menu display after Esc = %v, want none", got)
	}
	if got := selectAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after Esc = %q, want false", got)
	}

	// Outside click closes too.
	if err := page.Locator(trigger).Click(); err != nil {
		t.Fatalf("re-open trigger: %v", err)
	}
	if got := computedStyleSel(t, page, menu, "display"); got != "block" {
		t.Fatalf("menu did not re-open, display = %v", got)
	}
	if err := page.Locator("h2.sheet-h").First().Click(); err != nil {
		t.Fatalf("outside click: %v", err)
	}
	if got := computedStyleSel(t, page, menu, "display"); got != "none" {
		t.Errorf("menu display after outside click = %v, want none", got)
	}
	if got := selectAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after outside click = %q, want false", got)
	}
}

// TestSelectMenuKeyboard: ↑↓ move the highlighted option (--sel fill +
// aria-activedescendant on the trigger), ↵ picks it — trigger label,
// hidden form input and aria-selected all follow — and the menu closes.
func TestSelectMenuKeyboard(t *testing.T) {
	page := selectOpenSheet(t)
	trigger := "#sel-env .select-trigger"

	if err := page.Locator(trigger).Focus(); err != nil {
		t.Fatalf("focus trigger: %v", err)
	}
	if err := page.Keyboard().Press("ArrowDown"); err != nil {
		t.Fatalf("press ArrowDown: %v", err)
	}
	if got := computedStyleSel(t, page, "#sel-env .menu", "display"); got != "block" {
		t.Fatalf("ArrowDown did not open the menu, display = %v", got)
	}
	if got := selectAttr(t, page, trigger, "aria-activedescendant"); got != "sel-env-opt-0" {
		t.Errorf("aria-activedescendant = %q, want sel-env-opt-0", got)
	}
	if got := computedStyleSel(t, page, "#sel-env-opt-0", "backgroundColor"); got != gruvSel {
		t.Errorf("highlighted option background = %v, want %v (--sel)", got, gruvSel)
	}

	// ↓ to staging, ↓ to prod, ↑ back to staging.
	for _, key := range []string{"ArrowDown", "ArrowDown", "ArrowUp"} {
		if err := page.Keyboard().Press(key); err != nil {
			t.Fatalf("press %s: %v", key, err)
		}
	}
	if got := selectAttr(t, page, trigger, "aria-activedescendant"); got != "sel-env-opt-1" {
		t.Errorf("after ↓↓↑ aria-activedescendant = %q, want sel-env-opt-1", got)
	}
	if got := computedStyleSel(t, page, "#sel-env-opt-0", "backgroundColor"); got == gruvSel {
		t.Error("option 0 kept the --sel highlight after moving away")
	}

	// ↵ picks staging: label, hidden input, aria-selected, closed menu.
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if got := computedStyleSel(t, page, "#sel-env .menu", "display"); got != "none" {
		t.Errorf("menu display after Enter = %v, want none", got)
	}
	if got := selectAttr(t, page, trigger, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after Enter = %q, want false", got)
	}
	label, err := page.Locator("#sel-env .select-value").TextContent()
	if err != nil || strings.TrimSpace(label) != "staging" {
		t.Errorf("trigger label = %q (err=%v), want staging", label, err)
	}
	hidden, err := page.Locator(`#sel-env input[type="hidden"]`).InputValue()
	if err != nil || hidden != "staging" {
		t.Errorf("hidden input value = %q (err=%v), want staging", hidden, err)
	}
	if got := selectAttr(t, page, "#sel-env-opt-1", "aria-selected"); got != "true" {
		t.Errorf("picked option aria-selected = %q, want true", got)
	}
	if got := selectAttr(t, page, "#sel-env-opt-2", "aria-selected"); got != "false" {
		t.Errorf("previous option aria-selected = %q, want false", got)
	}
	// the ✓ check followed the pick (display-gated on .is-active; flex-item
	// blockification computes the inline rule as block — none means hidden)
	if got := computedStyleSel(t, page, "#sel-env-opt-1 .menu-check", "display"); got == "none" {
		t.Error("picked option check display = none, want visible")
	}
	if got := computedStyleSel(t, page, "#sel-env-opt-2 .menu-check", "display"); got != "none" {
		t.Errorf("unpicked option check display = %v, want none", got)
	}
}

// TestComboboxClientFilter: typing filters the option list locally
// (ComboboxFilter), the matched substring is wrapped accent-bold, meta
// matches keep options visible without a label highlight, no-match shows
// the empty row, and ↑↓ ↵ pick from the filtered set.
func TestComboboxClientFilter(t *testing.T) {
	page := selectOpenSheet(t)
	input := "#cb-client .input"
	visible := "#cb-client .menu-item:not([hidden])"

	if err := page.Locator(input).Click(); err != nil {
		t.Fatalf("click combobox input: %v", err)
	}
	if got := computedStyleSel(t, page, "#cb-client .menu", "display"); got != "block" {
		t.Fatalf("focus did not open the menu, display = %v", got)
	}
	if got := selectAttr(t, page, input, "aria-expanded"); got != "true" {
		t.Errorf("aria-expanded after focus = %q, want true", got)
	}

	if err := page.Locator(input).Fill("gw"); err != nil {
		t.Fatalf("type gw: %v", err)
	}
	if n, err := page.Locator(visible).Count(); err != nil || n != 2 {
		t.Errorf("visible options for 'gw' = %d (err=%v), want 2", n, err)
	}
	if n, err := page.Locator("#cb-client .menu-item[hidden]").Count(); err != nil || n != 5 {
		t.Errorf("hidden options for 'gw' = %d (err=%v), want 5", n, err)
	}
	hit, err := page.Locator("#cb-client .cb-match").First().TextContent()
	if err != nil || hit != "gw" {
		t.Errorf("highlighted substring = %q (err=%v), want gw", hit, err)
	}
	if got := computedStyle(t, page.Locator("#cb-client .cb-match").First(), "color"); got != gruvAccent {
		t.Errorf("highlight color = %v, want %v (--accent)", got, gruvAccent)
	}
	if got := computedStyle(t, page.Locator("#cb-client .cb-match").First(), "fontWeight"); got != "700" {
		t.Errorf("highlight font-weight = %v, want 700", got)
	}

	// meta matches keep the option visible but highlight nothing.
	if err := page.Locator(input).Fill("ap-south"); err != nil {
		t.Fatalf("type ap-south: %v", err)
	}
	if n, _ := page.Locator(visible).Count(); n != 2 {
		t.Errorf("visible options for meta 'ap-south' = %d, want 2 (batch-runner, log-tail)", n)
	}
	if n, _ := page.Locator("#cb-client .cb-match").Count(); n != 0 {
		t.Errorf("meta match produced %d label highlights, want 0", n)
	}

	// no match at all → the empty row shows.
	if err := page.Locator(input).Fill("zzz"); err != nil {
		t.Fatalf("type zzz: %v", err)
	}
	if n, _ := page.Locator(visible).Count(); n != 0 {
		t.Errorf("visible options for 'zzz' = %d, want 0", n)
	}
	if got := computedStyleSel(t, page, "#cb-client .menu-empty", "display"); got != "flex" {
		t.Errorf("empty row display = %v, want flex", got)
	}

	// ↑↓ ↵ pick from the filtered set.
	if err := page.Locator(input).Fill("api"); err != nil {
		t.Fatalf("type api: %v", err)
	}
	for _, key := range []string{"ArrowDown", "ArrowDown"} {
		if err := page.Keyboard().Press(key); err != nil {
			t.Fatalf("press %s: %v", key, err)
		}
	}
	if got := selectAttr(t, page, input, "aria-activedescendant"); got != "cb-client-opt-3" {
		t.Errorf("aria-activedescendant = %q, want cb-client-opt-3 (api-edge)", got)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if v, _ := page.Locator(input).InputValue(); v != "api-edge" {
		t.Errorf("input value after pick = %q, want api-edge", v)
	}
	if got := computedStyleSel(t, page, "#cb-client .menu", "display"); got != "none" {
		t.Errorf("menu display after pick = %v, want none", got)
	}
	if got := selectAttr(t, page, input, "aria-expanded"); got != "false" {
		t.Errorf("aria-expanded after pick = %q, want false", got)
	}
	if got := selectAttr(t, page, "#cb-client-opt-3", "aria-selected"); got != "true" {
		t.Errorf("picked option aria-selected = %q, want true", got)
	}

	// Esc closes a re-opened menu (MenuDismiss).
	if err := page.Locator(input).Click(); err != nil {
		t.Fatalf("re-focus input: %v", err)
	}
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	if got := computedStyleSel(t, page, "#cb-client .menu", "display"); got != "none" {
		t.Errorf("menu display after Esc = %v, want none", got)
	}
}

// TestComboboxServerRoundTrip: in server mode typing fires the debounced
// hx-get and the swapped-in fragment replaces the option list, filtered
// and highlighted server-side.
func TestComboboxServerRoundTrip(t *testing.T) {
	page := selectOpenSheet(t)
	input := "#cb-server .input"

	if n, _ := page.Locator("#cb-server .menu-item").Count(); n != 7 {
		t.Fatalf("seeded option count = %d, want 7", n)
	}
	if err := page.Locator(input).Click(); err != nil {
		t.Fatalf("focus server combobox: %v", err)
	}
	if err := page.Locator(input).Fill("api"); err != nil {
		t.Fatalf("type api: %v", err)
	}
	// debounce 200ms + round-trip, then the swap replaces the option list
	if _, err := page.WaitForFunction(
		`() => document.querySelectorAll('#cb-server .menu-item').length === 2`, nil,
	); err != nil {
		n, _ := page.Locator("#cb-server .menu-item").Count()
		t.Fatalf("hx-get swap never landed (still %d options): %v", n, err)
	}
	if got := selectAttr(t, page, "#cb-server-opt-0", "data-value"); got != "api-core" {
		t.Errorf("first swapped option = %q, want api-core", got)
	}
	hit, err := page.Locator("#cb-server .cb-match").First().TextContent()
	if err != nil || hit != "api" {
		t.Errorf("server-rendered highlight = %q (err=%v), want api", hit, err)
	}
	if n, _ := page.Locator("#cb-server .cb-match").Count(); n != 2 {
		t.Errorf("server-rendered highlight count = %d, want 2", n)
	}

	// keyboard pick works on the swapped-in fragment.
	if err := page.Keyboard().Press("ArrowDown"); err != nil {
		t.Fatalf("press ArrowDown: %v", err)
	}
	if got := selectAttr(t, page, input, "aria-activedescendant"); got != "cb-server-opt-0" {
		t.Errorf("aria-activedescendant = %q, want cb-server-opt-0", got)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	if v, _ := page.Locator(input).InputValue(); v != "api-core" {
		t.Errorf("input value after server-mode pick = %q, want api-core", v)
	}
}

// TestSelectMochaTokenFlow: a t-mocha root-class swap re-resolves every
// select token — menu panel fill, --sel keyboard highlight and the accent
// filter highlight — with no other DOM change.
func TestSelectMochaTokenFlow(t *testing.T) {
	page := selectOpenSheet(t)

	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.body).backgroundColor === 'rgb(17, 17, 27)'`, nil,
	); err != nil {
		t.Fatalf("t-mocha never applied: %v", err)
	}

	if err := page.Locator("#sel-env .select-trigger").Click(); err != nil {
		t.Fatalf("open menu: %v", err)
	}
	if got := computedStyleSel(t, page, "#sel-env .menu", "backgroundColor"); got != "rgb(24, 24, 37)" {
		t.Errorf("mocha menu background = %v, want rgb(24, 24, 37) (--bg-panel)", got)
	}
	if err := page.Keyboard().Press("ArrowDown"); err != nil {
		t.Fatalf("press ArrowDown: %v", err)
	}
	if got := computedStyleSel(t, page, "#sel-env .menu-item.hl-row", "backgroundColor"); got != mochaSel {
		t.Errorf("mocha highlight background = %v, want %v (--sel)", got, mochaSel)
	}
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}

	if err := page.Locator("#cb-client .input").Fill("gw"); err != nil {
		t.Fatalf("type gw: %v", err)
	}
	if got := computedStyle(t, page.Locator("#cb-client .cb-match").First(), "color"); got != mochaAccent {
		t.Errorf("mocha filter highlight = %v, want %v (--accent)", got, mochaAccent)
	}
}

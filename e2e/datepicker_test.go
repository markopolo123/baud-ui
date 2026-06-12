//go:build e2e

package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// datepickerWaitMenu polls until the #dp-demo menu's computed display
// matches want ("block" / "none") — open/close runs through hyperscript,
// so a direct read can race the event handler.
func datepickerWaitMenu(t *testing.T, page playwright.Page, want string) {
	t.Helper()
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => { const m = document.querySelector("#dp-demo .dp-menu"); return m && getComputedStyle(m).display === %q; }`,
		want), nil); err != nil {
		got := computedStyleSel(t, page, "#dp-demo .dp-menu", "display")
		t.Fatalf("menu display never became %q (now %q): %v", want, got, err)
	}
}

// datepickerWaitText polls until the element's textContent equals want.
func datepickerWaitText(t *testing.T, page playwright.Page, selector, want, what string) {
	t.Helper()
	if _, err := page.WaitForFunction(fmt.Sprintf(
		`() => document.querySelector(%q)?.textContent === %q`, selector, want), nil); err != nil {
		got, _ := page.Locator(selector).First().TextContent()
		t.Fatalf("%s: text never became %q (now %q): %v", what, want, got, err)
	}
}

func TestDatePickerInteractions(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	dp := page.Locator("#dp-demo")
	trigger := dp.Locator(".dp-trigger")
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	// Closed by default; trigger shows the pinned YYYY-MM-DD selection.
	if got := computedStyleSel(t, page, "#dp-demo .dp-menu", "display"); got != "none" {
		t.Fatalf("menu display = %q, want none before open", got)
	}
	datepickerWaitText(t, page, "#dp-demo .dp-value", monthStart.Format("2006-01-02"), "trigger value")
	// The empty sibling picker prompts instead.
	datepickerWaitText(t, page, "#dp-empty .dp-value", "pick date", "empty trigger value")

	// Keyboard open: the trigger is a real button — Enter activates it.
	if err := trigger.Focus(); err != nil {
		t.Fatalf("focus trigger: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter: %v", err)
	}
	datepickerWaitMenu(t, page, "block")
	if _, err := page.WaitForFunction(
		`() => document.querySelector("#dp-demo .dp-trigger").getAttribute("aria-expanded") === "true"`, nil); err != nil {
		t.Errorf("aria-expanded never became true after open: %v", err)
	}

	// Today outlined: inset hairline in --border-strong (t-gruvbox #504945).
	todayShadow := computedStyle(t, dp.Locator(".dp-day.today"), "boxShadow")
	if !strings.Contains(todayShadow, "inset") || !strings.Contains(todayShadow, "rgb(80, 73, 69)") {
		t.Errorf("today cell box-shadow = %q, want inset rgb(80, 73, 69)", todayShadow)
	}

	// Selected = accent fill with on-accent text (t-gruvbox #fabd2f / #1d2021).
	sel := dp.Locator(".dp-day.sel")
	if got := computedStyle(t, sel, "backgroundColor"); got != "rgb(250, 189, 47)" {
		t.Errorf("selected cell background = %q, want rgb(250, 189, 47)", got)
	}
	if got := computedStyle(t, sel, "color"); got != "rgb(29, 32, 33)" {
		t.Errorf("selected cell color = %q, want rgb(29, 32, 33)", got)
	}

	// Out-month cells are dimmed.
	if got := computedStyle(t, dp.Locator(".dp-day.out").First(), "opacity"); got != "0.45" {
		t.Errorf("out-month cell opacity = %q, want 0.45", got)
	}

	// Month nav is an htmx round-trip: › swaps in the server-rendered next
	// month (no client date math — the title and grid come from the server).
	if err := dp.Locator(`.dp-nav[aria-label="next month"]`).Click(); err != nil {
		t.Fatalf("click next month: %v", err)
	}
	nextMonth := monthStart.AddDate(0, 1, 0)
	datepickerWaitText(t, page, "#dp-demo .dp-title", nextMonth.Format("Jan 2006"), "title after › nav")
	if n, err := dp.Locator(".dp-day").Count(); err != nil || n != 42 {
		t.Errorf("day cells after swap = %d (err %v), want 42", n, err)
	}
	// The selection (first of the previous month) is not in this grid.
	if n, _ := dp.Locator(".dp-day.sel").Count(); n != 0 {
		t.Errorf("selected cells in next-month grid = %d, want 0", n)
	}

	// Preset -7d: server-computed data-date lands in the hidden input and
	// the trigger text, and the menu closes.
	want7 := now.AddDate(0, 0, -7).Format("2006-01-02")
	if err := dp.Locator(`.dp-preset[data-date="` + want7 + `"]`).Click(); err != nil {
		t.Fatalf("click -7d preset: %v", err)
	}
	datepickerWaitText(t, page, "#dp-demo .dp-value", want7, "trigger value after -7d preset")
	if v, err := page.Evaluate(`() => document.querySelector("#dp-demo .dp-input").value`); err != nil || v != want7 {
		t.Errorf("hidden input value = %v (err %v), want %q", v, err, want7)
	}
	datepickerWaitMenu(t, page, "none")
	if _, err := page.WaitForFunction(
		`() => document.querySelector("#dp-demo .dp-trigger").getAttribute("aria-expanded") === "false"`, nil); err != nil {
		t.Errorf("aria-expanded never became false after pick: %v", err)
	}

	// Esc closes.
	if err := trigger.Click(); err != nil {
		t.Fatalf("reopen trigger: %v", err)
	}
	datepickerWaitMenu(t, page, "block")
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatalf("press Escape: %v", err)
	}
	datepickerWaitMenu(t, page, "none")

	// Outside click closes.
	if err := trigger.Click(); err != nil {
		t.Fatalf("reopen trigger: %v", err)
	}
	datepickerWaitMenu(t, page, "block")
	if err := page.Locator("h2.sheet-h").First().Click(); err != nil {
		t.Fatalf("click outside: %v", err)
	}
	datepickerWaitMenu(t, page, "none")

	// Keyboard pick: day cells are real buttons — Enter on a focused
	// in-month cell selects it, moves the accent fill and closes the menu.
	if err := trigger.Click(); err != nil {
		t.Fatalf("reopen trigger: %v", err)
	}
	datepickerWaitMenu(t, page, "block")
	cell := dp.Locator(".dp-day:not(.out)").First()
	cellDate, err := cell.GetAttribute("data-date")
	if err != nil || cellDate == "" {
		t.Fatalf("first in-month cell data-date: %q, %v", cellDate, err)
	}
	if err := cell.Focus(); err != nil {
		t.Fatalf("focus day cell: %v", err)
	}
	if err := page.Keyboard().Press("Enter"); err != nil {
		t.Fatalf("press Enter on day cell: %v", err)
	}
	datepickerWaitText(t, page, "#dp-demo .dp-value", cellDate, "trigger value after keyboard pick")
	datepickerWaitMenu(t, page, "none")

	// Token flow: t-mocha root-class swap re-resolves the accent fill on
	// the picked cell (t-mocha --accent #89b4fa / --on-accent #11111b).
	if _, err := page.Evaluate(`() => document.body.classList.replace("t-gruvbox", "t-mocha")`); err != nil {
		t.Fatalf("swap theme class: %v", err)
	}
	if err := trigger.Click(); err != nil {
		t.Fatalf("reopen trigger: %v", err)
	}
	datepickerWaitMenu(t, page, "block")
	sel = dp.Locator(".dp-day.sel")
	if got := computedStyle(t, sel, "backgroundColor"); got != "rgb(137, 180, 250)" {
		t.Errorf("mocha selected cell background = %q, want rgb(137, 180, 250)", got)
	}
	if got := computedStyle(t, sel, "color"); got != "rgb(17, 17, 27)" {
		t.Errorf("mocha selected cell color = %q, want rgb(17, 17, 27)", got)
	}
}

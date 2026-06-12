//go:build e2e

package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

func assertStyle(t *testing.T, l playwright.Locator, prop, want, what string) {
	t.Helper()
	if got := computedStyle(t, l, prop); got != want {
		t.Errorf("%s: %s = %q, want %q", what, prop, got, want)
	}
}

// TestBtnComputedStyles asserts the Btn/BtnGroup/Kbd visuals per variant in
// the default t-gruvbox + d-dense modes, keyboard focus-visible, and that a
// theme root-class swap re-resolves the variant tokens (t-mocha).
func TestBtnComputedStyles(t *testing.T) {
	srv := startDemo(t)
	page := startBrowser(t)

	if _, err := page.Goto(srv.URL+"/sheet", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		t.Fatalf("goto /sheet: %v", err)
	}

	base := page.Locator("#btn-variants .btn").First()
	primary := page.Locator("#btn-variants .btn-primary").First()
	danger := page.Locator("#btn-variants .btn-danger").First()
	ghost := page.Locator("#btn-variants .btn-ghost").First()
	active := page.Locator("#btn-variants .btn.is-active").First()
	disabled := page.Locator("#btn-variants .btn:disabled").First()

	// Control height is --rh: 24px at d-dense; default fill is --bg-raised.
	assertStyle(t, base, "height", "24px", "default btn")
	assertStyle(t, base, "backgroundColor", "rgb(50, 48, 47)", "default btn") // --bg-raised #32302f
	assertStyle(t, base, "textTransform", "uppercase", "default btn")
	assertStyle(t, base, "fontSize", "10.5px", "default btn") // --fs-sm at d-dense
	assertStyle(t, base, "borderTopLeftRadius", "0px", "default btn")

	// Primary: accent fill, on-accent text (t-gruvbox #fabd2f / #1d2021).
	assertStyle(t, primary, "backgroundColor", "rgb(250, 189, 47)", "primary btn")
	assertStyle(t, primary, "color", "rgb(29, 32, 33)", "primary btn")

	// Danger at rest: transparent fill, err text (--err #fb4934).
	assertStyle(t, danger, "backgroundColor", "rgba(0, 0, 0, 0)", "danger btn")
	assertStyle(t, danger, "color", "rgb(251, 73, 52)", "danger btn")

	// Danger on hover: fills err, border resolves to solid --err.
	if err := danger.Hover(); err != nil {
		t.Fatalf("hover danger: %v", err)
	}
	assertStyle(t, danger, "backgroundColor", "rgb(251, 73, 52)", "danger btn (hover)")
	assertStyle(t, danger, "borderTopColor", "rgb(251, 73, 52)", "danger btn (hover)")
	if err := page.Mouse().Move(0, 0); err != nil {
		t.Fatalf("park mouse: %v", err)
	}

	// Ghost: borderless and muted; active segment: accent fill.
	assertStyle(t, ghost, "borderTopColor", "rgba(0, 0, 0, 0)", "ghost btn")
	assertStyle(t, ghost, "color", "rgb(168, 153, 132)", "ghost btn") // --fg-muted #a89984
	assertStyle(t, active, "backgroundColor", "rgb(250, 189, 47)", "active btn")

	// Disabled: dimmed, not-allowed cursor.
	assertStyle(t, disabled, "opacity", "0.42", "disabled btn")
	assertStyle(t, disabled, "cursor", "not-allowed", "disabled btn")

	// BtnGroup fuses hairlines: every button after the first pulls -1px.
	assertStyle(t, page.Locator(".btn-group .btn").Nth(0), "marginLeft", "0px", "group btn[0]")
	assertStyle(t, page.Locator(".btn-group .btn").Nth(1), "marginLeft", "-1px", "group btn[1]")

	// Kbd chip: --bg-raised fill, --fs-sm, bordered with --border-strong.
	kbd := page.Locator("#kbd-chips .kbd").First()
	assertStyle(t, kbd, "backgroundColor", "rgb(50, 48, 47)", "kbd chip")
	assertStyle(t, kbd, "fontSize", "10.5px", "kbd chip")
	assertStyle(t, kbd, "borderTopColor", "rgb(80, 73, 69)", "kbd chip") // --border-strong #504945

	// Keyboard: Tab until a .btn holds focus — focus-visible must paint the
	// 1px accent outline.
	focused := false
	for i := 0; i < 80; i++ {
		if err := page.Keyboard().Press("Tab"); err != nil {
			t.Fatalf("press Tab: %v", err)
		}
		v, err := page.Evaluate(`() => !!(document.activeElement && document.activeElement.classList.contains('btn'))`)
		if err != nil {
			t.Fatalf("inspect activeElement: %v", err)
		}
		if v == true {
			focused = true
			break
		}
	}
	if !focused {
		t.Fatal("tabbing never reached a .btn")
	}
	outline, err := page.Evaluate(`() => {
		const s = getComputedStyle(document.activeElement);
		return s.outlineStyle + " " + s.outlineWidth + " " + s.outlineColor;
	}`)
	if err != nil {
		t.Fatalf("evaluate outline: %v", err)
	}
	if outline != "solid 1px rgb(250, 189, 47)" {
		t.Errorf("focus-visible outline = %q, want solid 1px accent", outline)
	}

	// Token flow: swap the theme root class to t-mocha via the tweaks panel —
	// the danger colour must re-resolve to mocha --err (#f38ba8) and primary
	// to mocha --accent (#89b4fa) with no other DOM change.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#btn-variants .btn-danger')).color === 'rgb(243, 139, 168)'`, nil,
	); err != nil {
		got := computedStyle(t, danger, "color")
		t.Fatalf("t-mocha --err never applied to danger btn (color stayed %s): %v", got, err)
	}
	assertStyle(t, primary, "backgroundColor", "rgb(137, 180, 250)", "primary btn (t-mocha)")
}

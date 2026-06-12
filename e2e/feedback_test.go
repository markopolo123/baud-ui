//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// Theme-resolved tones asserted by this suite (assets/css/tokens.css):
//
//	t-gruvbox --ok #b8bb26, --err #fb4934, --fg-faint #7c6f64,
//	          --bg-active #504945
//	t-mocha   --warn #f9e2af, --ok #a6e3a1
const (
	fbGruvOk     = "rgb(184, 187, 38)"
	fbGruvErr    = "rgb(251, 73, 52)"
	fbGruvFaint  = "rgb(124, 111, 100)"
	fbGruvActive = "rgb(80, 73, 69)"
	fbMochaWarn  = "rgb(249, 226, 175)"
	fbMochaOk    = "rgb(166, 227, 161)"
)

// fbSpinnerFrames is the design frame set — the computed ::before content
// must always be one of these (quoted, as computed content serializes).
const fbSpinnerFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// fbPseudoStyle resolves a computed-style property of sel's ::before.
func fbPseudoStyle(t *testing.T, page playwright.Page, sel, prop string) string {
	t.Helper()
	v, err := page.Evaluate(
		`args => { const el = document.querySelector(args.sel); return el ? getComputedStyle(el, '::before')[args.prop] : 'MISSING'; }`,
		map[string]any{"sel": sel, "prop": prop},
	)
	if err != nil {
		t.Fatalf("pseudo %s of %q: %v", prop, sel, err)
	}
	s, ok := v.(string)
	if !ok || s == "MISSING" {
		t.Fatalf("pseudo %s of %q: got %v", prop, sel, v)
	}
	return s
}

// fbEmulateMotion flips the prefers-reduced-motion emulation.
func fbEmulateMotion(t *testing.T, page playwright.Page, reduce bool) {
	t.Helper()
	mode := playwright.ReducedMotionNoPreference
	if reduce {
		mode = playwright.ReducedMotionReduce
	}
	if err := page.EmulateMedia(playwright.PageEmulateMediaOptions{ReducedMotion: mode}); err != nil {
		t.Fatalf("emulate reduced-motion=%v: %v", reduce, err)
	}
}

// fbConfirmBtnDisabled reads the real disabled property of the confirm
// guard's danger button.
func fbConfirmBtnDisabled(t *testing.T, page playwright.Page) bool {
	t.Helper()
	v, err := page.Evaluate(`() => document.querySelector('#fb-confirm .btn-danger').disabled`)
	if err != nil {
		t.Fatalf("read confirm btn disabled: %v", err)
	}
	b, ok := v.(bool)
	if !ok {
		t.Fatalf("confirm btn disabled: got %T(%v)", v, v)
	}
	return b
}

// TestFeedbackProgressTones asserts the auto tone thresholds resolve to the
// theme tokens: accent at 50, ok at 100 and the forced err override under
// t-gruvbox, then — because gruvbox's warn equals its accent — proves the
// warn band plus the token flow on a t-mocha root-class swap (50/90/100 →
// mocha accent/warn/ok).
func TestFeedbackProgressTones(t *testing.T) {
	page := openSheet(t)

	if got := style(t, page, "#fb-prog-50 .prog-bar", "color"); got != gruvAccent {
		t.Errorf("50%% bar color = %q, want %q (--accent)", got, gruvAccent)
	}
	if got := style(t, page, "#fb-prog-100 .prog-bar", "color"); got != fbGruvOk {
		t.Errorf("100%% bar color = %q, want %q (--ok)", got, fbGruvOk)
	}
	if got := style(t, page, "#fb-prog-err .prog-bar", "color"); got != fbGruvErr {
		t.Errorf("forced err bar color = %q, want %q (--err)", got, fbGruvErr)
	}
	if got := style(t, page, "#fb-prog-50 .prog-rest", "color"); got != fbGruvFaint {
		t.Errorf("rest glyph color = %q, want %q (--fg-faint)", got, fbGruvFaint)
	}
	if got := style(t, page, "#fb-prog-50 .prog", "fontVariantNumeric"); got != "tabular-nums" {
		t.Errorf("prog font-variant-numeric = %q, want tabular-nums", got)
	}

	// Theme swap is a root-class change only: the same markup resolves to
	// the mocha tones, separating warn from accent.
	if err := page.Locator(`.tw-theme[data-tweak="t-mocha"]`).Click(); err != nil {
		t.Fatalf("click mocha tweak: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => getComputedStyle(document.querySelector('#fb-prog-50 .prog-bar')).color === 'rgb(137, 180, 250)'`, nil,
	); err != nil {
		t.Fatalf("t-mocha accent never applied to 50%% bar: %v", err)
	}
	if got := style(t, page, "#fb-prog-90 .prog-bar", "color"); got != fbMochaWarn {
		t.Errorf("t-mocha 90%% bar color = %q, want %q (--warn)", got, fbMochaWarn)
	}
	if got := style(t, page, "#fb-prog-100 .prog-bar", "color"); got != fbMochaOk {
		t.Errorf("t-mocha 100%% bar color = %q, want %q (--ok)", got, fbMochaOk)
	}
}

// TestFeedbackSpinner asserts the zero-JS frame cycle is really running in
// the browser: the ::before content starts on a braille frame and changes
// across a predicate wait, and reduced-motion emulation pins the static
// first frame with no animation.
func TestFeedbackSpinner(t *testing.T) {
	page := openSheet(t)
	const sel = "#fb-spinner .spinner"

	if got := style(t, page, sel, "color"); got != gruvAccent {
		t.Errorf("spinner color = %q, want %q (--accent)", got, gruvAccent)
	}
	if got := fbPseudoStyle(t, page, sel, "animationName"); got != "baud-spin" {
		t.Errorf("spinner animation-name = %q, want baud-spin", got)
	}
	frame0 := fbPseudoStyle(t, page, sel, "content")
	if !strings.Contains(fbSpinnerFrames, strings.Trim(frame0, `"`)) || strings.Trim(frame0, `"`) == "" {
		t.Fatalf("initial spinner frame %q is not a braille frame", frame0)
	}
	if _, err := page.WaitForFunction(
		`f0 => getComputedStyle(document.querySelector('#fb-spinner .spinner'), '::before').content !== f0`,
		frame0,
	); err != nil {
		t.Fatalf("spinner frame never changed from %q — content animation not running: %v", frame0, err)
	}
	frameN := fbPseudoStyle(t, page, sel, "content")
	if !strings.Contains(fbSpinnerFrames, strings.Trim(frameN, `"`)) {
		t.Errorf("animated spinner frame %q is not a braille frame", frameN)
	}

	// Reduced motion: no animation, static first frame.
	fbEmulateMotion(t, page, true)
	if got := fbPseudoStyle(t, page, sel, "animationName"); got != "none" {
		t.Errorf("spinner animates under reduced motion: animation-name = %q, want none", got)
	}
	if got := fbPseudoStyle(t, page, sel, "content"); got != `"⠋"` {
		t.Errorf("reduced-motion spinner frame = %q, want the static first frame %q", got, `"⠋"`)
	}
	fbEmulateMotion(t, page, false)
}

// TestFeedbackPanelStates asserts the four states' visuals: staggered
// skeleton bars animating (and static under reduced motion), the loading
// spinner, and the empty/error glyph tones.
func TestFeedbackPanelStates(t *testing.T) {
	page := openSheet(t)

	// Skeleton: --bg-active bars, staggered baud-skel-fade loop.
	if got := style(t, page, "#fb-skel .skel-cell", "backgroundColor"); got != fbGruvActive {
		t.Errorf("skeleton bar background = %q, want %q (--bg-active)", got, fbGruvActive)
	}
	if got := style(t, page, "#fb-skel .skel-cell", "animationName"); got != "baud-skel-fade" {
		t.Errorf("skeleton animation-name = %q, want baud-skel-fade", got)
	}
	if got := style(t, page, "#fb-skel .skel-row:nth-child(2) .skel-cell", "animationDelay"); got != "0.12s" {
		t.Errorf("second skeleton row delay = %q, want 0.12s (stagger)", got)
	}
	if got := style(t, page, "#fb-skel .skel-row:nth-child(5) .skel-cell", "animationDelay"); got != "0.48s" {
		t.Errorf("fifth skeleton row delay = %q, want 0.48s (stagger)", got)
	}
	fbEmulateMotion(t, page, true)
	if got := style(t, page, "#fb-skel .skel-cell", "animationName"); got != "none" {
		t.Errorf("skeleton animates under reduced motion: animation-name = %q, want none", got)
	}
	fbEmulateMotion(t, page, false)

	// Loading: the braille spinner sits in the glyph slot, accent.
	if got := style(t, page, "#fb-loading .pstate-glyph .spinner", "color"); got != gruvAccent {
		t.Errorf("loading spinner color = %q, want %q (--accent)", got, gruvAccent)
	}

	// Empty: faint ∅ glyph; muted title.
	if got := style(t, page, "#fb-empty .pstate-glyph", "color"); got != fbGruvFaint {
		t.Errorf("empty glyph color = %q, want %q (--fg-faint)", got, fbGruvFaint)
	}

	// Error: ✗ and title flip to --err; retry slot present.
	if got := style(t, page, "#fb-error .pstate.err .pstate-glyph", "color"); got != fbGruvErr {
		t.Errorf("error glyph color = %q, want %q (--err)", got, fbGruvErr)
	}
	if got := style(t, page, "#fb-error .pstate.err .pstate-title", "color"); got != fbGruvErr {
		t.Errorf("error title color = %q, want %q (--err)", got, fbGruvErr)
	}
	if n, err := page.Locator("#fb-error .pstate-act .btn").Count(); err != nil || n != 1 {
		t.Errorf("error retry action slot: count = %d (%v), want 1", n, err)
	}
}

// TestFeedbackConfirmInput drives the destructive guard end to end: the
// danger button carries a real disabled attribute until the typed value
// exactly matches data-confirm, a non-empty mismatch shows the err border
// while typing, and clearing re-disables without the err state.
func TestFeedbackConfirmInput(t *testing.T) {
	page := openSheet(t)

	// The guard is hyperscript — wait for the behaviors file to have
	// parsed and booted (ParseHealth sentinel) before typing.
	if err := page.Locator("body[data-hs-ok]").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("hyperscript never booted (body[data-hs-ok]): %v", err)
	}

	input := page.Locator("#fb-confirm-input")
	if err := input.ScrollIntoViewIfNeeded(); err != nil {
		t.Fatalf("scroll confirm input into view: %v", err)
	}
	if !fbConfirmBtnDisabled(t, page) {
		t.Fatalf("confirm button must start disabled")
	}

	// Mismatch: err border while typing, button stays disabled.
	if err := input.Fill("wrong-name"); err != nil {
		t.Fatalf("fill mismatch: %v", err)
	}
	if err := page.Locator("#fb-confirm .input-wrap.err").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("mismatch never flipped the err border: %v", err)
	}
	if got := wrapStyle(t, page, "#fb-confirm-input", "borderTopColor"); got != fbGruvErr {
		t.Errorf("mismatch border = %q, want %q (--err)", got, fbGruvErr)
	}
	if !fbConfirmBtnDisabled(t, page) {
		t.Errorf("confirm button enabled on a mismatch")
	}

	// Exact match: button gains a real enabled state, err border clears.
	if err := input.Fill("db-04"); err != nil {
		t.Fatalf("fill exact: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => !document.querySelector('#fb-confirm .btn-danger').disabled`, nil,
	); err != nil {
		t.Fatalf("exact match never enabled the confirm button: %v", err)
	}
	if err := page.Locator("#fb-confirm .input-wrap:not(.err)").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("exact match never cleared the err border: %v", err)
	}
	if got := wrapStyle(t, page, "#fb-confirm-input", "borderTopColor"); got == fbGruvErr {
		t.Errorf("err border still showing on an exact match")
	}

	// Back to a partial value: disabled + err again (the comparison is
	// exact, not prefix).
	if err := input.Fill("db-0"); err != nil {
		t.Fatalf("fill partial: %v", err)
	}
	if _, err := page.WaitForFunction(
		`() => document.querySelector('#fb-confirm .btn-danger').disabled`, nil,
	); err != nil {
		t.Fatalf("partial value never re-disabled the confirm button: %v", err)
	}
	if err := page.Locator("#fb-confirm .input-wrap.err").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("partial value never flipped the err border: %v", err)
	}

	// Clearing: still disabled, but no err state on an empty value.
	if err := input.Fill(""); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if err := page.Locator("#fb-confirm .input-wrap:not(.err)").WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateAttached,
	}); err != nil {
		t.Fatalf("clearing never dropped the err border: %v", err)
	}
	if !fbConfirmBtnDisabled(t, page) {
		t.Errorf("confirm button stayed enabled after clearing")
	}
}

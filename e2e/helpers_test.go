//go:build e2e

package e2e

// Generic test helpers shared by every e2e file: server/browser bootstrap and
// computed-style resolution. Component test files define ONLY
// component-prefixed helpers (see docs/WAYS_OF_WORKING.md review checklist).

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/playwright-community/playwright-go"

	"github.com/markopolo123/baud-ui/demo"
)

// startDemo starts the real demo handler on a random port.
func startDemo(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(demo.NewMux())
	t.Cleanup(srv.Close)
	return srv
}

// startBrowser launches headless chromium and returns an open page.
func startBrowser(t *testing.T) playwright.Page {
	t.Helper()
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("playwright: %v (run `just install-browsers` first)", err)
	}
	t.Cleanup(func() { pw.Stop() })
	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("launch chromium: %v", err)
	}
	t.Cleanup(func() { browser.Close() })
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("new page: %v", err)
	}
	return page
}

// computedStyle resolves one computed-style property on the first element
// matching the locator.
func computedStyle(t *testing.T, l playwright.Locator, prop string) string {
	t.Helper()
	v, err := l.Evaluate(`(el, prop) => getComputedStyle(el)[prop]`, prop)
	if err != nil {
		t.Fatalf("computed style %q: %v", prop, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed style %q: got %T(%v), want string", prop, v, v)
	}
	return s
}

// computedStyleSel returns the computed style property of the first element
// matching selector.
func computedStyleSel(t *testing.T, page playwright.Page, selector, prop string) string {
	t.Helper()
	v, err := page.Evaluate(fmt.Sprintf(
		`() => { const el = document.querySelector(%q); return el ? getComputedStyle(el)[%q] : "MISSING"; }`,
		selector, prop))
	if err != nil {
		t.Fatalf("computed %s of %q: %v", prop, selector, err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("computed %s of %q: non-string %v", prop, selector, v)
	}
	if s == "MISSING" {
		t.Fatalf("no element matches %q", selector)
	}
	return s
}

//go:build e2e

// Package e2e runs real-browser assertions with playwright-go against the
// demo handler mounted on a random-port test server. Gated behind the e2e
// build tag; run via `just e2e` (after `just install-browsers`).
package e2e

import (
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
	// (give hyperscript a moment — poll until it is non-none).
	if _, err := page.WaitForFunction(
		`() => {
			const el = document.querySelector('[data-panes]');
			return el && getComputedStyle(el).gridTemplateColumns !== 'none';
		}`, nil,
	); err != nil {
		t.Fatalf("Panes behavior never set grid-template-columns: %v", err)
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

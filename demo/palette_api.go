package demo

import (
	"net/http"
	"strings"

	"github.com/markopolo123/baud-ui/baud"
)

// paletteCommands is the in-memory command fixture the live ⌘K palette
// searches over. Row order is load-bearing for e2e/palette_test.go:
// row 0 is the navigation row ↵ exercises, "deploy canary build" is the
// only "deploy" match; action rows carry opaque command ids that the
// sheet's #pal-action-out listener renders on baud:paletteCmd.
var paletteCommands = []baud.PaletteCommand{
	{Category: "go", Label: "go to fleet console", Kbd: "g f", Href: "/"},
	{Category: "go", Label: "go to component sheet", Kbd: "g s", Href: "/sheet"},
	{Category: "fleet", Label: "deploy canary build", Kbd: "⌘D", Action: "deploy-canary"},
	{Category: "fleet", Label: "restart ingest workers", Action: "restart-ingest"},
	{Category: "fleet", Label: "drain batch-runner", Action: "drain-batch-runner"},
	{Category: "view", Label: "tail ingest logs", Kbd: "g l", Href: "/sheet"},
}

// HandlePaletteSearch is the palette's server-filter round-trip: the
// debounced keyup hx-get lands here and the filtered PaletteResults
// fragment is swapped back into the listbox. Filtering is a plain
// case-insensitive substring match over category + label, so arbitrary
// (garbage) queries simply match nothing and render the ∅ empty state.
func HandlePaletteSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	var hits []baud.PaletteCommand
	for _, c := range paletteCommands {
		if q == "" || strings.Contains(strings.ToLower(c.Category+" "+c.Label), q) {
			hits = append(hits, c)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	res := baud.PaletteResultsProps{ID: "pal-live", Commands: hits}
	if err := baud.PaletteResults(res).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

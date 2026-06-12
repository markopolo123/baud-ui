package demo

import "net/http"

// tabsRanges fixes the demo's metrics windows: range → sample count
// rendered into the pane, so each round-trip produces distinct,
// assertable content.
var tabsRanges = map[string]string{
	"5m": "60",
	"1h": "720",
	"1d": "17,280",
	"1w": "120,960",
}

func tabsSamples(rng string) string {
	if n, ok := tabsRanges[rng]; ok {
		return n
	}
	return "0"
}

// tabsPane serves GET /demo/tabs?range=… — the htmx tab round-trip. Each
// boxed tab on the component sheet hx-gets this fragment into the shared
// #tabs-range-pane tabpanel; the server owns the pane rendering.
func tabsPane(w http.ResponseWriter, r *http.Request) {
	rng := r.URL.Query().Get("range")
	if _, ok := tabsRanges[rng]; !ok {
		http.Error(w, "unknown range", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	TabsRangePane(rng).Render(r.Context(), w)
}

// tabsCount is the sheet's *int literal helper for Tab.Count.
func tabsCount(n int) *int { return &n }

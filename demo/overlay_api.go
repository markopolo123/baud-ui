package demo

import "net/http"

// handleModalDemo serves GET /demo/modal — the overlay pattern's htmx
// round-trip: the opener button targets body/beforeend and this fragment
// (an Overlay-installing Modal) opens on arrival and removes itself from
// the DOM on close.
func handleModalDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	ovDemoModal().Render(r.Context(), w)
}

// handleDrawerDemo serves GET /demo/drawer — same delivery as
// handleModalDemo for the right-side Drawer variant.
func handleDrawerDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	ovDemoDrawer().Render(r.Context(), w)
}

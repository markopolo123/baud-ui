// Package demo is the fleetctl demo console + component sheet. It is a
// plain stdlib net/http app; cmd/demo serves it, cmd/render writes it to
// disk for GitHub Pages, and e2e tests mount it on httptest servers.
package demo

import (
	"net/http"

	"github.com/a-h/templ"

	baudui "github.com/markopolo123/baud-ui"
)

// NewMux returns the demo HTTP handler.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/baud.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(baudui.CSS)
	})
	mux.HandleFunc("GET /assets/baud._hs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/hyperscript; charset=utf-8")
		w.Write(baudui.Behaviors)
	})
	mux.HandleFunc("GET /api/combobox", HandleComboboxSearch)
	mux.Handle("GET /{$}", templ.Handler(AppPage(ServerOpts())))
	mux.HandleFunc("GET /api/datepicker", handleDatePickerMenu)
	mux.Handle("GET /sheet", templ.Handler(SheetPage(ServerOpts())))
	mux.HandleFunc("GET /demo/tabs", tabsPane)
	return mux
}

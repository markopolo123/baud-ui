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
		w.Write(baudui.CSS())
	})
	mux.HandleFunc("GET /assets/baud._hs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/hyperscript; charset=utf-8")
		w.Write(baudui.HS())
	})
	mux.HandleFunc("GET /api/combobox", HandleComboboxSearch)
	mux.HandleFunc("GET /{$}", handleFleetConsole)
	mux.HandleFunc("GET /api/datepicker", handleDatePickerMenu)
	mux.Handle("GET /sheet", templ.Handler(SheetPage(ServerOpts())))
	mux.HandleFunc("GET /demo/tabs", tabsPane)
	mux.HandleFunc("GET /demo/pagination", handlePagination)
	mux.HandleFunc("GET /demo/datatable", handleDataTableSort)
	mux.HandleFunc("GET /demo/tree", treeChildren)
	mux.HandleFunc("GET /demo/toast", handleToast)
	mux.HandleFunc("GET /demo/modal", handleModalDemo)
	mux.HandleFunc("GET /demo/drawer", handleDrawerDemo)
	mux.HandleFunc("GET /api/palette", HandlePaletteSearch)
	// fleetctl console endpoints (demo/fleetctl_api.go).
	mux.HandleFunc("GET /fleet/hosts", handleFleetHosts)
	mux.HandleFunc("GET /fleet/tab", handleFleetTab)
	mux.HandleFunc("GET /fleet/host", handleFleetHost)
	mux.HandleFunc("GET /fleet/kill", handleFleetKill)
	mux.HandleFunc("GET /fleet/deploy", handleFleetDeploy)
	mux.HandleFunc("GET /fleet/deploy/run", handleFleetDeployRun)
	mux.HandleFunc("GET /fleet/tree", handleFleetTree)
	mux.HandleFunc("GET /fleet/palette", handleFleetPalette)
	mux.HandleFunc("GET /fleet/cmd", handleFleetCmd)
	return mux
}

package demo

import (
	"net/http"
	"time"

	"github.com/markopolo123/baud-ui/baud"
)

// handleDatePickerMenu re-renders the DatePicker menu fragment for the
// month (+ optional selected date) in the query string. The «‹›» nav
// buttons point their hx-get here — the grid is always server-computed,
// there is no client date math. Consumers of the library mount their own
// copy of this handler and pass its path as DatePickerProps.Endpoint.
func handleDatePickerMenu(w http.ResponseWriter, r *http.Request) {
	month, err := time.Parse("2006-01", r.URL.Query().Get("month"))
	if err != nil {
		http.Error(w, "bad month: want YYYY-MM", http.StatusBadRequest)
		return
	}
	p := baud.DatePickerProps{Month: month, Endpoint: r.URL.Path}
	if s := r.URL.Query().Get("selected"); s != "" {
		sel, err := time.Parse("2006-01-02", s)
		if err != nil {
			http.Error(w, "bad selected: want YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		p.Selected = sel
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := baud.DatePickerMenu(p).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

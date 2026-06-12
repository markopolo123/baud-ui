package demo

import (
	"net/http"
	"strings"

	"github.com/markopolo123/baud-ui/baud"
)

// comboHosts is the in-memory fixture list both combobox demos search
// over (label = host, meta = region).
var comboHosts = []baud.SelectOption{
	{Value: "ingest-gw-1", Meta: "eu-west"},
	{Value: "ingest-gw-2", Meta: "eu-west"},
	{Value: "api-core", Meta: "us-east"},
	{Value: "api-edge", Meta: "us-east"},
	{Value: "batch-runner", Meta: "ap-south"},
	{Value: "metrics-db", Meta: "eu-west"},
	{Value: "log-tail", Meta: "ap-south"},
}

// HandleComboboxSearch is the server-mode combobox round-trip: the
// debounced hx-get lands here and the filtered ComboboxOptions fragment
// is swapped back into the menu, with the match highlighted server-side.
// The query arrives under the input's name — "q" on the component sheet.
func HandleComboboxSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	needle := strings.ToLower(q)
	var hits []baud.SelectOption
	for _, o := range comboHosts {
		if needle == "" || strings.Contains(strings.ToLower(o.Value+" "+o.Meta), needle) {
			hits = append(hits, o)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	opts := baud.ComboboxOptionsProps{ID: "cb-server", Query: q, Options: hits}
	if err := baud.ComboboxOptions(opts).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

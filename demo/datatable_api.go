package demo

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/markopolo123/baud-ui/baud"
)

// dtHost is one fleet host row — the in-memory fixture behind the sheet's
// sortable DataTable.
type dtHost struct {
	ID, Name, Region, Status string
	CPU, Mem                 float64
}

var dtFleet = []dtHost{
	{"h1", "edge-cache", "use1", "ok", 22.5, 95.1},
	{"h2", "ingest-gw", "euw1", "err", 96.7, 88.9},
	{"h3", "search-idx", "usw2", "warn", 58.3, 71.5},
	{"h4", "auth-svc", "use1", "ok", 82.4, 61.0},
	{"h5", "notif-fan", "euw1", "ok", 12.9, 23.1},
	{"h6", "media-proc", "aps1", "warn", 77.0, 90.3},
	{"h7", "pg-primary", "use1", "ok", 61.2, 78.0},
	{"h8", "kafka-brk", "use1", "ok", 48.9, 66.4},
}

var dtColumns = []baud.Column{
	{Key: "host", Label: "Host", Sortable: true},
	{Key: "region", Label: "Reg"},
	{Key: "cpu", Label: "CPU%", Numeric: true, Sortable: true},
	{Key: "mem", Label: "Mem%", Numeric: true, Sortable: true},
	{Key: "status", Label: "St"},
}

// dtLess holds the ascending comparator per sortable column key — it is
// also the sort-parameter allowlist the handler validates against.
var dtLess = map[string]func(a, b dtHost) bool{
	"host": func(a, b dtHost) bool { return a.Name < b.Name },
	"cpu":  func(a, b dtHost) bool { return a.CPU < b.CPU },
	"mem":  func(a, b dtHost) bool { return a.Mem < b.Mem },
}

func dtPct(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

// dtCellTone is the sheet's cell-tone hook: cpu/mem threshold colours plus
// the status column mapped straight to its tone.
func dtCellTone(c baud.Column, v string) baud.Tone {
	switch c.Key {
	case "cpu":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return baud.ToneNone
		}
		switch {
		case f > 90:
			return baud.ToneErr
		case f > 70:
			return baud.ToneWarn
		}
	case "mem":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return baud.ToneNone
		}
		switch {
		case f > 90:
			return baud.ToneErr
		case f > 75:
			return baud.ToneWarn
		}
	case "status":
		switch v {
		case "ok":
			return baud.ToneOK
		case "warn":
			return baud.ToneWarn
		case "err":
			return baud.ToneErr
		}
	}
	return baud.ToneNone
}

// dtFleetProps assembles the fleet table for one sort state: stable-sorted
// hosts rendered to rows, the props mirroring the query string so the
// header URLs flip correctly. Selection is server state — h3 is pinned.
func dtFleetProps(key, dir string) baud.DataTableProps {
	hosts := make([]dtHost, len(dtFleet))
	copy(hosts, dtFleet)
	less := dtLess[key]
	sort.SliceStable(hosts, func(i, j int) bool {
		if dir == "desc" {
			return less(hosts[j], hosts[i])
		}
		return less(hosts[i], hosts[j])
	})
	rows := make([]baud.Row, len(hosts))
	for i, h := range hosts {
		rows[i] = baud.Row{Key: h.ID, Cells: []string{h.Name, h.Region, dtPct(h.CPU), dtPct(h.Mem), h.Status}}
	}
	return baud.DataTableProps{
		ID:       "dt-fleet",
		Columns:  dtColumns,
		Rows:     rows,
		Endpoint: "/demo/datatable",
		SortKey:  key,
		SortDir:  dir,
		Selected: "h3",
		CellTone: dtCellTone,
	}
}

// dtVariantProps is a small static table for the zebra/lines variant demos
// — no endpoint, so the headers render inert.
func dtVariantProps(id string, zebra, lines bool) baud.DataTableProps {
	return baud.DataTableProps{
		ID: id,
		Columns: []baud.Column{
			{Key: "region", Label: "Region"},
			{Key: "hosts", Label: "Hosts", Numeric: true},
			{Key: "rps", Label: "RPS", Numeric: true},
		},
		Rows: []baud.Row{
			{Key: "use1", Cells: []string{"use1", "14", "91,244"}},
			{Key: "euw1", Cells: []string{"euw1", "6", "48,022"}},
			{Key: "usw2", Cells: []string{"usw2", "9", "22,310"}},
			{Key: "aps1", Cells: []string{"aps1", "2", "422"}},
		},
		Zebra: zebra,
		Lines: lines,
	}
}

// handleDataTableSort serves GET /demo/datatable?sort=…&dir=… — the htmx
// column-sort round-trip. It validates both parameters (unknown column or
// direction ⇒ 400), stable-sorts the fixture and renders the tbody
// fragment the th targeted, followed by the out-of-band thead so the ▲/▼
// indicator and flipped URLs track the new state.
func handleDataTableSort(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("sort")
	dir := r.URL.Query().Get("dir")
	if _, ok := dtLess[key]; !ok {
		http.Error(w, "unknown sort column", http.StatusBadRequest)
		return
	}
	if dir != "asc" && dir != "desc" {
		http.Error(w, "bad dir: want asc|desc", http.StatusBadRequest)
		return
	}
	p := dtFleetProps(key, dir)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := baud.DataTableBody(p).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := baud.DataTableHead(p, true).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

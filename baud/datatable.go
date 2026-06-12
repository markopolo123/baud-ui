package baud

import "net/url"

// Column describes one DataTable column.
type Column struct {
	// Key is the stable column identifier: the ?sort= parameter value and
	// the th's data-key attribute.
	Key string
	// Label is the header text (rendered UPPERCASE by the type rules).
	Label string
	// Numeric right-aligns the column (tabular-nums comes from the table).
	Numeric bool
	// Sortable wires the header to the htmx sort round-trip. It needs
	// DataTableProps.Endpoint and ID set — without them the header renders
	// inert.
	Sortable bool
}

// Row is one data row: a stable key (selection is server state, matched
// against DataTableProps.Selected) plus one pre-rendered cell string per
// column.
type Row struct {
	Key   string
	Cells []string
}

// cell returns the i-th cell, tolerating short rows.
func (r Row) cell(i int) string {
	if i < len(r.Cells) {
		return r.Cells[i]
	}
	return ""
}

// Tone is the semantic colour a cell-tone hook applies to one cell — the
// threshold-colouring mechanism (e.g. cpu > 90 ⇒ ToneErr).
type Tone string

const (
	ToneNone   Tone = ""
	ToneOK     Tone = "ok"
	ToneWarn   Tone = "warn"
	ToneErr    Tone = "err"
	ToneInfo   Tone = "info"
	ToneAccent Tone = "accent"
	ToneFaint  Tone = "faint"
)

// DataTableProps configures a DataTable (design/README.md "Data"). Sort is
// a server round-trip: each sortable th hx-gets Endpoint?sort=<key>&dir=…
// swapping the tbody fragment (id <ID>-body), and the server's response
// appends DataTableHead(p, true) so the thead (id <ID>-head) re-renders
// out-of-band — the accent ▲/▼ indicator and the flipped hx-get URLs track
// the new state. The URL is the state: the active column's th flips
// asc⇄desc, every other sortable column offers asc.
type DataTableProps struct {
	// ID names the table and derives the fragment ids <ID>-body and
	// <ID>-head. Required for sorting (the th hx-target needs it).
	ID      string
	Columns []Column
	Rows    []Row
	// Zebra tints even rows with a half-strength --bg-raised wash.
	Zebra bool
	// Lines adds hairline column rules between cells.
	Lines bool
	// Endpoint is the sort URL base, e.g. "/demo/datatable". Empty
	// disables sorting regardless of Column.Sortable.
	Endpoint string
	// SortKey/SortDir mirror the query string this render answered
	// ("" = unsorted). SortDir is "asc" (default) or "desc".
	SortKey string
	SortDir string
	// Selected is the Row.Key of the selected row ("" = none): --sel fill,
	// 2px accent inset bar and the ▌ row mark. Selection is server state.
	Selected string
	// CellTone is the optional cell-tone hook: it maps a column + rendered
	// cell value to a Tone class (tone-err text colour etc.). ToneNone
	// leaves the cell unstyled.
	CellTone func(col Column, value string) Tone
}

func (p DataTableProps) bodyID() string { return p.ID + "-body" }
func (p DataTableProps) headID() string { return p.ID + "-head" }

// sortable reports whether a column header issues the hx-get round-trip.
func (p DataTableProps) sortable(c Column) bool {
	return c.Sortable && p.Endpoint != "" && p.ID != ""
}

// sorted reports whether c is the active sort column.
func (p DataTableProps) sorted(c Column) bool {
	return p.SortKey != "" && p.SortKey == c.Key
}

// dir normalizes SortDir: anything but "desc" is ascending.
func (p DataTableProps) dir() string {
	if p.SortDir == "desc" {
		return "desc"
	}
	return "asc"
}

// sortURL builds a header's next-state URL: the active column flips
// asc⇄desc, inactive columns start asc.
func (p DataTableProps) sortURL(c Column) string {
	dir := "asc"
	if p.sorted(c) && p.dir() == "asc" {
		dir = "desc"
	}
	return p.Endpoint + "?sort=" + url.QueryEscape(c.Key) + "&dir=" + dir
}

// ariaSort is the active column's aria-sort value.
func (p DataTableProps) ariaSort() string {
	if p.dir() == "desc" {
		return "descending"
	}
	return "ascending"
}

// arrow is the active column's indicator glyph.
func (p DataTableProps) arrow() string {
	if p.dir() == "desc" {
		return "▼"
	}
	return "▲"
}

// thClass assembles a header cell's class list.
func (p DataTableProps) thClass(c Column) string {
	cls := ""
	if c.Numeric {
		cls += " num"
	}
	if p.sortable(c) {
		cls += " sortable"
	}
	if p.sorted(c) {
		cls += " sorted"
	}
	if cls == "" {
		return ""
	}
	return cls[1:]
}

// tdClass assembles a data cell's class list: the num marker plus the
// cell-tone hook's tone class.
func (p DataTableProps) tdClass(c Column, value string) string {
	cls := ""
	if c.Numeric {
		cls += " num"
	}
	if p.CellTone != nil {
		if tone := p.CellTone(c, value); tone != ToneNone {
			cls += " tone-" + string(tone)
		}
	}
	if cls == "" {
		return ""
	}
	return cls[1:]
}

// isSelected reports whether r is the selected row.
func (p DataTableProps) isSelected(r Row) bool {
	return p.Selected != "" && r.Key == p.Selected
}

// mark is the row-mark cell content: ▌ on the selected row.
func (p DataTableProps) mark(r Row) string {
	if p.isSelected(r) {
		return "▌"
	}
	return ""
}

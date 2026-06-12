package baud

import (
	"strconv"
	"strings"
)

// PaginationProps configures the pager footer bar (design/README.md
// "Data"): `rows 1–50 of 12,403` + |‹ ‹ prev n/N next › ›| + an optional
// accent "load more ↓" htmx append. Two navigation modes:
//
//   - htmx (HxGet != ""): every enabled step button issues
//     hx-get={HxGet}?page=N into #Target — the server re-renders the
//     region (list + pager). URL is the state.
//   - link (HrefFor != nil): enabled steps render as plain anchors with
//     href=HrefFor(page) — for static pages (cmd/render) or full loads.
//
// Bound buttons (first/prev on page 1, next/last on the last page) always
// render as real disabled <button>s in both modes.
type PaginationProps struct {
	// Page is the 1-based current page.
	Page int
	// PerPage is the page size used for the range text and page count.
	PerPage int
	// Total is the full row count (formatted with thousands separators).
	Total int
	// HxGet is the htmx-mode base URL, without a query string; buttons
	// append ?page=N.
	HxGet string
	// Target is the id (no "#") of the region each step hx-gets into.
	Target string
	// HrefFor selects link mode (ignored when HxGet is set): it maps a
	// page number to the anchor href.
	HrefFor func(page int) string
	// MoreURL, when set, renders the accent "load more ↓" button: an htmx
	// append — hx-get={MoreURL} hx-swap="beforeend" into #MoreTarget.
	MoreURL    string
	MoreTarget string
}

// lastPage is the 1-based final page number — never below 1, so an empty
// result still reads "1/1" with both bounds disabled.
func (p PaginationProps) lastPage() int {
	if p.PerPage <= 0 || p.Total <= 0 {
		return 1
	}
	return (p.Total + p.PerPage - 1) / p.PerPage
}

// from is the 1-based index of the first visible row (0 when empty).
func (p PaginationProps) from() int {
	if p.Total <= 0 {
		return 0
	}
	return (p.Page-1)*p.PerPage + 1
}

// to is the index of the last visible row — partial on the final page.
func (p PaginationProps) to() int {
	t := p.Page * p.PerPage
	if t > p.Total {
		t = p.Total
	}
	return t
}

// rangeText renders the info cell: "rows 1–50 of 12,403" (en dash,
// thousands separators; tabular-nums comes from the .pager CSS).
func (p PaginationProps) rangeText() string {
	return "rows " + paginationThousands(p.from()) + "–" + paginationThousands(p.to()) +
		" of " + paginationThousands(p.Total)
}

// posText is the n/N position cell between prev and next.
func (p PaginationProps) posText() string {
	return strconv.Itoa(p.Page) + "/" + strconv.Itoa(p.lastPage())
}

func (p PaginationProps) htmx() bool { return p.HxGet != "" }

// linkMode reports href-anchor navigation: HrefFor set and no HxGet.
func (p PaginationProps) linkMode() bool { return !p.htmx() && p.HrefFor != nil }

// pageURL is the htmx round-trip URL for one step button.
func (p PaginationProps) pageURL(page int) string {
	return p.HxGet + "?page=" + strconv.Itoa(page)
}

// paginationThousands formats n with comma thousands separators.
func paginationThousands(n int) string {
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

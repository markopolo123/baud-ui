package demo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/markopolo123/baud-ui/baud"
)

// The pagination demo pages one in-memory fixture set: pagerTotal rows at
// pagerPerPage per page — a total chosen to exercise the thousands
// separator and a partial final page (12,401–12,403).
const (
	pagerPerPage = 5
	pagerTotal   = 12403
)

func pagerLastPage() int { return (pagerTotal + pagerPerPage - 1) / pagerPerPage }

// handlePagination serves GET /demo/pagination?page=N[&append=1] — the
// pager htmx round-trip. Plain requests re-render the whole #pager-demo
// region (list + pager); append=1 returns only the page's rows, for the
// "load more ↓" beforeend swap. An absent, non-numeric or out-of-range
// page is a 400.
func handlePagination(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 || page > pagerLastPage() {
		http.Error(w, "page must be 1…"+strconv.Itoa(pagerLastPage()), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.URL.Query().Get("append") == "1" {
		PagerRows(page).Render(r.Context(), w)
		return
	}
	PagerRegion(page).Render(r.Context(), w)
}

// pagerRowLabels renders one page of fixture rows. Zero-padded so e2e
// substring assertions are unambiguous (row 0001 vs row 0010).
func pagerRowLabels(page int) []string {
	from := (page-1)*pagerPerPage + 1
	to := page * pagerPerPage
	if to > pagerTotal {
		to = pagerTotal
	}
	out := make([]string, 0, pagerPerPage)
	for i := from; i <= to; i++ {
		out = append(out, fmt.Sprintf("row %04d — fixture", i))
	}
	return out
}

// pagerHref is the link-mode demo's HrefFor: query-param navigation.
func pagerHref(page int) string { return "?page=" + strconv.Itoa(page) }

// dataExtrasDiffSrc is the DiffViewer demo fixture — a real two-hunk
// unified diff run through baud.ParseUnified at render time.
const dataExtrasDiffSrc = `--- a/cmd/fleetctl/main.go
+++ b/cmd/fleetctl/main.go
@@ -12,6 +12,7 @@ func main() {
 	cfg := loadConfig()
-	pool := newPool(cfg, 4)
+	pool := newPool(cfg, 8)
+	pool.Audit = true
 	if err := pool.Start(); err != nil {
 		fatal(err)
 	}
@@ -41,5 +42,4 @@ func shutdown(p *Pool) {
 	p.Drain()
-	p.FlushLogs()
 	log.Println("bye")
 }
`

// dataExtrasDiffLines parses the sheet's diff fixture (constant input,
// covered by the baud unit tests — an error here means a broken fixture).
func dataExtrasDiffLines() []baud.DiffLine {
	lines, err := baud.ParseUnified(dataExtrasDiffSrc)
	if err != nil {
		return []baud.DiffLine{{Kind: baud.DiffHunk, Text: "@@ fixture parse error: " + err.Error() + " @@"}}
	}
	return lines
}

package demo

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"

	"github.com/markopolo123/baud-ui/baud"
)

// ---- fleet fixture -------------------------------------------------------
//
// fleetSvc is one fleet service row — the in-memory state behind the
// console (no database; design/baud-app.jsx is the copy reference).
type fleetSvc struct {
	Svc, Ver, Region, Pods string
	CPU, Mem               float64
	RPS, P99               int
	ErrPct                 float64
	Status                 string // ok | warn | err
}

var fleetSvcs = []fleetSvc{
	{"auth-svc", "v3.2.1", "use1", "9/9", 82.4, 61.0, 12480, 41, 0.02, "ok"},
	{"billing-api", "v1.9.0", "use1", "4/4", 34.1, 44.2, 3211, 87, 0.00, "ok"},
	{"ingest-gw", "v2.14.0", "euw1", "9/12", 96.7, 88.9, 48022, 312, 4.81, "err"},
	{"search-idx", "v5.0.3", "usw2", "6/6", 58.3, 71.5, 8904, 64, 0.11, "warn"},
	{"search-query", "v5.0.3", "usw2", "8/8", 41.7, 52.8, 22310, 38, 0.01, "ok"},
	{"notif-fan", "v0.8.2", "euw1", "3/3", 12.9, 23.1, 1502, 18, 0.00, "ok"},
	{"media-proc", "v4.1.0", "aps1", "5/5", 77.0, 90.3, 422, 1240, 0.92, "warn"},
	{"edge-cache", "v7.3.3", "use1", "16/16", 22.5, 95.1, 91244, 9, 0.00, "ok"},
	{"pg-primary", "15.6", "use1", "1/1", 61.2, 78.0, 5811, 22, 0.00, "ok"},
	{"kafka-brk", "3.7.0", "use1", "5/5", 48.9, 66.4, 30122, 14, 0.03, "ok"},
	{"webhook-out", "v1.2.7", "euw1", "2/2", 8.3, 19.9, 311, 92, 1.20, "warn"},
	{"flag-svc", "v2.0.0", "usw2", "3/3", 5.1, 12.2, 7240, 6, 0.00, "ok"},
}

// fleetSvcByName looks one service up by its (unique) name — the row key,
// selection value and drawer/kill parameter.
func fleetSvcByName(name string) (fleetSvc, bool) {
	for _, s := range fleetSvcs {
		if s.Svc == name {
			return s, true
		}
	}
	return fleetSvc{}, false
}

// fleetSelected pins the console's server-state selection: the SEV-1 host.
const fleetSelected = "ingest-gw"

// ---- hosts table (htmx sort) ---------------------------------------------

var fleetColumns = []baud.Column{
	{Key: "svc", Label: "Service", Sortable: true},
	{Key: "ver", Label: "Ver"},
	{Key: "region", Label: "Reg"},
	{Key: "pods", Label: "Pods", Numeric: true},
	{Key: "cpu", Label: "CPU%", Numeric: true, Sortable: true},
	{Key: "mem", Label: "Mem%", Numeric: true, Sortable: true},
	{Key: "rps", Label: "RPS", Numeric: true, Sortable: true},
	{Key: "p99", Label: "p99", Numeric: true, Sortable: true},
	{Key: "err", Label: "Err%", Numeric: true, Sortable: true},
	{Key: "st", Label: "St"},
}

// fleetLess holds the ascending comparator per sortable column key — also
// the ?sort= allowlist handleFleetHosts validates against.
var fleetLess = map[string]func(a, b fleetSvc) bool{
	"svc": func(a, b fleetSvc) bool { return a.Svc < b.Svc },
	"cpu": func(a, b fleetSvc) bool { return a.CPU < b.CPU },
	"mem": func(a, b fleetSvc) bool { return a.Mem < b.Mem },
	"rps": func(a, b fleetSvc) bool { return a.RPS < b.RPS },
	"p99": func(a, b fleetSvc) bool { return a.P99 < b.P99 },
	"err": func(a, b fleetSvc) bool { return a.ErrPct < b.ErrPct },
}

// The console's default sort mirrors the incident at hand: worst error
// rate first.
const (
	fleetDefaultSort = "err"
	fleetDefaultDir  = "desc"
)

func fleetPct(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

func fleetErrPct(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }

// fleetThousands formats with comma separators (RPS column).
func fleetThousands(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// fleetCellTone is the console's threshold-colouring hook
// (DataTableProps.CellTone): cpu/mem/p99/err thresholds from the design
// prototype plus the status column mapped straight to its tone.
func fleetCellTone(c baud.Column, v string) baud.Tone {
	switch c.Key {
	case "cpu":
		return fleetFloatTone(v, 90, 70)
	case "mem":
		return fleetFloatTone(v, 90, 75)
	case "p99":
		return fleetFloatTone(v, 300, 100)
	case "err":
		t := fleetFloatTone(v, 1, 0.1)
		if t == baud.ToneNone {
			return baud.ToneFaint
		}
		return t
	case "st":
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

// fleetFloatTone parses a rendered numeric cell and applies err/warn
// thresholds (strictly greater, per the prototype).
func fleetFloatTone(v string, errAt, warnAt float64) baud.Tone {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return baud.ToneNone
	}
	switch {
	case f > errAt:
		return baud.ToneErr
	case f > warnAt:
		return baud.ToneWarn
	}
	return baud.ToneNone
}

// fleetToneClass maps a Tone onto the console's scoped text-tone classes
// in assets/css/components/fleetctl.css ("" stays unstyled).
func fleetToneClass(t baud.Tone) string {
	if t == baud.ToneNone {
		return ""
	}
	return "fl-" + string(t)
}

// Drawer vitals reuse the table's exact thresholds via the rendered cell
// strings, so the drawer and the table can never disagree on a tone.
func fleetCPUClass(h fleetSvc) string { return fleetToneClass(fleetFloatTone(fleetPct(h.CPU), 90, 70)) }
func fleetMemClass(h fleetSvc) string { return fleetToneClass(fleetFloatTone(fleetPct(h.Mem), 90, 75)) }
func fleetP99Class(h fleetSvc) string {
	return fleetToneClass(fleetFloatTone(fleetP99Text(h), 300, 100))
}
func fleetErrClass(h fleetSvc) string {
	return fleetToneClass(fleetFloatTone(fleetErrPct(h.ErrPct), 1, 0.1))
}

func fleetP99Text(h fleetSvc) string { return strconv.Itoa(h.P99) }

// fleetShownText is the hosts-panel count chip ("12 shown").
func fleetShownText() string { return strconv.Itoa(len(fleetSvcs)) + " shown" }

// fleetHostsProps assembles the hosts table for one sort state, mirroring
// the query string so the header URLs flip correctly. Selection is server
// state: the SEV-1 host stays pinned.
func fleetHostsProps(key, dir string) baud.DataTableProps {
	hosts := make([]fleetSvc, len(fleetSvcs))
	copy(hosts, fleetSvcs)
	less := fleetLess[key]
	sort.SliceStable(hosts, func(i, j int) bool {
		if dir == "desc" {
			return less(hosts[j], hosts[i])
		}
		return less(hosts[i], hosts[j])
	})
	rows := make([]baud.Row, len(hosts))
	for i, h := range hosts {
		rows[i] = baud.Row{Key: h.Svc, Cells: []string{
			h.Svc, h.Ver, h.Region, h.Pods,
			fleetPct(h.CPU), fleetPct(h.Mem),
			fleetThousands(h.RPS), strconv.Itoa(h.P99),
			fleetErrPct(h.ErrPct), h.Status,
		}}
	}
	return baud.DataTableProps{
		ID:       "fleet-hosts",
		Columns:  fleetColumns,
		Rows:     rows,
		Endpoint: "/fleet/hosts",
		SortKey:  key,
		SortDir:  dir,
		Selected: fleetSelected,
		CellTone: fleetCellTone,
	}
}

// handleFleetHosts serves GET /fleet/hosts?sort=…&dir=… — the console's
// column-sort round-trip (the datatable_api.go endpoint pattern): both
// parameters validated (400 on garbage), tbody fragment first, then the
// out-of-band thead so the ▲/▼ indicator and flipped URLs track the state.
func handleFleetHosts(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("sort")
	dir := r.URL.Query().Get("dir")
	if _, ok := fleetLess[key]; !ok {
		http.Error(w, "unknown sort column", http.StatusBadRequest)
		return
	}
	if dir != "asc" && dir != "desc" {
		http.Error(w, "bad dir: want asc|desc", http.StatusBadRequest)
		return
	}
	p := fleetHostsProps(key, dir)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := baud.DataTableBody(p).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := baud.DataTableHead(p, true).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ---- log tail fixture ------------------------------------------------------

// fleetLogLine is one log-tail row: timestamp, level (info|warn|err), pod,
// message — static fixture state, tone-coded by level in the template.
type fleetLogLine struct {
	Time, Level, Pod, Msg string
}

var fleetLogLines = []fleetLogLine{
	{"14:32:41.802", "err", "ingest-gw-7f9c", "OOMKilled — container exceeded 2Gi limit"},
	{"14:32:38.114", "warn", "ingest-gw-7f9c", "GC pause 412ms exceeded threshold"},
	{"14:32:35.660", "info", "edge-cache-c2a1", "purged 14,203 keys (deploy hook)"},
	{"14:32:31.090", "err", "ingest-gw-44d0", "liveness probe failed: Get /healthz: context deadline"},
	{"14:32:28.554", "info", "auth-svc-9b21", "rotated signing keys ok"},
	{"14:32:25.013", "warn", "media-proc-1e8f", "transcode queue depth 1,422 (slo 500)"},
	{"14:32:21.371", "info", "kafka-brk-3", "ISR shrink topic=events partition=7"},
	{"14:32:18.876", "info", "flag-svc-0a44", "flag \"new-ranker\" 25% → 50%"},
	{"14:32:14.209", "err", "ingest-gw-44d0", "upstream connect error: 503 UC upstream_reset"},
	{"14:32:09.991", "info", "webhook-out-77b2", "retry budget exhausted for dest=hooks.stripe.com"},
}

// ---- tab panes -------------------------------------------------------------

// FleetTabPane resolves a topbar tab view to its server-rendered pane —
// the GET /fleet/tab fragment and the BDD render hook (false = unknown
// view). The fleet pane is exactly the console's initial main content, so
// returning to the tab re-renders fresh server state (sort resets — the
// URL is the state and the tab URL carries none).
func FleetTabPane(view string) (templ.Component, bool) {
	switch view {
	case "fleet":
		return fleetFleetPane(), true
	case "incidents":
		return fleetIncidentsPane(), true
	case "deploys":
		return fleetDeploysPane(), true
	}
	return nil, false
}

// handleFleetTab serves GET /fleet/tab?view=… — the topbar-tabs htmx
// round-trip into the shared #fleet-view tabpanel. Unknown views are 400.
func handleFleetTab(w http.ResponseWriter, r *http.Request) {
	pane, ok := FleetTabPane(r.URL.Query().Get("view"))
	if !ok {
		http.Error(w, "unknown view: want fleet|incidents|deploys", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	pane.Render(r.Context(), w)
}

// ---- host drawer + kill guard ----------------------------------------------

// fleetDiffSrc is the drawer's config-change fixture: the capacity bump
// proposed for the OOMKilled ingest gateway, parsed by baud.ParseUnified
// at render time.
const fleetDiffSrc = `--- a/deploy/ingest-gw.yaml
+++ b/deploy/ingest-gw.yaml
@@ -4,3 +4,3 @@ spec:
 spec:
-  replicas: 9
+  replicas: 12
   containers:
@@ -18,2 +18,3 @@ resources:
   limits:
-    memory: 2Gi
+    memory: 4Gi
+    cpu: "2"
`

func fleetDiffLines() []baud.DiffLine {
	lines, err := baud.ParseUnified(fleetDiffSrc)
	if err != nil {
		return []baud.DiffLine{{Kind: baud.DiffHunk, Text: "@@ fixture parse error: " + err.Error() + " @@"}}
	}
	return lines
}

// FleetHostDrawer resolves a service name to its detail-drawer fragment —
// the GET /fleet/host fragment and the BDD render hook (false = unknown
// service).
func FleetHostDrawer(name string) (templ.Component, bool) {
	h, ok := fleetSvcByName(name)
	if !ok {
		return nil, false
	}
	return fleetHostDrawer(h), true
}

// handleFleetHost serves GET /fleet/host?id=… — the host-detail drawer,
// hx-got into body/beforeend by the inspect action. Unknown ids are 400.
func handleFleetHost(w http.ResponseWriter, r *http.Request) {
	drawer, ok := FleetHostDrawer(r.URL.Query().Get("id"))
	if !ok {
		http.Error(w, "unknown host id", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	drawer.Render(r.Context(), w)
}

// handleFleetKill serves GET /fleet/kill?host=… — the destructive action
// behind the drawer's ConfirmInput guard. The trigger uses hx-swap="none":
// the response's only payload is the OOB toast confirming the kill.
// Unknown hosts are 400.
func handleFleetKill(w http.ResponseWriter, r *http.Request) {
	h, ok := fleetSvcByName(r.URL.Query().Get("host"))
	if !ok {
		http.Error(w, "unknown host", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baud.ToastOOB(baud.ToastProps{
		Tone:  "ok",
		Title: "pod killed: " + h.Svc + "-7f9c",
		Body:  "replicaset reschedules on the next healthy node",
	}).Render(r.Context(), w)
}

// ---- deploy modal ------------------------------------------------------------

// handleFleetDeploy serves GET /fleet/deploy — the toolbar's deploy-confirm
// Modal, hx-got into body/beforeend (overlay_api.go delivery pattern).
func handleFleetDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	FleetDeployModal().Render(r.Context(), w)
}

// handleFleetDeployRun serves GET /fleet/deploy/run — the modal's confirm
// action (hx-swap="none"): the OOB toast is the whole payload.
func handleFleetDeployRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baud.ToastOOB(baud.ToastProps{
		Tone:  "ok",
		Title: "deploy queued: ingest-gw v2.14.1",
		Body:  "rolling restart across 12 pods · no downtime expected",
	}).Render(r.Context(), w)
}

// ---- nav tree (lazy branch) ---------------------------------------------------

// fleetTreeNodes is the navigator fixture: regions → clusters → hosts,
// the SEV-1 host selected, eu-west-1's edge cluster lazy.
func fleetTreeNodes() []baud.TreeNode {
	return []baud.TreeNode{
		{Label: "us-east-1", Meta: "6 svc", Expanded: true, Children: []baud.TreeNode{
			{Label: "core", Expanded: true, Children: []baud.TreeNode{
				{Label: "auth-svc", Meta: "9"},
				{Label: "billing-api", Meta: "4"},
				{Label: "edge-cache", Meta: "16"},
			}},
			{Label: "data", Children: []baud.TreeNode{
				{Label: "pg-primary", Meta: "1"},
				{Label: "kafka-brk", Meta: "5"},
			}},
		}},
		{Label: "eu-west-1", Meta: "3 svc", Expanded: true, Children: []baud.TreeNode{
			{Label: "ingest", Expanded: true, Children: []baud.TreeNode{
				{Label: "ingest-gw", Meta: "9/12", Selected: true},
				{Label: "notif-fan", Meta: "3"},
				{Label: "webhook-out", Meta: "2"},
			}},
			{Label: "edge", LazyURL: "/fleet/tree?node=euw1/edge"},
		}},
		{Label: "us-west-2", Meta: "3 svc", Children: []baud.TreeNode{
			{Label: "search-idx", Meta: "6"},
			{Label: "search-query", Meta: "8"},
			{Label: "flag-svc", Meta: "3"},
		}},
		{Label: "ap-south-1", Meta: "1 svc", Children: []baud.TreeNode{
			{Label: "media-proc", Meta: "5"},
		}},
	}
}

// fleetTreeLazy fixes the lazy branch's server-side children. eu-west-1's
// edge branch is the last child of a root-level node, so its children sit
// under a three-space prefix (the server owns glyph continuity).
var fleetTreeLazy = map[string]treeFragment{
	"euw1/edge": {
		Prefix: "   ",
		Nodes: []baud.TreeNode{
			{Label: "edge-cache-ams", Meta: "warm"},
			{Label: "edge-cache-dub", Meta: "warm"},
			{Label: "edge-lb", Meta: "v2"},
		},
	},
}

// handleFleetTree serves GET /fleet/tree?node=… — the navigator's
// lazy-branch round-trip (tree_api.go pattern). Unknown nodes are 400.
func handleFleetTree(w http.ResponseWriter, r *http.Request) {
	frag, ok := fleetTreeLazy[r.URL.Query().Get("node")]
	if !ok {
		http.Error(w, "unknown node", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baud.TreeChildren(frag.Prefix, frag.Nodes).Render(r.Context(), w)
}

// ---- command palette ------------------------------------------------------------

// fleetPaletteCommands is the console's ⌘K command set. Hrefs come from
// Opts so the static Pages render keeps relative links; action rows carry
// opaque command ids — the topbar deploy/inspect buttons subscribe to
// baud:paletteCmd and replay their own click.
func fleetPaletteCommands(o Opts) []baud.PaletteCommand {
	return []baud.PaletteCommand{
		{Category: "fleet", Label: "deploy fleet to prod", Kbd: "d", Action: "fleet-deploy"},
		{Category: "fleet", Label: "inspect ingest-gw", Kbd: "i", Action: "fleet-inspect"},
		{Category: "go", Label: "go to component sheet", Kbd: "g s", Href: o.SheetHref},
		{Category: "go", Label: "go to fleet console", Kbd: "g f", Href: o.AppHref},
	}
}

// handleFleetPalette is the console palette's server-filter round-trip
// (palette_api.go pattern, console command set + root id). Arbitrary
// queries match nothing and render the ∅ empty state — never an error.
func handleFleetPalette(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	var hits []baud.PaletteCommand
	for _, c := range fleetPaletteCommands(ServerOpts()) {
		if q == "" || strings.Contains(strings.ToLower(c.Category+" "+c.Label), q) {
			hits = append(hits, c)
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	res := baud.PaletteResultsProps{ID: "fleet-palette", Commands: hits}
	if err := baud.PaletteResults(res).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

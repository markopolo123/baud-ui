package baud

import "strconv"

// Tab is one entry in a Tabs strip.
type Tab struct {
	Label string
	// Count renders an optional count badge after the label, reusing the
	// Badge primitive (compacted to tab scale by components/tabs.css).
	Count *int
	// Href is the htmx-mode pane URL: the tab issues hx-get={Href} into the
	// shared TabsProps.Target panel. Ignored in local mode.
	Href string
	// NavHref is the htmx-mode progressive-enhancement URL. When set the
	// tab renders as a real <a href={NavHref}> still carrying the hx-get
	// wiring: with scripting on, htmx intercepts the click and swaps the
	// pane; with scripting off, the browser navigates and the server
	// renders the page with that pane active — panes beyond the default
	// stay reachable (graceful degradation). Anchors activate on Enter
	// (not Space, unlike buttons). Ignored in local mode.
	NavHref string
	// Panel is the local-mode id of the pre-rendered TabPanel this tab
	// controls (aria-controls + hidden-attribute swap). Ignored in htmx mode.
	Panel string
}

func (t Tab) count() string {
	if t.Count == nil {
		return ""
	}
	return strconv.Itoa(*t.Count)
}

// TabsVariant selects the strip's visual variant.
type TabsVariant string

const (
	TabsUnderline TabsVariant = ""      // 2px accent underline on the active tab
	TabsBoxed     TabsVariant = "boxed" // bordered strip, active tab = accent fill
)

// TabsProps configures a Tabs strip (design/README.md "Structure"). The
// strip is a real ARIA tabs widget: role=tablist over <button role=tab>
// children with aria-selected and a roving tabindex; ←/→ move focus and
// ↵/Space activate (Tabs behavior in assets/baud._hs — buttons activate
// natively on Enter/Space, the behavior only roves focus and swaps state).
//
// Two switching modes:
//   - htmx (Target != ""): every tab hx-gets its Tab.Href pane into the one
//     shared tabpanel named by Target — server renders the pane.
//   - local (Target == ""): tabs point at pre-rendered TabPanel siblings via
//     Tab.Panel; the Tabs behavior swaps the hidden attribute locally.
type TabsProps struct {
	// ID names the strip; tab buttons derive stable ids <ID>-tab-<n>.
	ID      string
	Variant TabsVariant
	Tabs    []Tab
	// ActiveIndex marks the initially active tab (is-active +
	// aria-selected=true + tabindex=0).
	ActiveIndex int
	// Target is the id (no "#") of the shared tabpanel that every tab
	// hx-gets its pane into. Empty selects local mode.
	Target string
}

func (p TabsProps) variantClass() string {
	if p.Variant == TabsBoxed {
		return "tabs-boxed"
	}
	return "tabs-underline"
}

// htmx reports whether the strip swaps panes via server round-trips.
func (p TabsProps) htmx() bool { return p.Target != "" }

// controls resolves a tab's aria-controls id: the shared htmx target, or
// the tab's own pre-rendered panel.
func (p TabsProps) controls(t Tab) string {
	if p.htmx() {
		return p.Target
	}
	return t.Panel
}

func (p TabsProps) tabID(i int) string {
	if p.ID == "" {
		return ""
	}
	return p.ID + "-tab-" + strconv.Itoa(i)
}

func (p TabsProps) selected(i int) string {
	if i == p.ActiveIndex {
		return "true"
	}
	return "false"
}

// tabindex implements the roving tabindex: only the active tab sits in the
// document tab order; ←/→ move focus across the rest.
func (p TabsProps) tabindex(i int) string {
	if i == p.ActiveIndex {
		return "0"
	}
	return "-1"
}

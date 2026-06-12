package baud

import (
	"strconv"

	"github.com/a-h/templ"
)

// PanelStateKind selects which of the four panel-body states renders.
type PanelStateKind string

const (
	StateSkeleton PanelStateKind = "skeleton" // animated --bg-active bars
	StateLoading  PanelStateKind = "loading"  // spinner + cmd hint text
	StateEmpty    PanelStateKind = "empty"    // ∅ + title + sub + action
	StateError    PanelStateKind = "error"    // ✗ in --err + retry action
)

// PanelStateProps configures the PanelState component (design/README.md
// "Feedback"): the four states a panel body can sit in while it has no
// real content.
type PanelStateProps struct {
	Kind PanelStateKind
	// Title is the state headline (--fg-muted; --err under error).
	Title string
	// Sub is the supporting line — under loading it is the cmd hint
	// (e.g. "$ fleetctl get units --watch").
	Sub string
	// Action is the optional action slot: an empty-state CTA or the
	// error-state retry Btn. Nil renders no slot.
	Action templ.Component
}

// glyph is the state's mono glyph: ∅ for empty, ✗ for error; loading
// renders the Spinner instead and skeleton has no glyph at all.
func (p PanelStateProps) glyph() string {
	if p.Kind == StateError {
		return "✗"
	}
	return "∅"
}

// skelRows is the skeleton's bar layout: five rows of cell widths in ch
// (mono-honest), staggered organically like the design prototype's rows.
var skelRows = [][]int{
	{4, 15, 7, 25},
	{4, 12, 7, 30},
	{4, 18, 7, 20},
	{4, 14, 7, 28},
	{4, 11, 7, 23},
}

// skelWidth renders one bar's inline width (the only per-cell variation;
// everything else is components/feedback.css).
func skelWidth(ch int) string {
	return "width: " + strconv.Itoa(ch) + "ch"
}

package baud

import "strings"

// PanesProps configures a pane-tiling grid (the tmux-style workhorse).
type PanesProps struct {
	// Template is the grid template, sizes in ch (mono-honest) or fr,
	// e.g. "42ch 1fr". One track per pane.
	Template string
	// Rows tiles horizontally (data-panes-rows) instead of vertically.
	Rows bool
	// Resizable renders 7px drag gutters between panes (server-side —
	// no client-side DOM injection) and installs the Resizable behavior.
	Resizable bool
	// ID keys persisted pane sizes in localStorage (data-panes-id).
	ID string
	// Class adds extra classes to the wrapper.
	Class string
}

// templateAttr returns the data-panes value. When resizable, 7px gutter
// tracks are interleaved between the authored tracks so the grid template
// matches the server-rendered gutter elements. Note: tracks containing
// spaces (e.g. minmax(0, 1fr)) are not supported in resizable templates.
func (p PanesProps) templateAttr() string {
	if !p.Resizable {
		return p.Template
	}
	return strings.Join(strings.Fields(p.Template), " 7px ")
}

// installAttr is the hyperscript install list. Order matters: Panes must
// init before Resizable so a persisted template overrides the authored one.
func (p PanesProps) installAttr() string {
	if p.Resizable {
		return "install Panes install Resizable"
	}
	return "install Panes"
}

// rows flips a PanesProps to the data-panes-rows variant.
func rows(p PanesProps) PanesProps {
	p.Rows = true
	return p
}

func (p PanesProps) gutterOrientation() string {
	if p.Rows {
		return "horizontal"
	}
	return "vertical"
}

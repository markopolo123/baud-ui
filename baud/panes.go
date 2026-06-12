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
// matches the server-rendered gutter elements. Tracks are split at
// top-level whitespace only, so functional tracks containing spaces —
// e.g. minmax(0, 1fr) — stay intact as single tracks. Note: repeat(N, …)
// also counts as ONE track here; resizable templates must author one
// track per pane, so expand repeats by hand.
func (p PanesProps) templateAttr() string {
	if !p.Resizable {
		return p.Template
	}
	return strings.Join(splitTracks(p.Template), " 7px ")
}

// splitTracks tokenizes a grid template into tracks, splitting at
// whitespace outside parentheses so minmax(...)/repeat(...) survive whole.
func splitTracks(template string) []string {
	var tracks []string
	depth, start := 0, -1
	for i := 0; i < len(template); i++ {
		switch template[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ' ', '\t', '\n', '\r':
			if depth == 0 {
				if start >= 0 {
					tracks = append(tracks, template[start:i])
					start = -1
				}
				continue
			}
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tracks = append(tracks, template[start:])
	}
	return tracks
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

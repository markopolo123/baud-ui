package baud

import (
	"fmt"
	"strings"
)

// PaletteCommand is one command palette row: a faint UPPERCASE category
// column, the label, and right-aligned Kbd shortcut chips. Exactly one of
// Href/Action decides activation: Href rows are real anchors that
// navigate, Action rows are buttons carrying the command id as data-cmd.
type PaletteCommand struct {
	// Category fills the fixed-width faint UPPERCASE left column.
	Category string
	// Label is the command text.
	Label string
	// Kbd renders right-aligned shortcut chips, one per space-separated
	// token ("g s" → two chips); empty renders no chip group.
	Kbd string
	// Href makes the row an anchor navigating there on activation.
	Href string
	// Action is an OPAQUE COMMAND ID: it makes the row a button rendered
	// with data-cmd, and on activation the Palette behavior dispatches it
	// as a baud:paletteCmd event (detail.cmd) to <body>. The value is
	// never executed — safe for dynamic data; consumers subscribe to the
	// event and switch on the id. Ignored when Href is set.
	Action string
}

// PaletteProps configures the CommandPalette — the flagship htmx
// integration: PaletteKey (installed on <body>) sends baud:palette on
// ⌘K/Ctrl-K, the Palette behavior (assets/baud._hs) opens the overlay,
// traps focus in the query input, drives ↑↓/↵/Esc over the CURRENT row
// list (so it survives swaps) and restores focus on close. Typing fires
// a debounced hx-get that swaps the PaletteResults fragment into the
// listbox; ARIA follows the combobox/listbox pattern.
type PaletteProps struct {
	// ID is required — input (ID-input), listbox (ID-list) and row ids
	// (ID-cmd-n) that the ARIA wiring points at derive from it.
	ID string
	// SearchURL is the hx-get endpoint the debounced keyup round-trips
	// to; the query travels under Name. Empty renders no htmx wiring
	// (the initial Commands list is then fixed).
	SearchURL string
	// Name is the query parameter name; defaults to "q".
	Name string
	// Placeholder defaults to "type a command…".
	Placeholder string
	// Commands is the initial (unfiltered) command list.
	Commands []PaletteCommand
	// Static renders the palette open, inline and inert (no behavior,
	// no htmx) — the component-sheet styling variant.
	Static bool
	// Highlight pre-marks the 1-based nth row .hl (static styling
	// variant); 0 marks none. The live behavior owns .hl otherwise.
	Highlight int
}

func (p PaletteProps) inputID() string { return p.ID + "-input" }
func (p PaletteProps) listID() string  { return p.ID + "-list" }

func (p PaletteProps) name() string {
	if p.Name == "" {
		return "q"
	}
	return p.Name
}

func (p PaletteProps) placeholder() string {
	if p.Placeholder == "" {
		return "type a command…"
	}
	return p.Placeholder
}

// PaletteResultsProps configures the row-list fragment — the initial
// listbox content and the hx-get response body for server filtering.
type PaletteResultsProps struct {
	// ID is the palette root id the row ids derive from.
	ID string
	// Commands is the (already filtered, in server mode) command list;
	// empty renders the ∅ empty state.
	Commands []PaletteCommand
	// Highlight pre-marks the 1-based nth row .hl; 0 marks none.
	Highlight int
}

// paletteCmdID is the row id contract: <root id>-cmd-<index>. The
// Palette behavior points aria-activedescendant at these ids.
func paletteCmdID(root string, i int) string {
	return fmt.Sprintf("%s-cmd-%d", root, i)
}

// paletteKeys splits the Kbd spec into per-chip tokens.
func paletteKeys(kbd string) []string {
	return strings.Fields(kbd)
}

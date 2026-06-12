package baud

import "github.com/a-h/templ"

// PanelProps configures the Panel container (design/README.md "Structure").
type PanelProps struct {
	// Title renders in the header strip — UPPERCASE --fs-sm muted via the
	// type rules in CSS, so author it in lowercase. It also names the
	// section for assistive tech (aria-label).
	Title string
	// Actions is the optional right-aligned header slot: compose Btns,
	// Kbd chips, Badges, Selects…
	Actions templ.Component
	// ID names the panel — the natural hx-target for body swaps.
	ID string
	// Class adds composition classes to the root (e.g. "fill" in stacks).
	Class string
}

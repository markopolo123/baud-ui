package baud

// DrawerProps configures the Drawer overlay (design/README.md "Overlays").
type DrawerProps struct {
	// ID names the overlay root and derives the aria-labelledby title id —
	// required for correct dialog labelling.
	ID string
	// Title renders in the shared header strip (UPPERCASE via the type
	// rules in CSS — author it in lowercase) and labels the dialog.
	Title string
	// Static keeps the overlay in the DOM, hidden until .open — the
	// class-toggle variant for always-rendered pages. Open it with
	// `send baud:overlayOpen to #<id>`; htmx-injected drawers omit it.
	Static bool
}

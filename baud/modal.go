package baud

import "github.com/a-h/templ"

// ModalProps configures the Modal overlay (design/README.md "Overlays").
type ModalProps struct {
	// ID names the overlay root and derives the aria-labelledby title id —
	// required for correct dialog labelling.
	ID string
	// Title renders in the header strip (UPPERCASE via the type rules in
	// CSS — author it in lowercase) and labels the dialog.
	Title string
	// Footer is the optional right-aligned actions slot (.modal-ft).
	// Give dismissing actions (cancel…) the data-overlay-close attribute —
	// the Overlay behavior closes on any click that hits one.
	Footer templ.Component
	// Static keeps the overlay in the DOM, hidden until .open — the
	// class-toggle variant for always-rendered pages (the component sheet).
	// Open it with `send baud:overlayOpen to #<id>`. htmx-injected modals
	// (hx-target="body" hx-swap="beforeend") omit Static: they open on
	// arrival and are removed from the DOM on close.
	Static bool
}

// overlayTitleID derives the header title id the dialog's aria-labelledby
// points at; shared by Modal and Drawer.
func overlayTitleID(id string) string { return id + "-title" }

// overlayClasses builds the backdrop class list shared by Modal and Drawer.
func overlayClasses(variant string, static bool) string {
	cls := "overlay"
	if variant != "" {
		cls += " " + variant
	}
	if static {
		cls += " overlay-static"
	}
	return cls
}

package baud

// PopoverProps configures the anchored quick-actions popover: a trigger
// button (btn-styled) that toggles a 280px panel (--bg-panel, strong
// border, deep shadow) absolutely positioned under the trigger. Open and
// close are purely-local UI: inline hyperscript on the trigger toggles
// .open on the wrap and mirrors it onto aria-expanded; the shared
// MenuDismiss behavior (assets/baud._hs) drops .open on outside click /
// Esc — which also keeps popovers one-at-a-time, since any click on
// another trigger is an outside click for every other open wrap.
type PopoverProps struct {
	// ID is required — it lands on the wrap and derives the panel id
	// (ID-panel) that the trigger's aria-controls points at.
	ID string
	// Label is the trigger button text.
	Label string
	// Glyph is an optional mono glyph prefix on the trigger (e.g. "▾").
	Glyph string
	// Title renders an uppercase .field-label header inside the panel.
	Title string
	// Open server-renders the popover already open (wrap .open +
	// trigger aria-expanded=true).
	Open bool
	// Disabled disables the trigger button.
	Disabled bool
	// Class adds extra classes to the wrap.
	Class string
}

func (p PopoverProps) panelID() string { return p.ID + "-panel" }

// popAriaExpanded renders the trigger's aria-expanded literal.
func popAriaExpanded(open bool) string {
	if open {
		return "true"
	}
	return "false"
}

// popoverWrapHS is installed on every popover wrap: MenuDismiss owns
// dropping .open on outside click / Esc; the extra handlers only keep the
// trigger's aria-expanded in sync with it (same split as SelectKeys).
const popoverWrapHS = `install MenuDismiss
on keyup[key == 'Escape'] from window
  set ctl to me.querySelector('[aria-expanded]')
  if ctl is not null then call ctl.setAttribute('aria-expanded', 'false') end
end
on click from elsewhere
  set ctl to me.querySelector('[aria-expanded]')
  if ctl is not null then call ctl.setAttribute('aria-expanded', 'false') end
end`

// popoverTriggerHS toggles the panel and mirrors the open state onto
// aria-expanded — local UI, so hyperscript, inline because it is trivial.
const popoverTriggerHS = "on click set root to closest .pop-wrap then " +
	"toggle .open on root then " +
	"if root.matches('.open') then set @aria-expanded to 'true' " +
	"else set @aria-expanded to 'false'"

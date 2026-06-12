package baud

import "fmt"

// SelectOption is one option, shared by Select, SelectMenu and Combobox.
type SelectOption struct {
	Value string
	Label string // defaults to Value
	Meta  string // Combobox: right-aligned faint meta column
	// Disabled disables the option (native Select only — the custom menu
	// simply omits options that should not be offered).
	Disabled bool
}

func (o SelectOption) label() string {
	if o.Label == "" {
		return o.Value
	}
	return o.Label
}

// SelectProps configures the baseline Select: a REAL <select> element
// styled to match the trigger (appearance:none + overlaid ▼ chevron).
// This is the htmx-friendly default — it works with JS disabled and
// submits natively; SelectMenu is the progressive enhancement.
type SelectProps struct {
	// ID lands on the <select> for label association.
	ID string
	// Name is the form field name.
	Name string
	// Value marks the matching option selected.
	Value string
	// Options is the option list.
	Options []SelectOption
	// Disabled disables the whole control.
	Disabled bool
	// Width sizes the wrapper inline (e.g. "18ch"); empty = content width.
	Width string
	// Class adds extra classes to the wrapper.
	Class string
}

// SelectMenuProps configures the enhanced select: a custom trigger button
// (label + ▼/▲ chevron) and an absolutely-positioned listbox menu
// (--bg-panel, strong border, shadow). Open/close, keyboard (↑↓ ↵ esc)
// and picking are hyperscript behaviors (MenuDismiss + SelectKeys +
// SelectPick in assets/baud._hs); a hidden input keeps it form-capable.
type SelectMenuProps struct {
	// ID is required — the menu id (ID-menu) and option ids (ID-opt-n)
	// that aria-controls / aria-activedescendant point at derive from it.
	ID string
	// Name names the hidden form input; empty renders no hidden input.
	Name string
	// Value marks the matching option selected (is-active, ✓,
	// aria-selected) and puts its label on the trigger.
	Value string
	// Options is the option list.
	Options []SelectOption
	// Placeholder is the trigger text when no option matches Value;
	// defaults to "—".
	Placeholder string
	// AlignRight anchors the menu to the trigger's right edge.
	AlignRight bool
	// Disabled disables the trigger button.
	Disabled bool
	// Width sizes the wrapper inline (e.g. "18ch"); empty = content width.
	Width string
	// Class adds extra classes to the wrapper.
	Class string
}

func (p SelectMenuProps) menuID() string { return p.ID + "-menu" }

func (p SelectMenuProps) currentLabel() string {
	for _, o := range p.Options {
		if p.Value != "" && o.Value == p.Value {
			return o.label()
		}
	}
	if p.Placeholder != "" {
		return p.Placeholder
	}
	return "—"
}

// selAriaSelected renders the aria-selected literal for an option.
func selAriaSelected(selected bool) string {
	if selected {
		return "true"
	}
	return "false"
}

// selectOptionID is the option id contract: <root id>-opt-<index>. The
// SelectKeys behavior points aria-activedescendant at these ids.
func selectOptionID(root string, i int) string {
	return fmt.Sprintf("%s-opt-%d", root, i)
}

// selectMenuBehaviors is installed on every enhanced-select root:
// MenuDismiss (outside click / Esc), SelectKeys (↑↓ ↵ esc + ARIA sync),
// SelectPick (option click → value/label/aria update + close).
const selectMenuBehaviors = "install MenuDismiss install SelectKeys install SelectPick"

// selectTriggerHS toggles the menu and mirrors the open state onto
// aria-expanded — local UI, so hyperscript, inline because it is trivial.
const selectTriggerHS = "on click toggle .open on closest .select then " +
	"if closest .select matches .open then set @aria-expanded to 'true' " +
	"else set @aria-expanded to 'false'"

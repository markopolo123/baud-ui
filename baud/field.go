package baud

// InputProps configures the Input primitive: a real <input> inside a
// styled wrapper row (.input-wrap). Focus styling (accent border + 1px
// accent glow ring) is pure CSS via :focus-within — no client behaviour.
type InputProps struct {
	// ID is the input element id — pair it with FieldProps.For so the
	// field label is properly associated (a11y).
	ID string
	// Name is the form field name.
	Name string
	// Type is the input type; empty defaults to "text".
	Type string
	// Value pre-fills the input.
	Value string
	// Placeholder renders in --fg-faint.
	Placeholder string
	// Prefix is an optional affix rendered before the input (--fg-faint).
	Prefix string
	// Suffix is an optional affix rendered after the input (--fg-faint).
	Suffix string
	// Error sets the wrapper border to --err and aria-invalid on the
	// input. Pair with FieldProps.Error for the ✗ hint message.
	Error bool
	// Disabled disables the real input; the wrapper dims via CSS.
	Disabled bool
	// Class adds extra classes to the wrapper.
	Class string
}

func (p InputProps) typeAttr() string {
	if p.Type == "" {
		return "text"
	}
	return p.Type
}

// FieldProps configures the Field primitive: UPPERCASE label + control
// slot (templ children) + hint line.
type FieldProps struct {
	// Label renders UPPERCASE at --fs-sm. Empty omits the label.
	Label string
	// For associates the label with the control's id (a11y).
	For string
	// Hint renders on the hint line in --fg-faint.
	Hint string
	// Error, when non-empty, replaces the hint with "✗ message" in --err.
	// Set Error on the wrapped Input too so its border matches.
	Error string
	// Class adds extra classes to the field wrapper.
	Class string
}

package baud

// ChoiceProps configures a Checkbox or Radio. A real <input> sits underneath
// for forms and a11y (visually hidden but focusable); the text glyphs
// ([x] [ ] / (•) ( )) are pure-CSS presentation driven by :checked. The
// label element wraps the input, so the whole control is clickable with no
// for/id wiring and no client logic.
type ChoiceProps struct {
	Name     string
	Value    string // omitted when empty (checkbox then submits the native "on")
	Label    string
	ID       string // optional: lands on the input for external hooks
	Checked  bool
	Disabled bool
}

// ToggleOption is one segment of a segmented Toggle.
type ToggleOption struct {
	Value    string
	Label    string // defaults to Value
	Disabled bool   // disables the segment's underlying radio
}

func (o ToggleOption) label() string {
	if o.Label == "" {
		return o.Value
	}
	return o.Label
}

// ToggleProps configures a segmented Toggle: a bordered strip whose selected
// segment fills accent. Underneath it is a real radio group sharing Name, so
// arrow keys move the selection natively and the value submits with forms.
type ToggleProps struct {
	Name    string
	Value   string // the selected option's Value; empty selects nothing
	Options []ToggleOption
}

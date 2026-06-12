package baud

import "strings"

// ComboboxProps configures the Combobox: an input with a ⌕ affix whose
// menu filters as you type. Two modes:
//
//   - client (default): the ComboboxFilter hyperscript behavior shows/
//     hides the server-rendered options locally and wraps the matched
//     label substring in an accent-bold .cb-match span;
//   - server (SearchURL set): a debounced hx-get round-trip swaps the
//     ComboboxOptions fragment into the menu — the query travels under
//     the input's Name, the highlight is rendered server-side.
//
// Keyboard (↑↓ ↵ esc) and dismissal are the same SelectKeys/SelectPick/
// MenuDismiss behaviors the enhanced select uses; ARIA follows the
// combobox pattern (role=combobox, aria-expanded/-controls/-autocomplete,
// aria-activedescendant while navigating). With JS disabled it degrades
// to a plain named input.
type ComboboxProps struct {
	// ID is required — menu and option ids derive from it.
	ID string
	// Name names the input; in server mode it is also the query param.
	Name string
	// Value pre-selects the matching option and fills the input with its
	// label.
	Value string
	// Options is the option list (the full set in client mode, the
	// initial menu in server mode).
	Options []SelectOption
	// Placeholder defaults to "search…".
	Placeholder string
	// SearchURL switches to server mode: a debounced hx-get to this URL
	// swaps the returned ComboboxOptions fragment into the menu.
	SearchURL string
	// Width sizes the wrapper inline (e.g. "26ch"); empty = content width.
	Width string
	// Class adds extra classes to the wrapper.
	Class string
}

func (p ComboboxProps) menuID() string { return p.ID + "-menu" }

func (p ComboboxProps) placeholder() string {
	if p.Placeholder == "" {
		return "search…"
	}
	return p.Placeholder
}

func (p ComboboxProps) currentLabel() string {
	for _, o := range p.Options {
		if o.Value == p.Value {
			return o.label()
		}
	}
	return p.Value
}

// behaviors: the shared menu behaviors, plus the local filter only in
// client mode (server mode filters via the hx-get round-trip instead).
func (p ComboboxProps) behaviors() string {
	if p.SearchURL != "" {
		return selectMenuBehaviors
	}
	return selectMenuBehaviors + " install ComboboxFilter"
}

// ComboboxOptionsProps configures the option-list fragment — the initial
// menu content and the hx-get response body for server-mode filtering.
type ComboboxOptionsProps struct {
	// ID is the combobox root id the option ids derive from.
	ID string
	// Query highlights its first case-insensitive label match per option.
	Query string
	// Value marks the matching option selected.
	Value string
	// Options is the (already filtered, in server mode) option list.
	Options []SelectOption
}

// comboMatchParts is a label split around the first match of the query.
type comboMatchParts struct {
	Pre, Hit, Post string
	OK             bool
}

// comboMatch locates the first case-insensitive occurrence of q in label
// so the template can wrap it in the accent-bold .cb-match span.
func comboMatch(label, q string) comboMatchParts {
	q = strings.TrimSpace(q)
	if q == "" {
		return comboMatchParts{}
	}
	i := strings.Index(strings.ToLower(label), strings.ToLower(q))
	if i < 0 {
		return comboMatchParts{}
	}
	return comboMatchParts{
		Pre:  label[:i],
		Hit:  label[i : i+len(q)],
		Post: label[i+len(q):],
		OK:   true,
	}
}

// comboInputOpenHS opens the menu when the input gains focus or is
// clicked, mirroring the state onto aria-expanded.
const comboInputOpenHS = "on focus add .open to closest .select then set @aria-expanded to 'true' end " +
	"on click add .open to closest .select then set @aria-expanded to 'true' end"

package baud

// KV is one key/value row of a DefList. Plain-string values render
// tabular-nums; for rich values (badges, links, …) append DefItem rows
// through the DefList children slot instead.
type KV struct {
	Key   string
	Value string
}

// DefListProps configures the definition list (see components/lists.css):
// a semantic <dl> on a max-content/1fr grid — UPPERCASE --fg-faint keys,
// tabular-nums values.
type DefListProps struct {
	// Rows render in order as dt/dd pairs, before any children-slot rows.
	Rows []KV
	// Lines draws a hairline rule under every row (55% --border mix).
	Lines bool
}

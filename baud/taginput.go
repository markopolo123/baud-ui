package baud

// TagInputProps configures a TagInput: chips of "key=value" (or bare value)
// tags, each backed by a real <input type=hidden name=Name> so the tags
// submit with forms — removing a chip removes its hidden input. The visible
// text input is nameless so it never pollutes the submission. Suggestions
// render server-side into a menu the TagInput hyperscript behavior filters
// as you type; Esc/outside-click dismissal is the shared MenuDismiss.
type TagInputProps struct {
	Name        string   // form field name shared by every chip's hidden input
	Values      []string // initial chips: "key=value" or bare "value"
	Suggestions []string // suggestion menu entries, filtered client-side
	Placeholder string   // text input placeholder
	ID          string   // optional: lands on the text input (Field For pairing)
}

// tagKey returns the "key=" prefix of a key=value tag, or "" for bare values.
func tagKey(tag string) string {
	for i := 0; i < len(tag); i++ {
		if tag[i] == '=' {
			return tag[:i+1]
		}
	}
	return ""
}

// tagVal returns the value part of a key=value tag (the whole tag when bare).
func tagVal(tag string) string {
	return tag[len(tagKey(tag)):]
}

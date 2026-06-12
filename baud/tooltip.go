package baud

import "github.com/a-h/templ"

// Tooltips are an attribute, not an element: any host carrying data-tip
// grows a pure-CSS tip (::after renders attr(data-tip) with white-space:
// pre, so multi-line mono tips stay aligned; ::before is the arrow) above
// itself after a 150ms delay, on hover AND on :focus-visible — no
// hyperscript, no JS. Spread the returned attributes onto the host:
//
//	<span { baud.Tip("works on any element")... }>badge</span>
//
// Hosts that already declare a class attribute must not also spread a
// class-carrying helper (HTML keeps only the first class attribute), so
// Tip stays class-free; TipUnder adds the .tip-under variant class for
// plain prose hosts — dotted underline + help cursor.

// Tip returns the data-tip attribute for a tooltip host. Newlines in
// text render as aligned multi-line tips (white-space: pre).
func Tip(text string) templ.Attributes {
	return templ.Attributes{"data-tip": text}
}

// TipUnder is Tip plus the .tip-under variant class: dotted underline and
// a help cursor mark the host as explained-on-hover prose.
func TipUnder(text string) templ.Attributes {
	return templ.Attributes{"data-tip": text, "class": "tip-under"}
}

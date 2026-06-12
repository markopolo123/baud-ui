package baud

import "strconv"

// ToastDefaultMs is the ~4s default auto-dismiss interval (design spec:
// "auto-dismiss ~4s"), overridable per toast via DismissMs.
const ToastDefaultMs = 4000

// ToastProps configures one toast notification (see components/toasts.css
// and the Toast behavior in assets/baud._hs).
type ToastProps struct {
	// Tone: ok | err | warn | info (default info). Drives the 3px left
	// bar + glyph colour and picks the glyph (✓ ✗ ▲ ℹ — text glyphs,
	// not emoji).
	Tone string
	// Title is the first line of the message.
	Title string
	// Body is the optional faint second line.
	Body string
	// DismissMs overrides the auto-dismiss interval in milliseconds
	// (0 = ToastDefaultMs). Tests use a short interval; humans get ~4s.
	DismissMs int
	// Sticky disables auto-dismiss entirely (data-toast-ms="0") — for
	// errors that must be acknowledged, and for the static component
	// sheet. The ✕ button still dismisses.
	Sticky bool
}

func (p ToastProps) toneClass() string {
	return "tone-" + or(p.Tone, "info")
}

// glyph maps tones to their text glyphs. The fallback is the info glyph —
// tone defaults to info throughout.
func (p ToastProps) glyph() string {
	switch p.Tone {
	case "ok":
		return "✓"
	case "err":
		return "✗"
	case "warn":
		return "▲"
	default:
		return "ℹ"
	}
}

// dismissMs renders the data-toast-ms attribute the Toast behavior reads:
// "0" = sticky, otherwise the interval in milliseconds.
func (p ToastProps) dismissMs() string {
	if p.Sticky {
		return "0"
	}
	if p.DismissMs > 0 {
		return strconv.Itoa(p.DismissMs)
	}
	return strconv.Itoa(ToastDefaultMs)
}

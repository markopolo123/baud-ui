package baud

// BadgeProps configures a status chip (see components/badges.css).
type BadgeProps struct {
	// Tone: ok | warn | err | info | accent | neutral (default neutral).
	Tone string
	// Variant: tint | solid | outline (default tint). Tint backgrounds are
	// 15% color-mix of the tone var so new themes only supply base vars.
	Variant string
	// Dot renders a 5px square dot (currentColor) before the label.
	Dot bool
}

func (p BadgeProps) toneClass() string {
	return "tone-" + or(p.Tone, "neutral")
}

func (p BadgeProps) variantClass() string {
	switch p.Variant {
	case "solid":
		return "bd-solid"
	case "outline":
		return "bd-outline"
	default:
		return "bd-tint"
	}
}

// DotProps configures the 7px round status dot — the one sanctioned circle
// in the library.
type DotProps struct {
	// Tone: ok | warn | err | info | accent | neutral (default ok).
	Tone string
	// Pulse adds a box-shadow ping, gated behind prefers-reduced-motion.
	Pulse bool
}

func (p DotProps) toneClass() string {
	return "tone-" + or(p.Tone, "ok")
}

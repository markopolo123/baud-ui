package baud

import (
	"math"
	"strconv"
	"strings"
)

// Progress glyphs and defaults (design/README.md "Feedback"): an ASCII
// ▰▰▰▱▱ bar, 22 glyphs wide unless configured otherwise.
const (
	progFilledGlyph = "▰"
	progRestGlyph   = "▱"
	progDefaultBar  = 22
)

// ProgressProps configures the server-rendered ASCII progress bar. The bar
// string is computed in Go — no client behaviour at all.
type ProgressProps struct {
	// Value is the current progress, on the 0..Max scale (clamped).
	Value float64
	// Max is the full-bar value (default 100).
	Max float64
	// Label is the optional left label (--fg-muted, ellipsized).
	Label string
	// Chars is the bar width in glyphs (default 22).
	Chars int
	// Tone forces the bar tone (accent | warn | ok | err | info). Empty
	// selects automatically: accent normally, warn past 85%, ok at 100%.
	Tone string
}

func (p ProgressProps) max() float64 {
	if p.Max <= 0 {
		return 100
	}
	return p.Max
}

func (p ProgressProps) chars() int {
	if p.Chars <= 0 {
		return progDefaultBar
	}
	return p.Chars
}

// ratio is the clamped 0..1 completion fraction.
func (p ProgressProps) ratio() float64 {
	r := p.Value / p.max()
	return math.Min(1, math.Max(0, r))
}

// fillCount is how many of the chars() glyphs render filled (rounded, so a
// sliver of progress can still show an empty bar and 99% can fill it —
// the percent text carries the exact figure).
func (p ProgressProps) fillCount() int {
	return int(math.Round(p.ratio() * float64(p.chars())))
}

// Percent is the rounded 0–100 figure shown next to the bar and exposed as
// aria-valuenow.
func (p ProgressProps) Percent() int {
	return int(math.Round(p.ratio() * 100))
}

func (p ProgressProps) percentText() string {
	return strconv.Itoa(p.Percent()) + "%"
}

func (p ProgressProps) filled() string {
	return strings.Repeat(progFilledGlyph, p.fillCount())
}

func (p ProgressProps) rest() string {
	return strings.Repeat(progRestGlyph, p.chars()-p.fillCount())
}

// tone resolves the bar tone: forced via Tone, else the auto thresholds on
// the exact ratio — ok at 100%, warn strictly past 85%, accent otherwise.
func (p ProgressProps) tone() string {
	if p.Tone != "" {
		return p.Tone
	}
	r := p.ratio()
	switch {
	case r >= 1:
		return "ok"
	case r > 0.85:
		return "warn"
	default:
		return "accent"
	}
}

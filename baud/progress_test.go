package baud

import "testing"

// TestProgressFillMath pins the server-side bar computation: fill rounding,
// percent rounding, clamping, the auto-tone boundaries (85 stays accent,
// 86 flips warn, 100 flips ok) and the forced-tone override.
func TestProgressFillMath(t *testing.T) {
	cases := []struct {
		name    string
		p       ProgressProps
		fill    int
		percent int
		tone    string
	}{
		{"zero", ProgressProps{Value: 0}, 0, 0, "accent"},
		{"half", ProgressProps{Value: 50}, 11, 50, "accent"},
		{"boundary 85 stays accent", ProgressProps{Value: 85}, 19, 85, "accent"},
		{"boundary 86 flips warn", ProgressProps{Value: 86}, 19, 86, "warn"},
		{"99 still warn, bar rounds full", ProgressProps{Value: 99}, 22, 99, "warn"},
		{"100 flips ok", ProgressProps{Value: 100}, 22, 100, "ok"},
		{"fractional ratio past .85 is warn", ProgressProps{Value: 85.4}, 19, 85, "warn"},
		{"clamp above max", ProgressProps{Value: 150}, 22, 100, "ok"},
		{"clamp below zero", ProgressProps{Value: -10}, 0, 0, "accent"},
		{"rounds down", ProgressProps{Value: 2}, 0, 2, "accent"},     // 0.02*22 = 0.44
		{"rounds up", ProgressProps{Value: 3}, 1, 3, "accent"},       // 0.03*22 = 0.66
		{"thirds", ProgressProps{Value: 1, Max: 3}, 7, 33, "accent"}, // 7.33…
		{"custom max", ProgressProps{Value: 100, Max: 200}, 11, 50, "accent"},
		{"custom chars", ProgressProps{Value: 50, Chars: 10}, 5, 50, "accent"},
		{"custom chars full", ProgressProps{Value: 100, Chars: 5}, 5, 100, "ok"},
		{"zero max defaults to 100", ProgressProps{Value: 50, Max: 0}, 11, 50, "accent"},
		{"forced tone wins at 100", ProgressProps{Value: 100, Tone: "err"}, 22, 100, "err"},
		{"forced tone wins at 0", ProgressProps{Value: 0, Tone: "ok"}, 0, 0, "ok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.fillCount(); got != tc.fill {
				t.Errorf("fillCount = %d, want %d", got, tc.fill)
			}
			if got := tc.p.Percent(); got != tc.percent {
				t.Errorf("Percent = %d, want %d", got, tc.percent)
			}
			if got := tc.p.tone(); got != tc.tone {
				t.Errorf("tone = %q, want %q", got, tc.tone)
			}
		})
	}
}

// TestProgressBarStrings pins the rendered glyph strings: filled+rest always
// span exactly chars() glyphs of the right kind.
func TestProgressBarStrings(t *testing.T) {
	p := ProgressProps{Value: 50, Chars: 10}
	if got, want := p.filled(), "▰▰▰▰▰"; got != want {
		t.Errorf("filled = %q, want %q", got, want)
	}
	if got, want := p.rest(), "▱▱▱▱▱"; got != want {
		t.Errorf("rest = %q, want %q", got, want)
	}
	for _, v := range []float64{0, 2, 33.3, 85, 86, 99, 100, 150, -1} {
		p := ProgressProps{Value: v}
		if filled, rest := len([]rune(p.filled())), len([]rune(p.rest())); filled+rest != progDefaultBar {
			t.Errorf("value %v: filled %d + rest %d != %d glyphs", v, filled, rest, progDefaultBar)
		}
	}
}

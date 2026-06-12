package demo

import (
	"fmt"
	"strings"
)

// tweakGroup is one radio-style row of the tweaks panel: clicking a button
// removes every class in the group from <body> and adds the chosen one —
// root class swap only, exactly how consumers switch modes.
type tweakGroup struct {
	Label   string
	Class   string // marker class grouping the buttons, e.g. tw-theme
	Options []tweakOption
}

type tweakOption struct {
	Value string // the root class, e.g. t-mocha
	Label string
}

func tweakGroups() []tweakGroup {
	return []tweakGroup{
		{"theme", "tw-theme", []tweakOption{
			{"t-gruvbox", "gruvbox"}, {"t-mocha", "mocha"}, {"t-sollight", "sol light"},
		}},
		{"density", "tw-density", []tweakOption{
			{"d-ultra", "ultra"}, {"d-dense", "dense"}, {"d-cozy", "cozy"},
		}},
		{"borders", "tw-border", []tweakOption{
			{"b-line", "line"}, {"b-shade", "shade"}, {"b-ascii", "ascii"},
		}},
		{"type", "tw-type", []tweakOption{
			{"f-mono", "mono"}, {"f-mix", "mix"},
		}},
	}
}

// isDefault marks the buttons that match the Page default root classes.
func isDefault(class string) bool {
	switch class {
	case "t-gruvbox", "d-dense", "b-line", "f-mono":
		return true
	}
	return false
}

// tweakHS builds the inline hyperscript for one tweak button: swap the
// group's class on <body>, then move the active marker to this button.
func tweakHS(g tweakGroup, value string) string {
	var b strings.Builder
	b.WriteString("on click")
	for _, o := range g.Options {
		fmt.Fprintf(&b, " remove .%s from document.body then", o.Value)
	}
	fmt.Fprintf(&b, " add .%s to document.body then take .is-active from .%s for me", value, g.Class)
	return b.String()
}

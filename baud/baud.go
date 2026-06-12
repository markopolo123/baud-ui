// Package baud is a dense terminal-aesthetic component library:
// templ components, htmx for server round-trips, _hyperscript behaviors
// (assets/baud._hs) for purely-local UI. No JavaScript anywhere.
package baud

import (
	"strings"

	"github.com/a-h/templ"
)

// Pinned peer-dependency CDN URLs. htmx and _hyperscript are peer
// assumptions, not bundled; the baud._hs behaviors file MUST be loaded
// before the _hyperscript library so remotely-loaded behaviors are
// defined when elements install them.
const (
	HTMXSrc        = "https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js"
	HyperscriptSrc = "https://unpkg.com/hyperscript.org@0.9.14/dist/_hyperscript.min.js"
	FontsHref      = "https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap"
)

// PageProps configures the document wrapper. Zero value = defaults:
// t-gruvbox d-dense b-line f-mono, assets served at "assets/…"-relative
// hrefs (override CSSHref/HSHref for other mount points).
type PageProps struct {
	Title   string
	Theme   string // t-gruvbox | t-mocha | t-sollight
	Density string // d-ultra | d-dense | d-cozy
	Border  string // b-line | b-shade | b-ascii
	Type    string // f-mono | f-mix
	CSSHref string // default "assets/baud.css"
	HSHref  string // default "assets/baud._hs"
}

func (p PageProps) title() string {
	if p.Title == "" {
		return "baud/ui"
	}
	return p.Title
}

func (p PageProps) cssHref() string {
	if p.CSSHref == "" {
		return "assets/baud.css"
	}
	return p.CSSHref
}

func (p PageProps) hsHref() string {
	if p.HSHref == "" {
		return "assets/baud._hs"
	}
	return p.HSHref
}

func or(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// RootClasses resolves the theme/density/border/type mode classes with
// defaults applied. Mode switching is a root class swap — nothing else.
func (p PageProps) RootClasses() string {
	return strings.Join([]string{
		or(p.Theme, "t-gruvbox"),
		or(p.Density, "d-dense"),
		or(p.Border, "b-line"),
		or(p.Type, "f-mono"),
	}, " ")
}

// ShellProps configures the app frame. Nav is optional: nil renders the
// single-column variant. TopBar/StatusBar components fill the framing
// slots; main content is the templ children block.
type ShellProps struct {
	TopBar    templ.Component
	Nav       templ.Component
	StatusBar templ.Component
}

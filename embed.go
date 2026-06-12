// Package baudui embeds the built assets so the demo ships as one binary.
// dist/baud.css is produced by `just css` — run it before building.
package baudui

import _ "embed"

// CSS is the concatenated @layer bundle (dist/baud.css).
//
//go:embed dist/baud.css
var CSS []byte

// Behaviors is the hyperscript behaviors file — the only client logic in
// the project. It must be served and loaded BEFORE the _hyperscript library.
//
//go:embed assets/baud._hs
var Behaviors []byte

// Package baudui exposes the library's client assets, embedded from the
// COMMITTED sources so `go get` consumers build with no pre-step. The CSS
// bundle is concatenated in Go at first use — byte-identical to the
// dist/baud.css file the `just css` recipe emits (which remains available
// as a convenience artifact for consumers who want a file on disk).
package baudui

import (
	"bytes"
	"embed"
	"sync"
)

// Layered CSS sources. Concatenation order matters and mirrors `just css`:
// tokens, base, components/* sorted by filename, utilities.
//
//go:embed assets/css/tokens.css assets/css/base.css assets/css/components/*.css assets/css/utilities.css
var cssSources embed.FS

// behaviors is the hyperscript behaviors file — the only client logic in
// the project. It must be served and loaded BEFORE the _hyperscript library,
// and the ParseHealth behavior must stay last in the file.
//
//go:embed assets/baud._hs
var behaviors []byte

var cssBundle = sync.OnceValue(buildCSS)

// CSS returns the concatenated @layer bundle (tokens, base, components/*
// sorted by filename, utilities) — byte-identical to `just css`'s
// dist/baud.css. The bundle is built once and cached; callers must not
// mutate the returned slice.
func CSS() []byte { return cssBundle() }

// HS returns the hyperscript behaviors file (assets/baud._hs). Serve it
// before the _hyperscript library tag; callers must not mutate the
// returned slice.
func HS() []byte { return behaviors }

func buildCSS() []byte {
	var buf bytes.Buffer
	write := func(path string) {
		b, err := cssSources.ReadFile(path)
		if err != nil {
			panic("baudui: embedded CSS source missing: " + path + ": " + err.Error())
		}
		buf.Write(b)
	}
	write("assets/css/tokens.css")
	write("assets/css/base.css")
	entries, err := cssSources.ReadDir("assets/css/components")
	if err != nil {
		panic("baudui: embedded components dir missing: " + err.Error())
	}
	for _, e := range entries { // ReadDir returns entries sorted by filename
		write("assets/css/components/" + e.Name())
	}
	write("assets/css/utilities.css")
	return buf.Bytes()
}

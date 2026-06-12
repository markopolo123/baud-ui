package demo

// Opts wires page-level hrefs for the two environments: the live demo
// server (rooted routes) and the static GitHub Pages render (relative
// paths that survive the /baud-ui/ subpath).
type Opts struct {
	CSSHref   string
	HSHref    string
	AppHref   string
	SheetHref string
}

// ServerOpts are the hrefs used by the demo HTTP server.
func ServerOpts() Opts {
	return Opts{
		CSSHref:   "/assets/baud.css",
		HSHref:    "/assets/baud._hs",
		AppHref:   "/",
		SheetHref: "/sheet",
	}
}

// StaticOpts are the relative hrefs used by the static site render, where
// the sheet is index.html and the fleetctl placeholder is app.html.
func StaticOpts() Opts {
	return Opts{
		CSSHref:   "baud.css",
		HSHref:    "baud._hs",
		AppHref:   "app.html",
		SheetHref: "index.html",
	}
}

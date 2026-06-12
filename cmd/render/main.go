// Command render writes the static showcase to dist/site for GitHub Pages.
// All asset hrefs are relative so the site works under the /baud-ui/
// subpath; htmx and _hyperscript stay on their pinned CDNs.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/a-h/templ"

	baudui "github.com/markopolo123/baud-ui"
	"github.com/markopolo123/baud-ui/demo"
)

func main() {
	out := flag.String("out", "dist/site", "output directory")
	flag.Parse()

	if err := render(*out); err != nil {
		log.Fatal(err)
	}
	log.Printf("static site rendered to %s", *out)
}

func render(out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	o := demo.StaticOpts()
	pages := map[string]templ.Component{
		"index.html": demo.SheetPage(o), // the sheet is the landing page
		"app.html":   demo.AppPage(o),   // fleetctl placeholder
	}
	for name, page := range pages {
		f, err := os.Create(filepath.Join(out, name))
		if err != nil {
			return err
		}
		if err := page.Render(context.Background(), f); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	// Copy the embedded assets next to the pages (relative hrefs).
	if err := os.WriteFile(filepath.Join(out, "baud.css"), baudui.CSS, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "baud._hs"), baudui.Behaviors, 0o644)
}

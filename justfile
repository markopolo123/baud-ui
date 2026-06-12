# baud/ui — the only sanctioned entry point for build/test/lint/run.

set shell := ["bash", "-cu"]

default: check

# Regenerate *_templ.go from .templ sources (gitignored; CI regenerates).
generate:
    go tool templ generate

# Concatenate layered CSS sources into the single bundle file.
# Convenience artifact only — the root package embeds the sources and
# concatenates in Go (baudui.CSS()), so nothing below depends on this.
# Layer order is declared at the top of tokens.css; component files
# are globbed sorted so parallel waves can add components/<area>.css.
css:
    mkdir -p dist
    cat assets/css/tokens.css assets/css/base.css $(ls assets/css/components/*.css | sort) assets/css/utilities.css > dist/baud.css

build: generate
    go build ./...

run: generate
    go run ./cmd/demo

# Render the static showcase for GitHub Pages to dist/site.
render: generate
    go run ./cmd/render

lint: generate
    test -z "$(gofmt -l $(git ls-files -cmo --exclude-standard '*.go'))"
    go vet ./...

# godog (BDD) + unit tests.
test: generate
    go test ./...

# Real-browser tests (build tag e2e). Needs `just install-browsers` once.
e2e: generate
    go test -tags e2e -count=1 ./e2e/...

install-browsers:
    go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps chromium

check: lint build test e2e

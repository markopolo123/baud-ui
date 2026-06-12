# baud/ui — the only sanctioned entry point for build/test/lint/run.

set shell := ["bash", "-cu"]

# Concatenate layered CSS sources into the single bundle.
# Layer order is declared at the top of tokens.css; component files
# are globbed sorted so parallel waves can add components/<area>.css.
css:
    mkdir -p dist
    cat assets/css/tokens.css assets/css/base.css $(ls assets/css/components/*.css | sort) assets/css/utilities.css > dist/baud.css

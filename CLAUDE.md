# baud/ui — dense terminal component library

Server-rendered component library: Go `templ` + `htmx` + `_hyperscript`, zero Node
toolchain. The authoritative design spec is `design/README.md` (high-fidelity handoff —
component inventory and attribute API are the contract). Modeled on oat.ink: semantic
HTML, `data-*` attribute components, token-driven CSS, no class soup.

## Stack (adapted from ~/.claude/specs/go-web-stack.md)

- Go, single module `github.com/markopolo123/baud-ui`. Stdlib-first, minimal deps.
- This is a **library + demo app**, not a service. No database. Demo state is
  in-memory fixtures. The PocketBase/ConnectRPC persistence rules of the global spec
  do not apply here — recorded once, do not revisit.
- UI: `templ` for components, `htmx` for server round-trips (sort/filter/pagination/
  palette), `_hyperscript` for purely-local behaviour (menu toggle, tab switch, copy).
- **No JavaScript, full stop.** Client behaviour ships as _hyperscript behaviors in
  `assets/baud._hs`, loaded via `<script type="text/hyperscript" src=…>` placed BEFORE
  the _hyperscript library tag (remotely-loaded behaviors must be defined before
  hyperscript boots). Components opt in via `_="install BehaviorName"` attributes.
  NO `.js` files or inline `<script>` logic anywhere — PRs adding them are rejected.
- Dev deps via `go tool` (templ, etc.). Tasks via `just` — the only sanctioned entry
  point for build/test/lint/run/e2e. A command run twice belongs in the justfile.
- Assets embedded (`embed`) — demo builds to one binary.

## Layout

- `baud/` — the templ component library
- `assets/css/` — layered CSS sources (`tokens.css`, `base.css`, `components/*.css`,
  `utilities.css`); `just css` concatenates → `dist/baud.css`. Parallel agents own
  distinct files — never edit a CSS file another in-flight branch owns.
- `assets/baud._hs` — the hyperscript behaviors file (Panes, Resizable, MenuDismiss,
  PaletteKey, …) — the only client logic in the project
- `demo/` — fleetctl demo console + component sheet (importable package; the sheet
  registry lives in `demo/registry.go`, sections in `demo/sheet_<component>.templ`)
- `cmd/demo/` — thin HTTP server wrapper around `demo/`
- `cmd/render/` — static renderer: writes component sheet (all themes × densities) to
  `dist/site/` for GitHub Pages
- `features/` — godog `.feature` files
- `e2e/` — go-playwright tests
- `design/` — design handoff reference. Lift the CSS nearly verbatim; **never port the
  JSX** — it is throwaway prototype scaffolding.

## Testing — non-negotiable

1. **BDD first.** Every component gets a `.feature` file under `features/` BEFORE
   implementation code. Red first, then green under godog.
2. **Scenarios must cover the options matrix** — every prop/variant/tone/state
   combination that changes rendering. E.g. Btn: each Variant scenario asserts the
   rendered markup carries the right variant attribute/tokens; Badge: Tone × Variant
   grid; Field: error state flips hint + input border. godog steps assert on rendered
   templ output (parsed HTML), not string-contains.
3. **go-playwright for every component**: real-browser assertions of option visuals via
   computed styles (e.g. danger Btn border resolves to the theme's `--err`; hover fills
   err), keyboard operability (`Tab`, `↑↓ ↵ esc`), focus-visible, overlay dismiss
   (Esc + outside click), and theme/density switching by root-class swap only.
4. `just check` = lint + build + godog + playwright (each step runs `templ generate`
   and the CSS bundle first; `*_templ.go` is gitignored, so there is no
   generate-drift check). Green or it doesn't merge.

## Hard rules (from the design spec)

- **Zero border-radius. No emoji. No media queries** except `prefers-reduced-motion`.
- **No raw hex/px inside component CSS** — `var()` tokens only. Themes/density/border/
  type switch via root class swap, nothing else.
- CSS in `@layer theme, base, components, utilities`.
- Semantic HTML, correct ARIA roles, every interactive component keyboard-operable.
- Desktop only: `min-width: 1240px`, no mobile layout.
- No SPA framework. No JavaScript at all — hyperscript behaviors only.

## Git workflow — enforced

- **Never commit to `main`.** All work happens in git worktrees on `feat/<name>`
  branches and lands via PR.
- PRs merge only after gatekeeper review (see `docs/WAYS_OF_WORKING.md`).
- Branch naming: `feat/<component>`, `fix/<thing>`, `chore/<thing>`.

## Definition of Done (per component)

- [ ] `.feature` scenarios green under godog, covering the full options matrix
- [ ] go-playwright green: interactions, keyboard, per-option computed-style assertions
- [ ] Rendered on the component-sheet page (so it ships in the Pages showcase)
- [ ] Token-only CSS; correct under all 3 themes × 3 densities × border modes
- [ ] `just check` green in CI
- [ ] PR approved by the arbiter, merged; Pages deploy green

# baud/ui

Dense terminal-aesthetic component library for server-rendered Go apps — Bloomberg
terminal / k9s / lazygit, on the web. Built with [`templ`](https://templ.guide) +
[htmx](https://htmx.org) + [_hyperscript](https://hyperscript.org). Zero Node
toolchain, zero JavaScript: all client behaviour ships as _hyperscript behaviors in
one file (`assets/baud._hs`).

Modeled on [oat.ink](https://oat.ink): semantic HTML first, components activated by
`data-*` attributes (`data-shell`, `data-panel`, `data-panes`), token-driven CSS in
`@layer theme, base, components, utilities`, and a deliberately starved utility layer
(`.hstack .vstack .fill .scroll .gap-0…3`) — no class soup. Server state (sort,
filter, pagination, palette search) is htmx round-trips; hyperscript handles only
purely-local UI (menus, tabs, dismiss).

Live showcase (all themes × densities): **https://markopolo123.github.io/baud-ui/**

## Quickstart

```sh
go get github.com/markopolo123/baud-ui
```

Components live in `github.com/markopolo123/baud-ui/baud`. Two assets must be served
to the browser:

| asset | what | how to get it |
|---|---|---|
| the CSS bundle | `tokens.css` + `base.css` + `components/*.css` + `utilities.css`, concatenated | `baudui.CSS()` (root package; built from embedded sources at first use) — or `just css` if you want the `dist/baud.css` file |
| the hyperscript behaviors file | `assets/baud._hs`, committed as-is | `baudui.HS()` — or serve the file directly |

htmx and _hyperscript are **peer dependencies**, not bundled — pinned versions are
exported as `baud.HTMXSrc` (htmx 2.0.4) and `baud.HyperscriptSrc` (_hyperscript
0.9.14).

**Script ordering rule:** `baud._hs` is loaded via
`<script type="text/hyperscript" src=…>` and MUST appear **before** the
_hyperscript library tag. Remotely-loaded behaviors have to be defined before
hyperscript boots, otherwise every `_="install …"` attribute resolves against
nothing and silently does nothing. `baud.Page` wires all of this in the correct
order for you:

```go
package main

import "github.com/markopolo123/baud-ui/baud"

templ Console() {
	@baud.Page(baud.PageProps{Title: "fleetctl"}) {
		@baud.Shell(baud.ShellProps{StatusBar: baud.StatusBar([]baud.StatusCell{
			{Text: "FLEET", Mode: true}, {Spring: true}, {Text: "eu-west-1"},
		})}) {
			@baud.Panel(baud.PanelProps{Title: "fleet", ID: "fleet",
				Actions: actions()}) {
				@baud.DataTable(baud.DataTableProps{
					ID:       "hosts",
					Endpoint: "/hosts", // sortable th → hx-get round-trip
					Columns: []baud.Column{
						{Key: "host", Label: "host", Sortable: true},
						{Key: "cpu", Label: "cpu%", Numeric: true, Sortable: true},
					},
					Rows: []baud.Row{
						{Key: "a1", Cells: []string{"ingest-gw-01", "82.4"}},
					},
				})
			}
		}
	}
}

templ actions() {
	@baud.Btn(baud.BtnProps{Label: "deploy", Variant: baud.BtnPrimary, Kbd: "⌘D"})
}
```

`PageProps` defaults assets to `assets/baud.css` / `assets/baud._hs`-relative hrefs;
override `CSSHref`/`HSHref` for other mount points. The root package embeds the
committed sources, so a plain `go get` builds with no pre-step: mount
`baudui.CSS()` / `baudui.HS()` in handlers (see `demo/server.go`) or write them to
disk at build time (see `cmd/render`). `just css` still emits `dist/baud.css` if
you want the bundle as a file.

## The class-swap contract

Theme, density, border mode and type mode are classes on the root element (`<body>`,
set by `baud.PageProps`). Swapping these classes is the **only** switching
mechanism — no other DOM or CSS changes, ever.

| axis | classes | default |
|---|---|---|
| theme | `t-gruvbox` `t-mocha` `t-sollight` | `t-gruvbox` |
| density | `d-ultra` `d-dense` `d-cozy` | `d-dense` |
| border | `b-line` `b-shade` `b-ascii` | `b-line` |
| type | `f-mono` `f-mix` | `f-mono` |

- Density sets six vars (`--fs --fs-sm --rh --row --pad --gap`); theme sets ~20
  semantic vars (below). Component CSS contains **no raw hex or px** — `var()` only.
- `b-line` = hairline borders; `b-shade` = borders dropped, separation via
  `--bg-raised` fills; `b-ascii` = dashed borders + panel titles wrapped in
  `┌─ … ─┐` glyphs (experimental).
- `f-mono` = everything monospace; `f-mix` = IBM Plex Sans for labels/buttons/tabs.

## Theming: map any iTerm2 scheme

A theme is one CSS class defining ~20 custom properties in `@layer theme`
(`assets/css/tokens.css` — the **only** file where raw hex is allowed). Every tint,
wash and hover derivative in the component layer is computed with
`color-mix(in srgb, var(--tone) N%, transparent)`, so a new theme needs **only the
base vars** — no per-theme tint values.

Recipe, starting from any published iTerm2 scheme (Background Color, Foreground
Color, ANSI 0–15):

1. **Base.** Scheme Background Color → `--bg-panel`; scheme Foreground Color →
   `--fg`. (If the scheme ships a "hard"/darker background variant, that darker
   value is `--bg-app` and the normal background is `--bg-panel`.)
2. **Surface ladder.** Derive five surfaces stepping *away* from black (dark
   themes lighten; light themes darken):
   `--bg-app` (app chrome, one step below panel) → `--bg-panel` → `--bg-raised`
   (sticky headers, chips) → `--bg-hover` → `--bg-active`. The scheme's
   selection/current-line colour is usually a perfect `--bg-active`. `--bg-input`
   sits one step *below* `--bg-app` (inputs read as wells).
3. **Borders are surfaces.** On dark themes `--border` = `--bg-hover` and
   `--border-strong` = `--bg-active` (true of shipped gruvbox and mocha). Light
   themes need borders slightly darker than the hover wash to stay visible
   (see `t-sollight`).
4. **Foreground ladder.** `--fg` from step 1; `--fg-faint` = the scheme's
   comment/bright-black (ANSI 8) colour; `--fg-muted` ≈ the midpoint between the
   two.
5. **Accent.** The scheme's identity colour: gruvbox = yellow, mocha and solarized
   = blue (per the design spec), Dracula = purple, Nord = frost blue. `--accent-2`
   is a secondary cool colour used sparingly. `--on-accent` is text rendered on
   accent fills — pick whichever of `--bg-app`/`--fg` contrasts more against the
   accent (for vivid accents on dark themes that's the background).
6. **Status tones.** ANSI green → `--ok`, yellow → `--warn`, red → `--err`,
   cyan → `--info`.
7. **Selection and shadow.** `--sel` = the accent at 13–14% alpha
   (`rgba(…, 0.13)`). `--shadow` = black at ~0.5 alpha on dark themes, ~0.18 on
   light.

Worked example — **Dracula** (not shipped; derived from the published scheme,
background `#282a36`, foreground `#f8f8f2`, current-line `#44475a`, comment
`#6272a4`, purple `#bd93f9`):

```css
@layer theme {
  .t-dracula {
    --bg-app: #21222c;        /* one step darker than scheme bg */
    --bg-panel: #282a36;      /* scheme Background Color */
    --bg-raised: #2f3142;
    --bg-hover: #343746;
    --bg-active: #44475a;     /* scheme current-line/selection */
    --bg-input: #1e1f29;      /* one step below --bg-app */
    --border: #343746;        /* = --bg-hover (dark theme) */
    --border-strong: #44475a; /* = --bg-active */
    --fg: #f8f8f2;            /* scheme Foreground Color */
    --fg-muted: #adb5cb;      /* midpoint of --fg and comment */
    --fg-faint: #6272a4;      /* comment (ANSI bright black) */
    --accent: #bd93f9;        /* purple — the scheme's identity colour */
    --accent-2: #ff79c6;      /* pink */
    --on-accent: #21222c;     /* light accent ⇒ dark text */
    --ok: #50fa7b;            /* ANSI green  */
    --warn: #f1fa8c;          /* ANSI yellow */
    --err: #ff5555;           /* ANSI red    */
    --info: #8be9fd;          /* ANSI cyan   */
    --sel: rgba(189, 147, 249, 0.13);
    --shadow: rgba(0, 0, 0, 0.55);
  }
}
```

Append the block after the bundle (or to your own tokens layer), set
`PageProps{Theme: "t-dracula"}`, done — every component, tint and hover state
follows.

## Components

All in package `baud`. Every interactive component is keyboard-operable with
correct ARIA roles. "behavior" = a `_="install …"` hyperscript behavior from
`assets/baud._hs`.

| component | what | htmx / hyperscript |
|---|---|---|
| `Page` | document wrapper: CSS bundle, `baud._hs` before the _hyperscript lib, pinned htmx/hyperscript CDNs, root mode classes, toast region | installs `PaletteKey` (⌘K) on body |
| `Shell` | `data-shell` app frame: topbar / optional nav / scrolling main / statusbar | — |
| `Panes` / `PanesRows` | tmux-style tiling, sizes in `ch`/`fr` | `Panes` behavior applies the grid (re-applied on `htmx:afterSettle`); `Resizable` adds drag gutters, persisted by `data-panes-id` |
| `Toolbar` | `.hstack` strip of controls | — |
| `Panel` | the universal container: header (title + actions slot) + scrollable body | `ID` is the natural `hx-target` |
| `Tabs` / `TabPanel` | `underline` or `boxed` variants, optional count badge | `Target` set: per-tab `hx-get` into the shared panel; `Target` empty: local mode via the `Tabs` behavior |
| `Breadcrumb` | `prod › core › ingest-gw` trail | — |
| `DefList` / `DefItem` | key/value grid, UPPERCASE faint keys, tabular-nums values | — |
| `StatusBar` | full-width cell strip, vim-style `mode` cell, one `spring` cell | — |
| `Btn` / `BtnGroup` / `Kbd` | button: default / `primary` / `danger` / `ghost`, glyph prefix, kbd hint chip | — |
| `Badge` / `Dot` | tone (ok/warn/err/info/accent/neutral) × variant (tint/solid/outline); pulsing status dot | — |
| `Field` | label + control + hint; `Error` flips hint to `✗ message` and the input border to err | — |
| `Input` | input row with `Prefix`/`Suffix` affixes, accent focus ring | — |
| `Checkbox` / `Radio` / `Toggle` | `[x]` / `(•)` glyph presentation over real inputs; segmented toggle | — |
| `Select` / `SelectMenu` | styled trigger + positioned menu, hidden input for forms | `SelectKeys` / `SelectPick` / `MenuDismiss` behaviors |
| `Combobox` / `ComboboxOptions` | type-to-filter input with `⌕` affix, match highlighting | local: `ComboboxFilter` behavior; server: debounced `hx-get` when `SearchURL` set |
| `DatePicker` / `DatePickerMenu` | `YYYY-MM-DD` trigger, Monday-first grid, `«‹›»` nav, presets | month nav is `hx-get` swapping the menu fragment |
| `TagInput` | `key=value` chips, Enter adds, Backspace pops, suggestions | `TagInput` + `MenuDismiss` behaviors |
| `DataTable` (+`Head`/`Body`) | sticky header, row selection, zebra/lines, threshold cell tones | sort: `th` `hx-get`s `Endpoint?sort=…` swapping the tbody, thead re-renders out-of-band |
| `Tree` / `TreeChildren` | box-drawing branches `├─ └─`, `▸/▾` disclosure, right meta | expand/collapse local (`<details>`); lazy branches `hx-get` children on first toggle |
| `Pagination` | `rows 1–50 of 12,403` + pager + optional `load more ↓` | `hx-get` per button (or plain links via `HrefFor`); load-more appends |
| `DiffViewer` | unified diff, dual gutters, tinted `+`/`−`/hunk rows | pure CSS, server renders the diff |
| `Modal` / `Drawer` | centered dialog / right-side drawer, Esc + backdrop close, focus trap/restore | `Overlay` behavior; open via `hx-get` + `hx-swap="beforeend"` on body |
| `Palette` / `PaletteResults` | ⌘K command palette, accent border, category column, `↑↓ ↵ esc` | `PaletteKey` + `Palette` behaviors; debounced `hx-get` server filtering |
| `Popover` | anchored quick-actions panel | inline hyperscript toggle + dismiss |
| `Tip` / `TipUnder` | tooltip as an attribute: spread onto any host | pure CSS `data-tip`, no script at all |
| `ToastRegion` / `Toast` / `ToastOOB` | bottom-right notification stack | `Toast` behavior auto-dismisses; servers push via `hx-swap-oob` into `#toasts` |
| `Progress` | ASCII `▰▰▰▱▱` bar + %, tone auto (accent → warn → ok) or forced | — |
| `Spinner` | braille-frame spinner, static under reduced-motion | — |
| `PanelState` | `skeleton` / `loading` / `empty` / `error` panel bodies | — |
| `ConfirmInput` | destructive guard: type the exact name to enable the action | inline hyperscript match check; put your `hx-delete` on the action Btn |

## Development

`just` is the only sanctioned entry point:

| recipe | does |
|---|---|
| `just generate` | `templ generate` (`*_templ.go` is gitignored) |
| `just css` | concatenate layered sources → `dist/baud.css` (convenience artifact; nothing else needs it) |
| `just build` | generate + `go build ./...` |
| `just run` | run the fleetctl demo console (`cmd/demo`) |
| `just render` | write the static showcase to `dist/site` (`cmd/render`) |
| `just lint` | gofmt check + `go vet` |
| `just test` | godog BDD + unit tests |
| `just e2e` | go-playwright browser tests (`just install-browsers` once first) |
| `just check` | lint + build + test + e2e — green or it doesn't merge |

Testing philosophy: BDD first — every component gets a `.feature` file under
`features/` before implementation, and scenarios must cover the **full options
matrix** (every variant/tone/state that changes rendering). go-playwright then
asserts real-browser behaviour: per-option computed styles, keyboard operability,
focus-visible, overlay dismiss, and theme/density switching by root-class swap
only. See [docs/WAYS_OF_WORKING.md](docs/WAYS_OF_WORKING.md) for the full pipeline
and review checklist. The design contract is `design/README.md`.

## Hard rules

- Zero border-radius. No emoji. No media queries except `prefers-reduced-motion`.
- No raw hex/px in component CSS — `var()` tokens only; hex lives solely in
  `assets/css/tokens.css`.
- No JavaScript: no `.js` files, no inline `<script>` logic — hyperscript
  behaviors only. PRs adding JS are rejected.
- Desktop only: `body { min-width: 1240px }`, no mobile layout.
- Themes/density/border/type switch via root class swap, nothing else.

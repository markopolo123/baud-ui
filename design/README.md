# Handoff: baud/ui — dense terminal component library

**Target stack: Go `templ` + htmx + _hyperscript. Zero Node toolchain.**

## Overview

baud/ui is a server-rendered component library for building dense, terminal-aesthetic
ops tooling (think Bloomberg terminal / k9s / lazygit, on the web). It was designed and
validated as an interactive HTML prototype ("fleetctl" demo console + component sheet).
Your job is to implement it as a real library: templ components, one CSS file set,
attribute-driven behavior via htmx/_hyperscript, no client framework.

## About the Design Files

The files bundled here are **design references created in HTML/React** — a prototype
showing intended look and behavior, NOT production code. The React/JSX is throwaway
scaffolding used to make the prototype clickable. **Do not port the React.** Recreate
the rendered DOM + CSS as templ components. The CSS files, however, are close to
production-intent: selectors, custom properties and values can be lifted nearly verbatim
(rename freely). Open `Baud UI Prototype.html` in a browser to see everything working;
use the Tweaks panel to switch theme/density/border/type modes.

## Fidelity

**High-fidelity.** Colors, spacing, type sizes, row heights, hover/focus states and
interaction patterns are final design intent. Match them. The component inventory and
attribute API below are the contract; internal implementation is yours.

## Core philosophy (non-negotiables)

Modeled on knadh/oat (https://oat.ink) — adapted for dense desktop tooling:

1. **Semantic HTML first, `data-*` attributes for components, no class soup.**
   Native elements styled contextually. Components are activated by data attributes
   (`data-panel`, `data-shell`), not stacks of utility classes. A small, deliberately
   starved utility layer exists for composition glue only.
2. **Zero dependencies, no build step for consumers.** One CSS bundle + one tiny JS
   bundle (only for the few behaviors CSS can't do). htmx and _hyperscript are peer
   assumptions, not bundled.
3. **CSS `@layer theme, base, components, utilities`** so consumer overrides always win
   predictably.
4. **Everything flows through tokens.** Theme = ~20 custom properties. Density = 6
   custom properties. No raw hex or px inside component CSS — only `var()`.
5. **Desktop only. No mobile responsiveness.** Hard floor: `body { min-width: 1240px }`,
   horizontal scroll below that. No media queries except `prefers-reduced-motion`.
6. **No 12-column grid.** Terminals tile, they don't flow. Layout = shell + panes +
   stacks (below).
7. **Server-rendered state.** Sorting, pagination, filtering are htmx round-trips.
   _hyperscript handles purely-local UI (menu open/close, tab switch, copy-to-clipboard).

## Layout system

### 1. App shell — `data-shell`

One attribute builds the whole frame (CSS grid, named areas, `100dvh`).
*(Historical handoff edit: shipped as `<div data-shell>` directly inside `<body>`.)*

```html
<div data-shell>
  <header data-topbar>…brand, tabs, global actions…</header>
  <aside data-nav>…tree / navigator…</aside>
  <main>…panes…</main>
  <footer data-statusbar>…</footer>
</div>
```

- Grid: columns `var(--nav-w, 28ch) 1fr`, rows `auto 1fr auto`.
- `data-topbar` and `data-statusbar` span `1 / -1`. `main` is the only scroll container
  (`overflow: auto; min-width: 0; min-height: 0`).
- `data-nav` is optional; without it the shell is a single column.
- The statusbar is load-bearing in this aesthetic — always render it.

### 2. Panes — `data-panes`

The workhorse. Tiles a region like tmux; value is the grid template, sizes in `ch`
(mono-honest) or `fr`:

```html
<div data-panes="42ch 1fr">          <!-- vertical split -->
<div data-panes-rows="1.6fr 1fr">    <!-- horizontal split -->
```

- Implementation: `display: grid; grid-template-columns: attr()`-style is not possible
  in CSS — read the attribute in the tiny JS bundle once at boot (and on
  `htmx:afterSettle`), set `style.gridTemplateColumns`. Children get
  `min-width: 0; min-height: 0; overflow: auto`.
- Nestable arbitrarily.
- **Resizing (progressive enhancement):** panes work 100% without JS. Adding
  `data-resizable` injects `7px` gutters (`cursor: col-resize`, `⋮` glyph, accent
  highlight on hover/drag — see `.split-gutter` in prototype CSS) handled by the JS
  bundle with pointer events. Persist sizes to `localStorage` keyed by `data-panes-id`
  when present. This fits the philosophy: CSS-only core, sliver of JS as opt-in.

### 3. Stacks — the only utilities

```css
.hstack  /* flex row, align-items center, gap: var(--gap) */
.vstack  /* flex column, gap: var(--gap) */
.fill    /* flex: 1; min-width: 0 */
.scroll  /* overflow: auto; min-height: 0 */
.gap-0 .gap-1 .gap-2 .gap-3   /* 0, calc(var(--gap)*0.5), var(--gap), calc(var(--gap)*2) */
```

That is the complete utility inventory. Resist adding more. Forms use a
`data-form-grid` component (2-col `max-content 1fr` or equal halves via
`data-form-grid="2"`), not a general grid.

## Design tokens

### Density (class on root: one of three)

| token | `d-ultra` | `d-dense` (default) | `d-cozy` |
|---|---|---|---|
| `--fs` (base font) | 11px | 12px | 13px |
| `--fs-sm` (labels/meta) | 10px | 10.5px | 11px |
| `--rh` (control height) | 20px | 24px | 30px |
| `--row` (table/list row) | 18px | 22px | 27px |
| `--pad` (cell/panel pad) | 4px | 6px | 8px |
| `--gap` (stack gap) | 4px | 6px | 8px |

Line-height 1.45. `font-variant-numeric: tabular-nums` on all numeric/table content.

### Type

- `--font-mono`: JetBrains Mono (default) or IBM Plex Mono. Data, inputs, code, logs.
- `--font-ui`: mode-switched. `f-mono` mode ⇒ same as mono. `f-mix` mode ⇒ IBM Plex Sans
  for labels/buttons/tab text. Both modes ship; mono is default.
- Labels/buttons/panel titles: `--fs-sm`, weight 600, `letter-spacing 0.07em`, UPPERCASE.
- **Zero border-radius everywhere. No rounded corners. Ever.**

### Themes (iTerm2 schemes → ~20 semantic vars)

Theme = class on root. Ship these three; the mapping recipe is the deliverable that
makes "any iTerm2 scheme" possible later (bg/fg from scheme base, accent from yellow
(gruvbox) / blue (mocha/solarized), ok/warn/err/info from green/yellow/red/cyan).

| var | `t-gruvbox` (default) | `t-mocha` | `t-sollight` |
|---|---|---|---|
| `--bg-app` | `#1d2021` | `#11111b` | `#eee8d5` |
| `--bg-panel` | `#282828` | `#181825` | `#fdf6e3` |
| `--bg-raised` | `#32302f` | `#1e1e2e` | `#f4eedb` |
| `--bg-hover` | `#3c3836` | `#313244` | `#e9e2cd` |
| `--bg-active` | `#504945` | `#45475a` | `#ddd6c1` |
| `--bg-input` | `#1d2021` | `#0d0d15` | `#fdf6e3` |
| `--border` | `#3c3836` | `#313244` | `#d9d2bc` |
| `--border-strong` | `#504945` | `#45475a` | `#c4bda6` |
| `--fg` | `#ebdbb2` | `#cdd6f4` | `#586e75` |
| `--fg-muted` | `#a89984` | `#a6adc8` | `#657b83` |
| `--fg-faint` | `#7c6f64` | `#6c7086` | `#93a1a1` |
| `--accent` | `#fabd2f` | `#89b4fa` | `#268bd2` |
| `--accent-2` | `#83a598` | `#cba6f7` | `#2aa198` |
| `--on-accent` | `#1d2021` | `#11111b` | `#fdf6e3` |
| `--ok` | `#b8bb26` | `#a6e3a1` | `#859900` |
| `--warn` | `#fabd2f` | `#f9e2af` | `#b58900` |
| `--err` | `#fb4934` | `#f38ba8` | `#dc322f` |
| `--info` | `#83a598` | `#89dceb` | `#2aa198` |
| `--sel` (selection bg) | `rgba(250,189,47,.14)` | `rgba(137,180,250,.13)` | `rgba(38,139,210,.13)` |
| `--shadow` | `rgba(0,0,0,.5)` | `rgba(0,0,0,.55)` | `rgba(0,0,0,.18)` |

Derived tints use `color-mix(in srgb, var(--tone) 15%, transparent)` — keep that
technique so new themes need only the base vars. `::selection` is accent/on-accent.
Style scrollbars (track = `--bg-app`, thumb = `--bg-active`, 10px).

### Border modes (class on root)

- `b-line` (default): 1px `--border` / `--border-strong` hairlines everywhere.
- `b-shade`: borders transparent; separation via `--bg-raised` fills.
- `b-ascii`: dashed borders + panel titles wrapped in `┌─ … ─┐` glyphs (prototype
  approximates with `::before/::after`). Treat as experimental; keep behind the mode
  switch.

## Component inventory & attribute API

Each component = one templ func returning semantic HTML. Names/props below are the
contract; suggested signatures are Go-flavored pseudocode. Every interactive component
must work with keyboard and expose correct ARIA roles.

### Primitives

- **Btn** `templ Btn(p BtnProps)` — `<button>`. Props: `Variant` (default/primary/
  danger/ghost), `Glyph` (mono char prefix), `Kbd` (inline shortcut hint chip),
  `Active`, `Disabled`. Height `--rh`; label style per Type rules. Danger = transparent
  with err border, fills err on hover. **BtnGroup** fuses borders (-1px margin).
- **Kbd** — bordered chip, `--bg-raised`, `--fs-sm`.
- **Badge** — `Tone` (ok/warn/err/info/accent/neutral) × `Variant` (tint/solid/outline),
  optional 5px square dot. Tint = 15% color-mix bg + 38% border.
- **Dot** — 7px round status dot; `Pulse` adds box-shadow ping (reduced-motion gated).
- **Field** — label (UPPERCASE `--fs-sm`) + control + hint line; `Error` string flips
  hint to `✗ message` in `--err` and sets input border err.
- **Input** — wrapper row with optional `Prefix`/`Suffix` affixes (`--fg-faint`).
  Focus: accent border + 1px accent glow ring.
- **Checkbox / Radio** — text glyphs `[x] [ ]` / `(•) ( )`, accent when on. Real
  `<input>` underneath for forms/a11y; glyphs are presentation.
- **Toggle** (segmented) — bordered strip, selected segment = accent bg.
- **Select** — trigger `▼/▲` + absolutely-positioned menu (`--bg-panel`, strong border,
  shadow). htmx-friendly: can be a real `<select>` styled to match, with the custom
  menu as enhancement.
- **Combobox** — input with `⌕` affix; type-to-filter menu, match substring highlighted
  in accent bold, per-option right-aligned meta, `↑↓ ↵ esc` keys. Filtering can be
  client-side (hyperscript) or `hx-get` server round-trip — support both.
- **DatePicker** — trigger shows `YYYY-MM-DD`; menu = Monday-first 6×7 grid, `«‹›»`
  month/year nav, today outlined, selected = accent fill, out-month dimmed; preset row
  `today / -1d / -7d / -30d`.
- **TagInput** — chips `key=value` (key in `--fg-faint`), accent-tinted chip bg, `✕`
  remove, Enter adds, Backspace pops, suggestion menu filtered.

### Structure

- **Panel** — `data-panel`: header (height `--rh`, title UPPERCASE `--fs-sm` muted,
  right-aligned actions slot) + scrollable body. The universal container.
- **Tabs** — two variants: `underline` (2px accent underline on active) and `boxed`
  (bordered strip, active = accent fill). Optional count badge per tab. Tab switching
  is htmx (`hx-get` into a target) or hyperscript for local panes.
- **Breadcrumb** — `prod › core › ingest-gw`; links faint, current bold, `›` separators.
- **DefList** — CSS grid `max-content 1fr`; keys UPPERCASE faint, values tabular-nums,
  optional hairline row rules.
- **StatusBar** — full-width cell strip, cells separated by hairlines, first cell may be
  `mode` (accent bg, bold) — vim style. One `spring` cell flexes.
- **Toolbar** — `.hstack` of Btns/Selects/Inputs.

### Data

- **DataTable** — sticky header row (`--bg-raised`, UPPERCASE), row height `--row`,
  hover `--bg-hover`, selected row `--sel` + 2px accent inset bar + `▌` row mark.
  Numeric columns right-aligned tabular-nums. Optional `zebra`, `lines` (column rules).
  **Sort = htmx**: `<th hx-get="?sort=cpu&dir=desc">` swapping the table body; arrow
  `▲/▼` in accent on the sorted column. Tone-code threshold cells (e.g. cpu>90 ⇒ err).
- **Tree** — box-drawing branches `├─ └─ │` in `--fg-faint`, `▸/▾` disclosure, optional
  right-aligned meta, selected = `--sel` + accent inset bar. Expand/collapse local
  (hyperscript or `<details>`); lazy-load children via `hx-get` is the intended pattern.
- **Pagination** — footer bar: `rows 1–50 of 12,403` + `|‹ ‹ prev n/N next › ›|` +
  optional accent `load more ↓` (htmx append).
- **DiffViewer** — unified diff; dual line-number gutters, `+`/`−` rows tinted 10%
  ok/err, hunk headers tinted info. Server renders the diff; this is pure CSS.

### Overlays (htmx `hx-target="body" hx-swap="beforeend"` pattern, or `<dialog>`)

- **Modal** — centered at `9vh`, 540px, strong border + deep shadow. Esc + backdrop
  close. Footer right-aligned actions.
- **Drawer** — right side, 420px, full height. Same close behavior.
- **CommandPalette** — `⌘K`. 560px at `12vh`, **accent border**, `›` prompt, category
  column (64px, faint, UPPERCASE) + label + right shortcut chips, `↑↓ ↵ esc`,
  highlighted row = `--sel` + accent inset bar. Command list can be server-filtered
  (`hx-get` on keyup, debounced) — this is the flagship htmx integration.
- **Popover** — anchored 280px panel for quick actions.
- **Tooltip** — pure CSS `data-tip` attribute (`::after` = `pre` whitespace so
  multi-line aligned mono tips work; 150ms delay; arrow `::before`). `tip-under` adds
  dotted underline + help cursor.
- **Toasts** — bottom-right stack above statusbar; 3px left tone bar + tone glyph
  `✓ ✗ ▲ ℹ`; slide-in 160ms; auto-dismiss ~4s. htmx: server pushes via OOB swap into
  a `#toasts` region.

### Feedback

- **Progress** — ASCII `▰▰▰▱▱` (22 chars default) + right-aligned %, tone auto
  (accent → warn >85% → ok at 100%) or forced.
- **Spinner** — braille frames `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` @80ms, accent. Reduced-motion: static.
- **PanelState** — `skeleton` (animated `--bg-active` bars, staggered), `loading`
  (spinner + cmd hint), `empty` (`∅` + title + sub + optional action), `error`
  (`✗` err + retry action).
- **ConfirmInput** — destructive guard: "type *name* to confirm", action Btn disabled
  until exact match; mismatch shows err border while typing.

## Interactions & behavior

- Focus-visible: 1px accent outline, 1px offset — everywhere.
- Menus/popovers close on outside-click and Esc. One overlay at a time.
- Keyboard-first is *subtle*: Kbd hint chips on primary actions, `⌘K` palette, `↑↓ ↵`
  in menus. No vim-mode by default.
- Animation policy: ~160ms entrances for overlays/toasts, spinner/skeleton/pulse loops —
  all inside `@media (prefers-reduced-motion: no-preference)`. Nothing else animates.
- htmx conventions: `hx-indicator` shows Spinner; failed requests toast err via OOB;
  table sort/filter/pagination always server round-trip; URL is the state.

## Deliverables checklist (suggested order)

1. `baud.css` (layers: theme/base/components/utilities) + client behaviours (panes,
   palette keys, menu dismiss). *(Historical handoff edit: shipped as `assets/baud._hs`
   _hyperscript behaviors — the project's no-JS rule replaced the planned `baud.js`.)*
2. templ package: primitives → structure → data → overlays → feedback (inventory above).
3. A Go demo app reproducing the prototype's **fleetctl console** (shell, metrics strip
   via DefList/Panel, sortable DataTable with htmx, Tree nav, log tail, drawer, modal,
   palette, toasts) — this is the acceptance test.
4. A component-sheet page (kitchen sink) rendered from the templ components.
5. README documenting the theme-mapping recipe for arbitrary iTerm2 schemes.

## Acceptance criteria

- Pixel-comparable to the prototype at `d-dense` + `t-gruvbox` + `b-line` + `f-mono`.
- Theme/density/border/type switchable by swapping root classes only — no other DOM or
  CSS changes.
- Every component keyboard-operable; menus/overlays trap and restore focus.
- Works with JS disabled except: palette, pane-resize, combobox filter, spinner
  (graceful: real `<select>`, static panes, plain input, static glyph).
- No border-radius anywhere. No emoji. No media queries except reduced-motion.
- `min-width: 1240px` on body; no mobile layout.

## Files in this bundle

| file | what it is |
|---|---|
| `Baud UI Prototype.html` | Entry point — open in a browser. View switcher: demo console / component sheet. Tweaks panel switches theme/density/border/type. |
| `baud-tokens.css` | **Lift-ready.** Tokens, themes, density, primitives (buttons, badges, inputs, selects, date picker). |
| `baud-components.css` | **Lift-ready.** Tabs, panels, tables, tree, statusbar, overlays, palette, toasts, border modes. |
| `baud-extras.css` | **Lift-ready.** Tags, progress, deflist, breadcrumb, tooltip, states, pagination, split pane, diff. |
| `baud-shared.jsx` `baud-widgets.jsx` `baud-pickers.jsx` `baud-extras.jsx` | Prototype-only React. Reference for DOM structure & behavior, do not port. |
| `baud-sheet.jsx` `baud-sheet-extras.jsx` `baud-app.jsx` | Prototype-only: component sheet + fleetctl demo (content/copy reference). |
| `tweaks-panel.jsx` | Prototype tooling only. Ignore entirely. |

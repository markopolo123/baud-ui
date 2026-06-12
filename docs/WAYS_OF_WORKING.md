# Ways of Working — agent orchestration

How this library gets built. The orchestrator (main Claude session) manages builder
agents and gatekeeps their output. Humans audit asynchronously via PR history.

## Roles

- **Orchestrator / arbiter** — plans waves, spawns builders, reviews every PR, merges.
  Final authority. Never rubber-stamps: reads the diff, runs `just check` itself when
  in doubt.
- **Builder agent** — owns one component (or tight group) per wave. Works in an
  isolated git worktree on a `feat/<name>` branch. BDD-first per CLAUDE.md. Opens a PR
  with the DoD checklist filled in. Never reviews its own work, never merges.
- **Reviewer agent** — spawned per PR. Adversarial: tries to find spec-fidelity gaps
  (vs `design/README.md`), missing option-matrix coverage, token violations
  (raw hex/px, border-radius), a11y/keyboard gaps, untested htmx paths. Returns a
  verdict: approve / request-changes with a concrete list.

## Pipeline (per component)

1. Builder gets a worktree + `feat/<component>` branch.
2. Writes `features/<component>.feature` first — red.
3. Implements templ + CSS (lift from `design/*.css` where possible) — green.
4. Adds go-playwright coverage in `e2e/`.
5. Adds the component to the component-sheet page (DoD: it must ship in the showcase).
6. `just check` green locally → push → `gh pr create` with DoD checklist.
7. Reviewer agent audits → verdict on the PR.
8. Arbiter inspects the diff + verdict. Rework loops back to the same builder via a
   fresh agent prompt with the reviewer's findings. Merge only when DoD is met.
9. Merge to main triggers Pages deploy of the re-rendered showcase.

## Waves

| wave | scope | status |
|---|---|---|
| 0 | Foundation: module, justfile, CSS layer scaffold + token lift, baud.js, shell/panes/stacks, godog + playwright harnesses, demo server skeleton, component-sheet skeleton, `cmd/render`, CI + Pages workflows | pending |
| 1 | Primitives: Btn/BtnGroup/Kbd · Badge/Dot · Field/Input · Checkbox/Radio/Toggle | pending |
| 2 | Form controls: Select · Combobox · DatePicker · TagInput; Structure: Panel · Tabs · Breadcrumb · DefList · StatusBar · Toolbar | pending |
| 3 | Data: DataTable (htmx sort) · Tree · Pagination · DiffViewer | pending |
| 4 | Overlays: Modal · Drawer · CommandPalette · Popover · Tooltip · Toasts; Feedback: Progress · Spinner · PanelState · ConfirmInput | pending |
| 5 | fleetctl demo console (acceptance test) · theme-mapping recipe README · pixel pass vs prototype | pending |

Within a wave, builders run in parallel. A wave starts only when the previous wave is
merged (components depend on foundation; demo depends on components).

## Conflict avoidance

- One CSS source file per component area under `assets/css/components/`; bundle via
  `just css`. Builders never touch files owned by another in-flight branch.
- The component sheet is one page but sectioned; builders append their own section
  file (`demo/sheet_<component>.templ`) and register it in the shared registry
  (`demo/registry.go`) — the registry line is the only expected merge-conflict point
  and is trivial.

## Review checklist (reviewer agent + arbiter)

- Feature file exists, was written first (check commit order), scenarios enumerate the
  options matrix — not just the happy path.
- Playwright asserts computed styles per option (e.g. variant colours), keyboard, focus.
- No raw hex/px in component CSS; no border-radius; tokens resolve in all 3 themes.
- Semantic HTML + ARIA per design spec; htmx round-trips for server state, hyperscript
  only for local UI.
- No `.js` files or inline script logic — client behaviour is hyperscript only
  (`assets/baud._hs` behaviors or inline `_=` attributes).
- Component-sheet section present and rendering.
- Generic test helpers live ONLY in steps_test.go / e2e/helpers_test.go — component
  files define component-prefixed helpers only.
- CI green. No drive-by changes outside the component's scope.

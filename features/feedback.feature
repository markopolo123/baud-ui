Feature: Feedback components — Progress, Spinner, PanelState, ConfirmInput
  Progress is a server-rendered ASCII bar (▰ filled / ▱ rest, 22 glyphs by
  default) plus a right-aligned tabular-nums percent; its tone is auto
  (accent → warn past 85% → ok at 100%) unless forced. Spinner is an accent
  braille glyph whose frames cycle in CSS (::before content animation), so
  the markup is a single status span. PanelState fills a panel body with one
  of four states: skeleton (AT-hidden animated bars), loading (spinner + cmd
  hint), empty (∅ + title + sub + optional action) and error (✗ in --err +
  retry action). ConfirmInput is the destructive guard: the danger action
  button stays disabled until the typed value exactly matches data-confirm —
  compared by inline hyperscript on the input, no JavaScript anywhere.

  # ---- Progress -----------------------------------------------------------

  Scenario: progress renders bar, percent and progressbar ARIA
    When I render a progress at 50
    Then exactly 1 element matches "span.prog[role=progressbar]"
    And the element "span.prog" has attribute "aria-valuemin" equal to "0"
    And the element "span.prog" has attribute "aria-valuemax" equal to "100"
    And the element "span.prog" has attribute "aria-valuenow" equal to "50"
    And the element "span.prog-bar" has text "▰▰▰▰▰▰▰▰▰▰▰▱▱▱▱▱▱▱▱▱▱▱"
    And the element "span.prog-rest" has text "▱▱▱▱▱▱▱▱▱▱▱"
    And the element "span.prog-pct" has text "50%"

  Scenario: zero progress is an all-rest accent bar
    When I render a progress at 0
    Then the element "span.prog-bar" has classes "prog-bar tone-accent"
    And the element "span.prog-rest" has text "▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱▱"
    And the element "span.prog-pct" has text "0%"

  Scenario: tone stays accent at the 85 boundary
    When I render a progress at 85
    Then the element "span.prog-bar" has classes "prog-bar tone-accent"

  Scenario: tone flips to warn just past 85
    When I render a progress at 86
    Then the element "span.prog-bar" has classes "prog-bar tone-warn"

  Scenario: warn holds at 99
    When I render a progress at 99
    Then the element "span.prog-bar" has classes "prog-bar tone-warn"

  Scenario: complete progress is an all-filled ok bar
    When I render a progress at 100
    Then the element "span.prog-bar" has classes "prog-bar tone-ok"
    And the element "span.prog-bar" has text "▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰▰"
    And the element "span.prog-pct" has text "100%"
    And no element matches "span.prog-rest"

  Scenario: a forced tone overrides the auto thresholds
    When I render a progress at 100 with forced tone "err"
    Then the element "span.prog-bar" has classes "prog-bar tone-err"
    And no element matches "span.prog-bar.tone-ok"

  Scenario: values clamp to the 0–100 range
    When I render a progress at 150
    Then the element "span.prog" has attribute "aria-valuenow" equal to "100"
    And the element "span.prog-pct" has text "100%"
    And the element "span.prog-bar" has classes "prog-bar tone-ok"

  Scenario: negative values clamp to zero
    When I render a progress at -10
    Then the element "span.prog" has attribute "aria-valuenow" equal to "0"
    And the element "span.prog-pct" has text "0%"

  Scenario: the label renders and the bar width is configurable
    When I render a progress at 50 labelled "rebalance" with 10 chars
    Then the element "span.prog-label" has text "rebalance"
    And the element "span.prog-bar" has text "▰▰▰▰▰▱▱▱▱▱"
    And the element "span.prog" has attribute "aria-label" equal to "rebalance"

  Scenario: no label span renders without a label
    When I render a progress at 50
    Then no element matches "span.prog-label"

  # ---- Spinner -------------------------------------------------------------

  Scenario: spinner is a single status span — frames live in CSS
    When I render a spinner
    Then exactly 1 element matches "span.spinner[role=status]"
    And the element "span.spinner" has attribute "aria-label" equal to "loading"
    And the element "span.spinner" has text ""

  # ---- PanelState ----------------------------------------------------------

  Scenario: skeleton renders five staggered bar rows, hidden from AT
    When I render a "skeleton" panel state
    Then exactly 1 element matches "div.pstate-skel[aria-hidden=true]"
    And exactly 5 elements match "div.pstate-skel > div.skel-row"
    And exactly 20 elements match "div.skel-row > span.skel-cell"
    And no element matches "div.pstate"

  Scenario: loading renders the spinner plus the cmd hint text
    When I render a loading panel state hinting "$ fleetctl get units --watch"
    Then exactly 1 element matches "div.pstate > span.pstate-glyph > span.spinner"
    And the element "span.pstate-title" has text "loading units"
    And the element "span.pstate-sub" has text "$ fleetctl get units --watch"
    And no element matches "div.pstate.err"

  Scenario: empty renders the ∅ glyph with title and sub
    When I render an "empty" panel state
    Then exactly 1 element matches "div.pstate"
    And the element "span.pstate-glyph" has text "∅"
    And the element "span.pstate-title" has text "no units"
    And the element "span.pstate-sub" has text "try widening the filter"
    And no element matches "div.pstate.err"
    And no element matches "span.pstate-act"

  Scenario: the empty action slot renders when given
    When I render an "empty" panel state with an action
    Then exactly 1 element matches "div.pstate > span.pstate-act"
    And the element "span.pstate-act > span" has text "clear filters"

  Scenario: error flips to err styling with the ✗ glyph and a retry slot
    When I render an "error" panel state with an action
    Then exactly 1 element matches "div.pstate.err[role=alert]"
    And the element "span.pstate-glyph" has text "✗"
    And the element "span.pstate-title" has text "fetch failed"
    And exactly 1 element matches "div.pstate.err > span.pstate-act"

  # ---- ConfirmInput ---------------------------------------------------------

  Scenario: confirm input guards with a hint, the expected name and a disabled danger action
    When I render a confirm input expecting "db-04" with action "decommission"
    Then the element "span.field-hint" has text "type db-04 to confirm"
    And the element "span.confirm-expect" has text "db-04"
    And exactly 1 element matches "div.confirm > div.confirm-row > span.input-wrap > input.input[data-confirm=db-04]"
    And exactly 1 element matches "div.confirm > div.confirm-row > button.btn.btn-danger[disabled]"
    And the element "button.btn-danger" has text "decommission"
    And the element "input.input" has attribute "placeholder" equal to "db-04"

  Scenario: the comparison is wired as inline hyperscript on the input
    When I render a confirm input expecting "db-04" with action "decommission"
    Then the confirm input compares its value against data-confirm via hyperscript

  Scenario: real form semantics — the input is named for submission
    When I render a confirm input expecting "db-04" named "unit"
    Then the element "input.input" has attribute "name" equal to "unit"

  Scenario: the form name defaults to confirm
    When I render a confirm input expecting "db-04" with action "decommission"
    Then the element "input.input" has attribute "name" equal to "confirm"

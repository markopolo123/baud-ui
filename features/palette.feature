Feature: CommandPalette — ⌘K overlay with a server-filtered htmx command list
  The flagship htmx integration: PaletteKey (on <body>) sends baud:palette,
  the Palette behavior opens the overlay — 560px at 12vh, accent border,
  › prompt affix — and traps focus in the query input. Rows are category
  column + label + right-aligned Kbd shortcut chips; typing fires a
  debounced hx-get that swaps the PaletteResults fragment into the
  listbox; ↑↓ move the .hl highlight (--sel + accent inset bar), ↵
  activates (anchor rows navigate; action rows carry an opaque data-cmd
  id the behavior dispatches as baud:paletteCmd to body — never executed),
  Esc/backdrop click closes and restores focus. ARIA follows the
  combobox/listbox pattern (aria-activedescendant on the input,
  role=option rows). Command spec format in steps:
  "cat|label|kbd|target" — target "/…" is an Href, anything else an
  Action, empty is inert.

  # ---- overlay shell -------------------------------------------------------

  Scenario: live palette renders the overlay closed with the Palette behavior
    When I render a palette with id "pal" searching "/api/palette" with commands "go|go to sheet|g s|/sheet"
    Then exactly 1 element matches "div.palette-overlay[id=pal]"
    And the element "div.palette-overlay" has attribute "_" equal to "install Palette"
    And no element matches "div.palette-overlay.open"
    And exactly 1 element matches "div.palette-overlay > div.palette[role=dialog][aria-modal=true][aria-label=command palette]"

  Scenario: the prompt affix and keyboard hint chips frame the query input
    When I render a palette with id "pal" searching "/api/palette" with commands "go|go to sheet|g s|/sheet"
    Then exactly 1 element matches "div.palette-input > span.prompt[aria-hidden=true]"
    And the element "span.prompt" has text "›"
    And exactly 1 element matches "div.palette-input > input.input[type=text]"
    And exactly 2 elements match "div.palette-input > kbd.kbd"

  Scenario: the input carries the combobox ARIA wiring over the listbox
    When I render a palette with id "pal" searching "/api/palette" with commands "go|go to sheet|g s|/sheet,fleet|deploy canary|⌘D|deploy"
    Then exactly 1 element matches "input.input[role=combobox][id=pal-input]"
    And the element "input.input" has attribute "aria-expanded" equal to "true"
    And the element "input.input" has attribute "aria-controls" equal to "pal-list"
    And the element "input.input" has attribute "aria-autocomplete" equal to "list"
    And the element "input.input" has attribute "autocomplete" equal to "off"
    And no element matches "input.input[aria-activedescendant]"
    And exactly 1 element matches "div.palette-list[role=listbox][id=pal-list]"
    And exactly 2 elements match "div.palette-list > .palette-item[role=option][tabindex=-1]"
    And each palette row has an id derived from "pal"

  Scenario: typing round-trips through a debounced hx-get swapping the list
    When I render a palette with id "pal" searching "/api/palette" with commands "go|go to sheet|g s|/sheet"
    Then the element "input.input" has attribute "hx-get" equal to "/api/palette"
    And the element "input.input" has attribute "hx-trigger" equal to "keyup changed delay:150ms"
    And the element "input.input" has attribute "hx-target" equal to "#pal-list"
    And the element "input.input" has attribute "hx-swap" equal to "innerHTML"
    And the element "input.input" has attribute "name" equal to "q"

  # ---- row anatomy ---------------------------------------------------------

  Scenario: rows are category column + label + right-aligned shortcut chips
    When I render a palette with id "pal" searching "/api/palette" with commands "fleet|deploy canary|⌘ D|deploy"
    Then exactly 1 element matches ".palette-item > span.pi-cat"
    And the element "span.pi-cat" has text "fleet"
    And exactly 1 element matches ".palette-item > span.pi-label"
    And the element "span.pi-label" has text "deploy canary"
    And exactly 2 elements match ".palette-item > span.pi-keys > kbd.kbd"

  Scenario: a command without a shortcut renders no chip group
    When I render a palette with id "pal" searching "/api/palette" with commands "fleet|restart workers||restart"
    Then exactly 1 element matches ".palette-item"
    And no element matches "span.pi-keys"

  Scenario: href commands are real anchors that navigate on activation
    When I render a palette with id "pal" searching "/api/palette" with commands "go|go to sheet|g s|/sheet"
    Then exactly 1 element matches "a.palette-item[role=option][href=/sheet]"
    And no element matches "button.palette-item"

  Scenario: action commands are buttons carrying the opaque command id, never live script
    When I render a palette with id "pal" searching "/api/palette" with commands "fleet|deploy canary|⌘D|deploy-canary"
    Then exactly 1 element matches "button.palette-item[role=option][type=button]"
    And the element "button.palette-item" has attribute "data-cmd" equal to "deploy-canary"
    And no element matches ".palette-item[_]"
    And no element matches "a.palette-item"

  # ---- result fragment (the hx-get response body) --------------------------

  Scenario: the results fragment pre-highlights exactly one row when asked
    When I render palette results for id "pal" highlighting row 2 from "go|a||/a,go|b||/b,go|c||/c"
    Then exactly 3 elements match ".palette-item[role=option]"
    And exactly 1 element matches ".palette-item.hl"
    And exactly 1 element matches ".palette-item.hl[id=pal-cmd-1][aria-selected=true]"
    And exactly 2 elements match ".palette-item[aria-selected=false]"

  Scenario: an unhighlighted fragment marks no row
    When I render palette results for id "pal" highlighting row 0 from "go|a||/a,go|b||/b"
    Then no element matches ".palette-item.hl"
    And exactly 2 elements match ".palette-item[aria-selected=false]"

  Scenario: an empty result set shows the ∅ empty state
    When I render palette results for id "pal" with no commands
    Then exactly 1 element matches "div.palette-empty"
    And the element "div.palette-empty" has text "∅ no matches"
    And no element matches "[role=option]"

  # ---- static sheet variant -------------------------------------------------

  Scenario: the static palette renders open and inert for styling assertions
    When I render a static palette with id "pal-s" highlighting row 1 with commands "go|go to sheet|g s|/sheet,fleet|deploy|⌘D|deploy"
    Then exactly 1 element matches "div.palette-overlay.open.palette-static[id=pal-s]"
    And no element matches "div.palette-overlay[_]"
    And no element matches "input[hx-get]"
    And exactly 1 element matches ".palette-item.hl[id=pal-s-cmd-0]"

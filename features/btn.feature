Feature: Btn, BtnGroup and Kbd primitives
  Buttons are real <button> elements styled by class markers; Kbd is a
  bordered shortcut chip; BtnGroup fuses adjacent buttons. The options
  matrix below (variants x states x glyph/kbd slots) is the contract from
  design/README.md "Primitives".

  Scenario: default button renders the base marker only
    When I render a button "deploy"
    Then exactly 1 element matches "button.btn"
    And the element "button.btn" has attribute "type" equal to "button"
    And the element "button.btn" has text "deploy"
    And no element matches "button.btn-primary"
    And no element matches "button.btn-danger"
    And no element matches "button.btn-ghost"
    And no element matches "button.is-active"
    And no element matches "button[disabled]"
    And no element matches "button.btn > span.btn-glyph"
    And no element matches "button.btn > kbd.kbd"

  Scenario Outline: each variant renders its variant marker
    When I render a "<variant>" button "act"
    Then exactly 1 element matches "button.btn"
    And the element "button.btn" has classes "btn btn-<variant>"

    Examples:
      | variant |
      | primary |
      | danger  |
      | ghost   |

  Scenario: disabled button renders the disabled attribute
    When I render a disabled button "halt"
    Then exactly 1 element matches "button.btn[disabled]"

  Scenario Outline: active state adds is-active on any variant
    When I render an active "<variant>" button "6h"
    Then exactly 1 element matches "button.btn.is-active"
    And the element "button.btn" has classes "<classes>"

    Examples:
      | variant | classes                   |
      | default | btn is-active             |
      | primary | btn btn-primary is-active |
      | danger  | btn btn-danger is-active  |
      | ghost   | btn btn-ghost is-active   |

  Scenario: glyph renders as a mono prefix before the label
    When I render a button "deploy" with glyph "▸"
    Then exactly 1 element matches "button.btn > span.btn-glyph"
    And the element "span.btn-glyph" has text "▸"
    And the first element child of "button.btn" is "span.btn-glyph"

  Scenario: kbd hint renders an inline chip after the label
    When I render a button "palette" with kbd hint "⌘K"
    Then exactly 1 element matches "button.btn > kbd.kbd"
    And the element "kbd.kbd" has text "⌘K"
    And the last element child of "button.btn" is "kbd.kbd"

  Scenario: a fully loaded button keeps glyph, label and kbd in order
    When I render a "primary" button "deploy" with glyph "▸" and kbd hint "⌘⏎"
    Then the element "button.btn" has classes "btn btn-primary"
    And the element "button.btn" has own text "deploy"
    And the first element child of "button.btn" is "span.btn-glyph"
    And the last element child of "button.btn" is "kbd.kbd"

  Scenario: standalone Kbd renders a semantic chip element
    When I render a kbd chip "esc"
    Then exactly 1 element matches "kbd.kbd"
    And the element "kbd.kbd" has text "esc"

  Scenario: button group fuses its buttons
    When I render a button group of buttons "1h,6h,24h" with active "6h"
    Then exactly 1 element matches "span.btn-group[role=group]"
    And exactly 3 elements match "span.btn-group > button.btn"
    And exactly 1 element matches "span.btn-group > button.btn.is-active"

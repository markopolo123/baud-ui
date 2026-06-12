Feature: Choice controls — checkbox, radio, segmented toggle
  A real <input> element sits underneath every choice control for forms and
  a11y; the text glyphs ([x] / (•)) and the accent segment fill are pure-CSS
  presentation driven by :checked. Labels are clickable because the label
  element wraps the input — no for/id wiring needed, no client logic at all.

  Scenario: checkbox is a real input wrapped in a clickable label
    When I render a checkbox named "telemetry" with value "on" and label "telemetry"
    Then exactly 1 element matches "label.cbx > input.cbx-input[type=checkbox][name=telemetry][value=on]"
    And exactly 1 element matches "label.cbx > span.cbx-box[aria-hidden=true]"
    And the element "label.cbx > span.cbx-label" has text "telemetry"
    And no element matches "input[checked]"
    And no element matches "input[disabled]"

  Scenario: checked checkbox carries the checked attribute on the real input
    When I render a checked checkbox named "auto-restart" with label "auto-restart"
    Then exactly 1 element matches "label.cbx > input[type=checkbox][checked]"
    And no element matches "input[disabled]"

  Scenario: disabled checkbox disables the real input
    When I render a disabled checkbox named "legacy" with label "legacy mode"
    Then exactly 1 element matches "label.cbx > input[type=checkbox][disabled]"
    And no element matches "input[checked]"

  Scenario: checkbox id lands on the input for external hooks
    When I render a checkbox with id "chk-1" named "telemetry"
    Then exactly 1 element matches "label.cbx > input[type=checkbox][id=chk-1]"

  Scenario: radio is a real radio input with the shared glyph presentation
    When I render a radio named "region" with value "eu-west" and label "eu-west"
    Then exactly 1 element matches "label.cbx > input.cbx-input[type=radio][name=region][value=eu-west]"
    And exactly 1 element matches "label.cbx > span.cbx-box[aria-hidden=true]"
    And the element "label.cbx > span.cbx-label" has text "eu-west"
    And no element matches "input[checked]"

  Scenario: radio group shares one name and has a single checked member
    When I render a radio group named "region" with options "eu-west,us-east,ap-south" selecting "us-east"
    Then exactly 3 elements match "label.cbx > input[type=radio][name=region]"
    And exactly 1 element matches "input[type=radio][checked]"
    And exactly 1 element matches "input[type=radio][checked][value=us-east]"

  Scenario: toggle renders every segment as a radio in one named group
    When I render a toggle named "view" with options "table,json,raw" selecting "json"
    Then exactly 1 element matches "span.tg"
    And exactly 3 elements match "span.tg > label.tg-opt > input.tg-input[type=radio][name=view]"
    And exactly 3 elements match "label.tg-opt > span.tg-opt-label"
    And exactly 1 element matches "label.tg-opt > input[checked]"
    And exactly 1 element matches "label.tg-opt > input[checked][value=json]"

  Scenario: toggle with no matching value selects nothing
    When I render a toggle named "view" with options "table,json,raw" selecting ""
    Then exactly 3 elements match "label.tg-opt > input[type=radio][name=view]"
    And no element matches "input[checked]"

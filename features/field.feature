Feature: Field and Input primitives
  Field stacks an UPPERCASE label, a control slot and a hint line; an error
  string flips the hint to "✗ message" and marks the input border. Input is
  a real <input> in a styled wrapper row with optional prefix/suffix affixes.
  Focus styling is pure CSS (asserted in e2e via computed styles).

  Scenario: plain input renders a real input element inside the wrap
    When I render an input with id "host" and placeholder "search hosts"
    Then exactly 1 element matches "span.input-wrap > input.input"
    And the element "input.input" has attribute "id" equal to "host"
    And the element "input.input" has attribute "placeholder" equal to "search hosts"
    And the element "input.input" has attribute "type" equal to "text"
    And no element matches "span.input-wrap > span.affix"
    And no element matches "span.input-wrap.err"

  Scenario: prefix affix renders before the input
    When I render an input with prefix "⌕"
    Then exactly 1 element matches "span.input-wrap > span.affix"
    And the input is immediately preceded by an affix with text "⌕"

  Scenario: suffix affix renders after the input
    When I render an input with suffix "ms"
    Then exactly 1 element matches "span.input-wrap > span.affix"
    And the input is immediately followed by an affix with text "ms"

  Scenario: prefix and suffix affixes render around the input
    When I render an input with prefix "host:" and suffix ".local"
    Then exactly 2 elements match "span.input-wrap > span.affix"
    And the input is immediately preceded by an affix with text "host:"
    And the input is immediately followed by an affix with text ".local"

  Scenario: disabled input carries the disabled attribute
    When I render a disabled input
    Then exactly 1 element matches "span.input-wrap > input.input[disabled]"

  Scenario: error input carries the error marker and aria-invalid
    When I render an input in the error state
    Then exactly 1 element matches "span.input-wrap.err > input.input"
    And the element "input.input" has attribute "aria-invalid" equal to "true"

  Scenario: field label is uppercase-styled and associated with its input
    When I render a field labelled "Hostname" for id "host" with hint "fqdn preferred"
    Then exactly 1 element matches "div.field > label.field-label[for=host]"
    And the element "label.field-label" has text "Hostname"
    And exactly 1 element matches "div.field span.input-wrap > input.input[id=host]"

  Scenario: field hint renders on the hint line
    When I render a field labelled "Hostname" for id "host" with hint "fqdn preferred"
    Then exactly 1 element matches "div.field > span.field-hint"
    And the element "span.field-hint" has text "fqdn preferred"
    And no element matches "span.field-hint.err"

  Scenario: field error flips the hint to ✗ message and marks the input
    When I render a field labelled "Hostname" for id "host" with error "host unreachable"
    Then exactly 1 element matches "div.field > span.field-hint.err"
    And the element "span.field-hint.err" has text "✗ host unreachable"
    And exactly 1 element matches "div.field span.input-wrap.err > input.input[id=host]"
    And exactly 1 element matches "div.field > span.field-hint"

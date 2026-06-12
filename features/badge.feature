Feature: Badge and Dot primitives
  Badges are dense status chips: Tone (ok/warn/err/info/accent/neutral) ×
  Variant (tint/solid/outline), optional 5px square dot. Dot is the 7px
  round status indicator with an optional reduced-motion-gated pulse.

  Scenario Outline: badge renders every tone × variant combination
    When I render a badge with tone "<tone>" and variant "<variant>" labelled "deploy"
    Then exactly 1 element matches "span.badge.<class>.tone-<tone>"
    And the element "span.badge" has text "deploy"
    And no element matches "span.badge > span.badge-dot"

    Examples:
      | tone    | variant | class      |
      | ok      | tint    | bd-tint    |
      | warn    | tint    | bd-tint    |
      | err     | tint    | bd-tint    |
      | info    | tint    | bd-tint    |
      | accent  | tint    | bd-tint    |
      | neutral | tint    | bd-tint    |
      | ok      | solid   | bd-solid   |
      | warn    | solid   | bd-solid   |
      | err     | solid   | bd-solid   |
      | info    | solid   | bd-solid   |
      | accent  | solid   | bd-solid   |
      | neutral | solid   | bd-solid   |
      | ok      | outline | bd-outline |
      | warn    | outline | bd-outline |
      | err     | outline | bd-outline |
      | info    | outline | bd-outline |
      | accent  | outline | bd-outline |
      | neutral | outline | bd-outline |

  Scenario: zero-value badge defaults to neutral tint
    When I render a default badge labelled "queued"
    Then exactly 1 element matches "span.badge.bd-tint.tone-neutral"
    And the element "span.badge" has text "queued"

  Scenario: badge dot is opt-in and renders before the label
    When I render a badge with tone "ok" and variant "tint" and a dot labelled "ready"
    Then exactly 1 element matches "span.badge.bd-tint.tone-ok > span.badge-dot"
    And the element "span.badge" has text "ready"

  Scenario Outline: dot renders each tone without pulse by default
    When I render a dot with tone "<tone>"
    Then exactly 1 element matches "span.dot.tone-<tone>"
    And no element matches "span.dot.pulse"

    Examples:
      | tone    |
      | ok      |
      | warn    |
      | err     |
      | info    |
      | accent  |
      | neutral |

  Scenario: zero-value dot defaults to ok tone
    When I render a default dot
    Then exactly 1 element matches "span.dot.tone-ok"

  Scenario: pulse flag adds the pulse class
    When I render a pulsing dot with tone "err"
    Then exactly 1 element matches "span.dot.tone-err.pulse"

Feature: Toasts — OOB notification stack
  Toasts are server-pushed notifications: any htmx response can carry an
  hx-swap-oob fragment that lands in the page-level #toasts region (a fixed
  bottom-right stack above the statusbar, rendered by Page on every
  document). Each toast is a role=status row — 3px left tone bar + tone
  glyph (✓ ✗ ▲ ℹ — text glyphs, not emoji) + title + optional body — that
  auto-dismisses after data-toast-ms milliseconds via the Toast hyperscript
  behavior (0 = sticky), with a manual ✕ dismiss button.

  Scenario: the page renders the toast region for OOB swaps
    When I render the default page
    Then exactly 1 element matches "div.toasts[id=toasts]"
    And the element "div.toasts" has attribute "aria-live" equal to "polite"

  Scenario Outline: each tone carries its tone class and glyph
    When I render a toast with tone "<tone>" titled "t" with body "b"
    Then exactly 1 element matches "div.toast.tone-<tone>[role=status]"
    And the element "span.toast-glyph" has text "<glyph>"
    And the element "span.toast-glyph" has attribute "aria-hidden" equal to "true"

    Examples:
      | tone | glyph |
      | ok   | ✓     |
      | err  | ✗     |
      | warn | ▲     |
      | info | ℹ     |

  Scenario: the default tone is info
    When I render a toast with no tone titled "cluster note"
    Then exactly 1 element matches "div.toast.tone-info[role=status]"
    And the element "span.toast-glyph" has text "ℹ"

  Scenario: title and body render as separate lines
    When I render a toast with tone "ok" titled "Deployed ingest-gw" with body "rollout complete on 12 pods"
    Then the element "div.toast > span.toast-msg > span.toast-title" has text "Deployed ingest-gw"
    And the element "div.toast > span.toast-msg > span.toast-body" has text "rollout complete on 12 pods"

  Scenario: the body line is optional
    When I render a toast with tone "warn" titled "p99 above SLO"
    Then exactly 1 element matches "span.toast-title"
    And no element matches "span.toast-body"

  Scenario: toasts install the Toast behavior with the ~4s default interval
    When I render a toast with tone "ok" titled "t"
    Then the element "div.toast" has attribute "_" equal to "install Toast"
    And the element "div.toast" has attribute "data-toast-ms" equal to "4000"

  Scenario: the auto-dismiss interval is configurable per toast
    When I render a toast with tone "ok" titled "t" dismissing after 600 ms
    Then the element "div.toast" has attribute "data-toast-ms" equal to "600"

  Scenario: sticky toasts opt out of auto-dismiss
    When I render a sticky toast with tone "err" titled "still broken"
    Then the element "div.toast" has attribute "data-toast-ms" equal to "0"
    And the element "div.toast" has attribute "_" equal to "install Toast"

  Scenario: every toast carries a manual dismiss button
    When I render a toast with tone "ok" titled "t"
    Then exactly 1 element matches "div.toast > button.x-btn[type=button]"
    And the element "button.x-btn" has attribute "aria-label" equal to "dismiss notification"
    And the element "button.x-btn" has text "✕"

  Scenario: the OOB wrapper targets the toast region with a beforeend swap
    When I render an OOB toast with tone "err" titled "Connection lost"
    Then exactly 1 element matches "div[hx-swap-oob=beforeend:#toasts]"
    And exactly 1 element matches "div[hx-swap-oob=beforeend:#toasts] > div.toast.tone-err[role=status]"
    And the element "span.toast-title" has text "Connection lost"

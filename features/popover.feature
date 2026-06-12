Feature: Popover and Tooltip — anchored quick-actions panel + pure-CSS tips
  Popover is an anchored 280px panel (--bg-panel, strong border, shadow)
  opened from a trigger button: the trigger toggles .open on the wrap with
  inline hyperscript and mirrors it onto aria-expanded; the shared
  MenuDismiss behavior closes it on outside click / Esc (one at a time —
  any other open popover loses .open when a new trigger is clicked).
  Tooltip is PURE CSS: a data-tip attribute rendered by ::after with
  white-space: pre (multi-line aligned mono tips) after a 150ms delay,
  arrow via ::before, shown on hover AND :focus-visible. No hyperscript,
  no element — baud.Tip/baud.TipUnder are attribute helpers.

  # ---- Popover: structure & ARIA ------------------------------------------

  Scenario: popover renders closed with full ARIA wiring
    When I render a popover with id "pop" labelled "Actions"
    Then exactly 1 element matches "span.pop-wrap[id=pop]"
    And no element matches "span.pop-wrap.open"
    And exactly 1 element matches "button.pop-trigger.btn[type=button]"
    And no element matches "button.pop-trigger[aria-haspopup]"
    And the element "button.pop-trigger" has attribute "aria-expanded" equal to "false"
    And the element "button.pop-trigger" has attribute "aria-controls" equal to "pop-panel"
    And the element "button.pop-trigger" has text "Actions"
    And exactly 1 element matches "span.pop-wrap > div.popover[id=pop-panel]"

  Scenario: popover wires the dismiss behavior and the local toggle
    When I render a popover with id "pop" labelled "Actions"
    Then the popover wrap installs the MenuDismiss behavior
    And the popover wrap syncs aria-expanded when MenuDismiss closes it
    And the popover trigger toggles the open class with aria-expanded in step

  Scenario: popover children render inside the anchored panel
    When I render a popover with id "pop" labelled "Actions" containing "Flush cache"
    Then the element "div.popover" has text "Flush cache"

  Scenario: popover title renders as an uppercase field-label header
    When I render a popover with id "pop" labelled "Actions" titled "Quick actions"
    Then exactly 1 element matches "div.popover > span.field-label"
    And the element "span.field-label" has text "Quick actions"

  Scenario: popover without a title renders no header
    When I render a popover with id "pop" labelled "Actions"
    Then no element matches "span.field-label"

  Scenario: popover trigger carries the mono glyph prefix
    When I render a popover with id "pop" labelled "Actions" with glyph "▾"
    Then exactly 1 element matches "button.pop-trigger > span.btn-glyph"
    And the element "span.btn-glyph" has text "▾"

  Scenario: server-rendered open popover marks wrap and trigger expanded
    When I render an open popover with id "pop" labelled "Actions"
    Then the element "span.pop-wrap" has classes "pop-wrap open"
    And the element "button.pop-trigger" has attribute "aria-expanded" equal to "true"

  Scenario: disabled popover disables the trigger button
    When I render a disabled popover with id "pop" labelled "Actions"
    Then exactly 1 element matches "button.pop-trigger[disabled]"

  # ---- Tooltip: pure-CSS data-tip attribute helpers ------------------------

  Scenario: Tip renders the data-tip attribute on its host
    When I render a tooltip host with tip "works on any element"
    Then the element "span[data-tip]" has attribute "data-tip" equal to "works on any element"
    And no element matches "span.tip-under"

  Scenario: multi-line tips survive into the attribute verbatim
    When I render a tooltip host with tip "p99 = 312ms\np95 = 104ms\np50 =  22ms"
    Then the tooltip host preserves the multi-line tip "p99 = 312ms\np95 = 104ms\np50 =  22ms"

  Scenario: TipUnder adds the dotted-underline variant class
    When I render an under-styled tooltip host with tip "38.2% of 30d budget consumed"
    Then the element "span[data-tip]" has classes "tip-under"
    And the element "span[data-tip]" has attribute "data-tip" equal to "38.2% of 30d budget consumed"

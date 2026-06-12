Feature: Overlays — Modal & Drawer
  Modal is a centered dialog (540px at 9vh, strong border + deep shadow)
  inside a fixed .overlay backdrop; Drawer is the right-side variant (420px,
  full height) whose backdrop adds .drawer-wrap. Both carry role=dialog,
  aria-modal and aria-labelledby wired to the header title, a ✕ close glyph
  button, and install the Overlay hyperscript behavior (focus trap, Esc +
  backdrop + [data-overlay-close] dismiss, opener focus restore, one overlay
  at a time). htmx-injected overlays (hx-target="body" hx-swap="beforeend")
  are removed from the DOM on close; the Static variant stays in the DOM
  and toggles .open instead (send baud:overlayOpen to it).

  Scenario: a modal renders dialog chrome inside the backdrop
    When I render a modal with id "confirm" titled "kill pod"
    Then exactly 1 element matches "div.overlay > div.modal[role=dialog][aria-modal=true]"
    And the element "div.overlay" has attribute "id" equal to "confirm"
    And the element "div.overlay" has attribute "_" equal to "install Overlay"
    And no element matches "div.overlay.drawer-wrap"

  Scenario: the modal title labels the dialog for assistive tech
    When I render a modal with id "confirm" titled "kill pod"
    Then the element "div.modal" has attribute "aria-labelledby" equal to "confirm-title"
    And the element "header.modal-hd > span.modal-title" has attribute "id" equal to "confirm-title"
    And the element "span.modal-title" has text "kill pod"

  Scenario: the header carries the ✕ close glyph button
    When I render a modal with id "confirm" titled "kill pod"
    Then exactly 1 element matches "header.modal-hd > button.x-btn[type=button][data-overlay-close]"
    And the element "button.x-btn" has attribute "aria-label" equal to "close"
    And the element "button.x-btn" has text "✕"

  Scenario: the body slot renders the children
    When I render a modal with id "confirm" titled "kill pod"
    Then the element "div.modal > div.modal-bd" has text "body content"

  Scenario: without a footer no actions strip renders
    When I render a modal with id "confirm" titled "kill pod"
    Then no element matches "footer.modal-ft"

  Scenario: the footer slot renders the right-aligned actions strip
    When I render a modal with footer actions with id "confirm" titled "kill pod"
    Then exactly 1 element matches "div.modal > footer.modal-ft"
    And the element "footer.modal-ft" has text "apply"

  Scenario: an htmx-injected modal is not the static variant
    When I render a modal with id "confirm" titled "kill pod"
    Then no element matches ".overlay-static"

  Scenario: a static modal stays in the DOM and opens by class toggle
    When I render a static modal with id "sheet-modal" titled "deploy fleet"
    Then the element "div.overlay" has classes "overlay overlay-static"
    And exactly 1 element matches "div.overlay > div.modal[role=dialog]"

  Scenario: a drawer renders the right-side variant
    When I render a drawer with id "svc" titled "api-core"
    Then exactly 1 element matches "div.overlay.drawer-wrap > div.drawer[role=dialog][aria-modal=true]"
    And the element "div.overlay" has attribute "_" equal to "install Overlay"
    And no element matches "div.modal"

  Scenario: the drawer shares the header chrome and ARIA wiring
    When I render a drawer with id "svc" titled "api-core"
    Then the element "div.drawer" has attribute "aria-labelledby" equal to "svc-title"
    And the element "header.modal-hd > span.modal-title" has attribute "id" equal to "svc-title"
    And the element "span.modal-title" has text "api-core"
    And exactly 1 element matches "header.modal-hd > button.x-btn[data-overlay-close]"
    And the element "button.x-btn" has attribute "aria-label" equal to "close"

  Scenario: the drawer body is the scrollable slot and has no footer
    When I render a drawer with id "svc" titled "api-core"
    Then the element "div.drawer > div.drawer-bd" has text "body content"
    And no element matches "footer.modal-ft"

  Scenario: a static drawer stays in the DOM and opens by class toggle
    When I render a static drawer with id "sheet-drawer" titled "service"
    Then the element "div.overlay" has classes "overlay drawer-wrap overlay-static"

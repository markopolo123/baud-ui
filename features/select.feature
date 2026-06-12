Feature: Select and Combobox — native baseline, custom menu, type-to-filter
  Select is htmx-friendly: the baseline is a REAL <select> styled to match
  (works with JS disabled); SelectMenu is the progressive enhancement — a
  custom trigger + absolutely-positioned listbox menu driven by hyperscript
  (MenuDismiss + SelectKeys + SelectPick behaviors), with a hidden input
  underneath for forms. Combobox is an input with a ⌕ affix filtering the
  same menu, client-side (ComboboxFilter) or via a debounced hx-get
  round-trip, with the matched substring highlighted accent bold.

  # ---- Select: the real <select> baseline --------------------------------

  Scenario: select is a real native element with the full option list
    When I render a select named "region" with options "eu-west,us-east,ap-south"
    Then exactly 1 element matches "span.select > select.select-native[name=region]"
    And exactly 3 elements match "select.select-native > option"
    And exactly 1 element matches "span.select > span.select-chev[aria-hidden=true]"
    And no element matches "option[selected]"
    And no element matches "select[disabled]"

  Scenario: selecting a value marks exactly one option selected
    When I render a select named "region" with options "eu-west,us-east,ap-south" selecting "us-east"
    Then exactly 1 element matches "option[selected]"
    And exactly 1 element matches "option[selected][value=us-east]"
    And the element "option[selected]" has text "us-east"

  Scenario: disabled select disables the real element
    When I render a disabled select named "region" with options "eu-west,us-east"
    Then exactly 1 element matches "select.select-native[disabled]"

  Scenario: a disabled option disables only that option
    When I render a select named "region" with options "eu-west,us-east,ap-south" disabling "ap-south"
    Then exactly 1 element matches "option[disabled]"
    And exactly 1 element matches "option[disabled][value=ap-south]"

  Scenario: select id lands on the real element for label association
    When I render a select with id "sel-1" named "region" and options "eu-west,us-east"
    Then exactly 1 element matches "select.select-native[id=sel-1]"

  Scenario: select width sizes the wrapper inline
    When I render a select named "region" with options "eu-west,us-east" sized "18ch"
    Then the element "span.select" has attribute "style" equal to "width: 18ch;"

  # ---- SelectMenu: custom trigger + listbox enhancement -------------------

  Scenario: select menu renders the ARIA listbox wiring closed
    When I render a select menu with id "env" named "env" and options "dev,staging,prod" selecting "prod"
    Then exactly 1 element matches "span.select[id=env]"
    And exactly 1 element matches "button.select-trigger[type=button][aria-haspopup=listbox]"
    And the element "button.select-trigger" has attribute "aria-expanded" equal to "false"
    And the element "button.select-trigger" has attribute "aria-controls" equal to "env-menu"
    And exactly 1 element matches "div.menu[role=listbox][id=env-menu]"
    And exactly 3 elements match "div.menu > button.menu-item[role=option][type=button][tabindex=-1]"
    And exactly 1 element matches "button.select-trigger > span.select-chev[aria-hidden=true]"

  Scenario: select menu installs the dismiss, keyboard and pick behaviors
    When I render a select menu with id "env" named "env" and options "dev,staging,prod" selecting "prod"
    Then the element "span.select" has attribute "_" equal to "install MenuDismiss install SelectKeys install SelectPick"

  Scenario: select menu carries a hidden input so forms still submit
    When I render a select menu with id "env" named "env" and options "dev,staging,prod" selecting "prod"
    Then exactly 1 element matches "span.select > input[type=hidden][name=env][value=prod]"

  Scenario: the selected option is active, checked and aria-selected
    When I render a select menu with id "env" named "env" and options "dev,staging,prod" selecting "prod"
    Then the element "button.select-trigger > span.select-value" has text "prod"
    And exactly 1 element matches "button.menu-item[aria-selected=true]"
    And exactly 1 element matches "button.menu-item.is-active[data-value=prod][aria-selected=true]"
    And exactly 2 elements match "button.menu-item[aria-selected=false]"
    And exactly 1 element matches "button.menu-item.is-active > span.menu-side > span.menu-check"
    And each option has an id derived from "env"

  Scenario: select menu without a selection shows the placeholder
    When I render a select menu with id "env" and options "dev,staging,prod" with placeholder "pick env"
    Then the element "button.select-trigger > span.select-value" has text "pick env"
    And no element matches "button.menu-item[aria-selected=true]"
    And no element matches "input[type=hidden]"

  Scenario: disabled select menu disables the trigger
    When I render a disabled select menu with id "env" named "env" and options "dev,staging"
    Then exactly 1 element matches "button.select-trigger[disabled]"

  Scenario: right-aligned select menu
    When I render a right-aligned select menu with id "env" named "env" and options "dev,staging"
    Then the element "div.menu" has classes "menu menu-right"

  # ---- Combobox: ⌕ input + type-to-filter menu ---------------------------

  Scenario: combobox renders the ARIA combobox wiring closed
    When I render a combobox with id "host" named "q" and options "ingest-gw-1,ingest-gw-2,api-core"
    Then exactly 1 element matches "span.select.combobox[id=host]"
    And exactly 1 element matches "span.input-wrap > input.input[role=combobox][autocomplete=off]"
    And the element "input.input" has attribute "aria-expanded" equal to "false"
    And the element "input.input" has attribute "aria-controls" equal to "host-menu"
    And the element "input.input" has attribute "aria-autocomplete" equal to "list"
    And exactly 1 element matches "div.menu[role=listbox][id=host-menu]"
    And exactly 3 elements match "button.menu-item[role=option][data-value]"
    And the input is immediately preceded by an affix with text "⌕"
    And exactly 1 element matches "span.input-wrap > span.select-chev[aria-hidden=true]"
    And exactly 1 element matches "div.menu > div.menu-empty"

  Scenario: client-mode combobox installs the local filter behavior
    When I render a combobox with id "host" named "q" and options "ingest-gw-1,api-core"
    Then the element "span.select" has attribute "_" equal to "install MenuDismiss install SelectKeys install SelectPick install ComboboxFilter"
    And no element matches "input[hx-get]"

  Scenario: server-mode combobox swaps the menu via a debounced hx-get
    When I render a server combobox with id "host" named "q" searching "/api/combobox"
    Then the element "input.input" has attribute "hx-get" equal to "/api/combobox"
    And the element "input.input" has attribute "hx-trigger" equal to "input changed delay:200ms"
    And the element "input.input" has attribute "hx-target" equal to "#host-menu"
    And the element "input.input" has attribute "hx-swap" equal to "innerHTML"
    And the element "span.select" has attribute "_" equal to "install MenuDismiss install SelectKeys install SelectPick"

  Scenario: combobox options carry right-aligned meta when given
    When I render a combobox with id "host" named "q" and options "ingest-gw-1=eu-west,api-core=us-east,scratch"
    Then exactly 2 elements match "button.menu-item > span.menu-side > span.cb-meta"
    And exactly 1 element matches "button.menu-item[data-value=ingest-gw-1][data-meta=eu-west]"
    And the element "button.menu-item[data-value=api-core] > span.menu-side > span.cb-meta" has text "us-east"
    And no element matches "button.menu-item[data-value=scratch][data-meta]"

  Scenario: combobox with a value pre-fills the input and marks the option
    When I render a combobox with id "host" named "q" and options "ingest-gw-1,api-core" selecting "api-core"
    Then the element "input.input" has attribute "value" equal to "api-core"
    And exactly 1 element matches "button.menu-item.is-active[data-value=api-core][aria-selected=true]"

  Scenario: server-filtered option fragment highlights the matched substring
    When I render combobox options for id "host" matching "gw" from "ingest-gw-1=eu-west,api-core"
    Then exactly 1 element matches "button.menu-item > span.cb-label > span.cb-match"
    And the element "span.cb-match" has text "gw"
    And the element "button.menu-item[data-value=ingest-gw-1] > span.cb-label" has text "ingest-gw-1"
    And exactly 1 element matches "div.menu-empty"

  Scenario: option fragment for an empty query carries no highlight marks
    When I render combobox options for id "host" matching "" from "ingest-gw-1,api-core"
    Then exactly 2 elements match "button.menu-item > span.cb-label"
    And no element matches "span.cb-match"

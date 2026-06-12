Feature: App shell and layout foundation
  The shell frame, page document wrapper and pane tiling are the layout
  contract every later component builds on.

  Scenario: shell with nav renders all four regions
    When I render a shell with a nav
    Then exactly 1 element matches "div[data-shell] > header[data-topbar]"
    And exactly 1 element matches "div[data-shell] > aside[data-nav]"
    And exactly 1 element matches "div[data-shell] > main"
    And exactly 1 element matches "div[data-shell] > footer[data-statusbar]"

  Scenario: shell without nav is the single-column variant
    When I render a shell without a nav
    Then no element matches "div[data-shell] > aside[data-nav]"
    And exactly 1 element matches "div[data-shell] > main"
    And exactly 1 element matches "div[data-shell] > footer[data-statusbar]"

  Scenario: page root carries the default mode classes
    When I render the default page
    Then exactly 1 element matches "body"
    And the element "body" has classes "t-gruvbox d-dense b-line f-mono"
    And the element "body" has attribute "_" equal to "install PaletteKey"

  Scenario: page loads the behaviors file before the hyperscript library
    When I render the default page
    Then exactly 1 element matches "head > script[type=text/hyperscript]"
    And the behaviors script comes before the hyperscript library script

  Scenario: panes wrapper emits its grid template and behavior install
    When I render panes "42ch 1fr" with 2 panes
    Then exactly 1 element matches "div[data-panes]"
    And the element "div[data-panes]" has attribute "data-panes" equal to "42ch 1fr"
    And the element "div[data-panes]" has attribute "_" equal to "install Panes"
    And no element matches "div[data-panes] > .split-gutter"

  Scenario: row panes wrapper emits the rows template
    When I render row panes "1.6fr 1fr" with 2 panes
    Then exactly 1 element matches "div[data-panes-rows]"
    And the element "div[data-panes-rows]" has attribute "data-panes-rows" equal to "1.6fr 1fr"
    And the element "div[data-panes-rows]" has attribute "_" equal to "install Panes"

  Scenario: resizable panes render server-side gutters and a widened template
    When I render resizable panes "42ch 1fr" with id "demo" and 2 panes
    Then exactly 1 element matches "div[data-panes] > .split-gutter"
    And the element "div[data-panes]" has attribute "data-panes" equal to "42ch 7px 1fr"
    And the element "div[data-panes]" has attribute "data-panes-id" equal to "demo"
    And the element "div[data-panes]" has attribute "_" equal to "install Panes install Resizable"

  Scenario: resizable panes keep functional tracks intact when widening
    When I render resizable panes "minmax(0, 1fr) 1fr" with id "mm" and 2 panes
    Then the element "div[data-panes]" has attribute "data-panes" equal to "minmax(0, 1fr) 7px 1fr"

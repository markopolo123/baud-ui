Feature: Panel, StatusBar and Toolbar structure components
  Panel is the universal container: a data-panel section with a header strip
  (UPPERCASE --fs-sm muted title, right-aligned actions slot) and a
  scrollable body that every later wave nests content into. StatusBar is the
  full-width vim-style cell strip (mode / spring / plain cells separated by
  hairlines). Toolbar is the thin hstack wrapper composing existing
  controls. Contract per design/README.md "Structure".

  Scenario: panel renders header, title and body
    When I render a panel titled "log tail"
    Then exactly 1 element matches "section[data-panel]"
    And the element "section[data-panel]" has attribute "aria-label" equal to "log tail"
    And exactly 1 element matches "section[data-panel] > header.panel-hd"
    And exactly 1 element matches "header.panel-hd > span.panel-title"
    And the element "span.panel-title" has text "log tail"
    And exactly 1 element matches "section[data-panel] > div.panel-bd"
    And the element "div.panel-bd" has text "body"

  Scenario: panel without actions renders no actions slot
    When I render a panel titled "navigator"
    Then no element matches "span.panel-acts"

  Scenario: panel actions slot sits right-aligned in the header
    When I render a panel titled "navigator" with a kbd action "j/k"
    Then exactly 1 element matches "header.panel-hd > span.panel-acts"
    And exactly 1 element matches "span.panel-acts > kbd.kbd"
    And the element "kbd.kbd" has text "j/k"
    And the first element child of "header.panel-hd" is "span.panel-title"
    And the last element child of "header.panel-hd" is "span.panel-acts"

  Scenario: panel body nests arbitrary children
    When I render a panel titled "metrics" with 3 body blocks
    Then exactly 3 elements match "div.panel-bd > span"

  Scenario: panel accepts composition class and id hooks
    When I render a panel titled "fleet" with id "fleet-panel" and class "fill"
    Then the element "section[data-panel]" has classes "fill"
    And the element "section[data-panel]" has attribute "id" equal to "fleet-panel"

  Scenario: status bar renders one separated cell per entry
    When I render a status bar with cells "host fleet-a,zone eu-west,uptime 14d"
    Then exactly 1 element matches "div.statusbar"
    And exactly 3 elements match "div.statusbar > span.sb-cell"
    And no element matches "span.sb-mode"
    And no element matches "span.sb-spring"

  Scenario: leading mode cell carries the vim-style marker
    When I render a status bar with mode "NORMAL" and cells "fleet.go,utf-8"
    Then exactly 3 elements match "div.statusbar > span.sb-cell"
    And exactly 1 element matches "div.statusbar > span.sb-cell.sb-mode"
    And the element "span.sb-cell.sb-mode" has text "NORMAL"
    And the first element child of "div.statusbar" is "span.sb-cell.sb-mode"
    And no element matches "span.sb-spring"

  Scenario: spring cell flexes between the cell groups
    When I render a status bar with cells "NORMAL,main" then a spring then cells "utf-8,ln 1:1"
    Then exactly 5 elements match "div.statusbar > span.sb-cell"
    And exactly 1 element matches "div.statusbar > span.sb-cell.sb-spring"
    And the element "span.sb-cell.sb-spring" has text ""
    And the first element child of "div.statusbar" is "span.sb-cell"
    And the last element child of "div.statusbar" is "span.sb-cell"

  Scenario: toolbar preserves its children order
    When I render a toolbar of a button "deploy" and an input
    Then exactly 1 element matches "div.toolbar"
    And no element matches "div.toolbar[role]"
    And exactly 1 element matches "div.toolbar > button.btn"
    And exactly 1 element matches "div.toolbar > span.input-wrap"
    And the first element child of "div.toolbar" is "button.btn"
    And the last element child of "div.toolbar" is "span.input-wrap"

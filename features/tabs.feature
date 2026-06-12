Feature: Tabs — underline & boxed variants, htmx & local switching
  A Tabs strip is a real ARIA tabs widget: role=tablist over <button role=tab>
  children with aria-selected and roving tabindex, each tab pointing at its
  tabpanel via aria-controls. Switching is either a server round-trip (every
  tab hx-gets its pane into one shared target panel) or purely local (the Tabs
  hyperscript behavior swaps pre-rendered panels). Count badges reuse the
  Badge primitive.

  Scenario: underline variant renders an ARIA tablist of tab buttons
    When I render underline tabs "pods,events,logs" with active index 0
    Then exactly 1 element matches "div.tabs.tabs-underline[role=tablist]"
    And exactly 3 elements match "div.tabs > button.tab[role=tab][type=button]"
    And no element matches "div.tabs-boxed"

  Scenario: boxed variant renders the boxed marker class
    When I render boxed tabs "5m,1h,1d,1w" with active index 0
    Then exactly 1 element matches "div.tabs.tabs-boxed[role=tablist]"
    And exactly 4 elements match "div.tabs-boxed > button.tab[role=tab]"
    And no element matches "div.tabs-underline"

  Scenario: the active tab is marked and selected, the rest are not
    When I render underline tabs "pods,events,logs" with active index 1
    Then exactly 1 element matches "button.tab.is-active"
    And the element "button.tab.is-active" has text "events"
    And the element "button.tab.is-active" has attribute "aria-selected" equal to "true"
    And exactly 2 elements match "button.tab[aria-selected=false]"
    And no element matches "button.tab.is-active[aria-selected=false]"

  Scenario: roving tabindex keeps only the active tab in the tab order
    When I render underline tabs "pods,events,logs" with active index 2
    Then exactly 1 element matches "button.tab[tabindex=0]"
    And the element "button.tab[tabindex=0]" has text "logs"
    And exactly 2 elements match "button.tab[tabindex=-1]"

  Scenario: count badges render through the Badge primitive
    When I render underline tabs with counts "pods=39,events=7,logs"
    Then exactly 2 elements match "button.tab > span.badge"
    And the element "button.tab.is-active > span.badge" has text "39"
    And the element "button.tab[aria-selected=false] > span.badge" has text "7"

  Scenario: boxed tabs carry count badges too
    When I render boxed tabs with counts "errs=3,warns=12"
    Then exactly 2 elements match "div.tabs-boxed > button.tab > span.badge"

  Scenario: htmx mode wires every tab to the shared target panel
    When I render htmx tabs "5m=/demo/tabs?range=5m,1h=/demo/tabs?range=1h" targeting "range-pane"
    Then exactly 2 elements match "button.tab[hx-get][hx-target=#range-pane]"
    And the element "button.tab.is-active" has attribute "hx-get" equal to "/demo/tabs?range=5m"
    And exactly 2 elements match "button.tab[aria-controls=range-pane]"
    And the element "div.tabs" has attribute "_" equal to "install Tabs"
    And no element matches "div.tabs[data-tabs-local]"

  Scenario: htmx tabs with nav hrefs degrade to real links
    When I render htmx nav tabs "5m=/demo/tabs?range=5m|/?tab=5m,1h=/demo/tabs?range=1h|/?tab=1h" targeting "range-pane"
    Then exactly 2 elements match "div.tabs > a.tab[role=tab]"
    And no element matches "button.tab"
    And the element "a.tab.is-active" has attribute "href" equal to "/?tab=5m"
    And the element "a.tab.is-active" has attribute "hx-get" equal to "/demo/tabs?range=5m"
    And exactly 2 elements match "a.tab[hx-target=#range-pane]"
    And exactly 2 elements match "a.tab[aria-controls=range-pane]"
    And exactly 2 elements match "a.tab[tabindex=0]"
    And no element matches "a.tab[tabindex=-1]"

  Scenario: local mode wires tabs to their pre-rendered panels via hyperscript
    When I render local tabs "pods=pane-pods,events=pane-events" with active index 0
    Then the element "div.tabs" has attribute "data-tabs-local" equal to "true"
    And the element "div.tabs" has attribute "_" equal to "install Tabs"
    And the element "button.tab.is-active" has attribute "aria-controls" equal to "pane-pods"
    And exactly 1 element matches "button.tab[aria-controls=pane-events][aria-selected=false]"
    And no element matches "button.tab[hx-get]"

  Scenario: tab buttons derive stable ids from the strip id
    When I render local tabs "pods=pane-pods,events=pane-events" with active index 0
    Then the element "div.tabs" has attribute "id" equal to "demo-tabs"
    And exactly 1 element matches "button.tab[id=demo-tabs-tab-0][aria-selected=true]"
    And exactly 1 element matches "button.tab[id=demo-tabs-tab-1][aria-selected=false]"

  Scenario: an active tab panel is a visible ARIA tabpanel labelled by its tab
    When I render an active tab panel "pane-pods" labelled by "demo-tabs-tab-0"
    Then exactly 1 element matches "div.tab-panel[role=tabpanel][id=pane-pods][tabindex=0]"
    And the element "div.tab-panel" has attribute "aria-labelledby" equal to "demo-tabs-tab-0"
    And no element matches "div.tab-panel[hidden]"

  Scenario: an inactive tab panel is hidden until its tab activates
    When I render an inactive tab panel "pane-events" labelled by "demo-tabs-tab-1"
    Then exactly 1 element matches "div.tab-panel[role=tabpanel][id=pane-events][hidden]"
    And the element "div.tab-panel" has attribute "aria-labelledby" equal to "demo-tabs-tab-1"

  Scenario: a tab panel without a tab id renders no aria-labelledby
    When I render an active tab panel "pane-free"
    Then exactly 1 element matches "div.tab-panel[role=tabpanel][id=pane-free][tabindex=0]"
    And no element matches "div.tab-panel[aria-labelledby]"

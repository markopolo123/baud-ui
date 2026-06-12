Feature: DataTable — sticky-header data grid with htmx column sort
  The flagship data component (design/README.md "Data"): a semantic table
  with a sticky UPPERCASE header, --row rows, a ▌-marked selected row,
  right-aligned numeric columns and opt-in zebra/lines variants. Sort is a
  server round-trip: sortable headers hx-get ?sort=<key>&dir=<dir> swapping
  the tbody fragment, the thead re-rendering out-of-band so the accent ▲/▼
  indicator and flipped URLs track the state. A cell-tone hook tone-codes
  threshold cells (e.g. cpu > 90 ⇒ err).

  The step fixture: columns host (sortable), region, cpu (numeric,
  sortable), mem (numeric); rows alpha/bravo/charlie keyed h1/h2/h3 with
  cpu 12.5 / 75.2 / 96.7 and mem 40.0 / 60.1 / 88.9.

  Scenario: a data table renders a semantic table with header and body
    When I render a data table
    Then exactly 1 element matches "table.dt"
    And exactly 1 element matches "table.dt > thead > tr"
    And exactly 4 elements match "thead > tr > th[scope=col][data-key]"
    And exactly 1 element matches "thead > tr > th.row-mark"
    And exactly 1 element matches "table.dt > tbody[id=dt-fleet-body]"
    And exactly 3 elements match "tbody > tr"
    And no element matches "thead[hx-swap-oob]"

  Scenario: column labels render in the header cells
    When I render a data table
    Then the element "th[data-key=host]" has text "Host"
    And the element "th[data-key=cpu]" has text "CPU%"

  Scenario: numeric columns carry the right-align num marker
    When I render a data table
    Then exactly 2 elements match "thead > tr > th.num"
    And exactly 1 element matches "th.num[data-key=cpu]"
    And exactly 1 element matches "th.num[data-key=mem]"
    And exactly 6 elements match "tbody > tr > td.num"
    And no element matches "th.num[data-key=host]"
    And no element matches "th.num[data-key=region]"

  Scenario: the default table has neither zebra nor lines
    When I render a data table
    Then no element matches "table.dt-zebra"
    And no element matches "table.dt-lines"

  Scenario: zebra is an opt-in variant class
    When I render a zebra data table
    Then exactly 1 element matches "table.dt.dt-zebra"
    And no element matches "table.dt-lines"

  Scenario: lines is an opt-in variant class
    When I render a lines data table
    Then exactly 1 element matches "table.dt.dt-lines"
    And no element matches "table.dt-zebra"

  Scenario: the selected row carries the sel marker and the ▌ row mark
    When I render a data table with row "h2" selected
    Then exactly 1 element matches "tbody > tr.sel"
    And the element "tr.sel > td.row-mark" has text "▌"
    And exactly 3 elements match "tbody > tr > td.row-mark"
    And the ▌ row mark appears on selected rows only

  Scenario: no row is selected by default
    When I render a data table
    Then no element matches "tbody > tr.sel"
    And the ▌ row mark appears on selected rows only

  Scenario: sortable headers hx-get the body fragment swap
    When I render a data table sorted by "cpu" "asc"
    Then exactly 2 elements match "th[hx-get]"
    And the element "th[data-key=host]" has attribute "hx-get" equal to "/demo/datatable?sort=host&dir=asc"
    And the element "th[data-key=cpu]" has attribute "hx-target" equal to "#dt-fleet-body"
    And the element "th[data-key=cpu]" has attribute "hx-swap" equal to "outerHTML"
    And the element "th[data-key=cpu]" has attribute "tabindex" equal to "0"
    And no element matches "th[data-key=region][hx-get]"
    And no element matches "th[data-key=mem][hx-get]"

  Scenario: the active ascending column flips its URL to desc
    When I render a data table sorted by "cpu" "asc"
    Then the element "th[data-key=cpu]" has attribute "hx-get" equal to "/demo/datatable?sort=cpu&dir=desc"
    And the element "th[data-key=cpu]" has attribute "aria-sort" equal to "ascending"

  Scenario: the active descending column flips its URL back to asc
    When I render a data table sorted by "cpu" "desc"
    Then the element "th[data-key=cpu]" has attribute "hx-get" equal to "/demo/datatable?sort=cpu&dir=asc"
    And the element "th[data-key=cpu]" has attribute "aria-sort" equal to "descending"
    And the element "th[data-key=host]" has attribute "hx-get" equal to "/demo/datatable?sort=host&dir=asc"

  Scenario: the ▲ indicator renders on the ascending column only
    When I render a data table sorted by "cpu" "asc"
    Then exactly 1 element matches "th.sorted"
    And exactly 1 element matches "th.sorted[data-key=cpu] > span.sort-arrow"
    And the element "span.sort-arrow" has text "▲"

  Scenario: the ▼ indicator renders on the descending column
    When I render a data table sorted by "cpu" "desc"
    Then exactly 1 element matches "th.sorted[data-key=cpu] > span.sort-arrow"
    And the element "span.sort-arrow" has text "▼"

  Scenario: an unsorted table renders no indicator and no aria-sort
    When I render a data table
    Then no element matches "th.sorted"
    And no element matches "span.sort-arrow"
    And no element matches "th[aria-sort]"

  Scenario: the cell-tone hook tone-codes threshold cells per column
    When I render a data table with a cpu threshold tone hook
    Then exactly 1 element matches "tbody > tr > td.tone-err"
    And the element "td.tone-err" has text "96.7"
    And exactly 1 element matches "tbody > tr > td.tone-warn"
    And the element "td.tone-warn" has text "75.2"

  Scenario: the tbody fragment renders standalone for the htmx swap
    When I render only the data table body
    Then exactly 1 element matches "tbody[id=dt-fleet-body]"
    And exactly 3 elements match "tbody > tr"
    And exactly 3 elements match "tbody > tr > td.row-mark"

  Scenario: the thead fragment re-renders out-of-band beside the body
    When I render the data table head as an out-of-band fragment sorted by "cpu" "asc"
    Then exactly 1 element matches "thead[id=dt-fleet-head][hx-swap-oob=outerHTML]"
    And exactly 1 element matches "th.sorted[data-key=cpu]"

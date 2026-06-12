Feature: DatePicker — trigger + Monday-first month grid with server-side navigation
  The trigger shows the selected date as YYYY-MM-DD; a real hidden input
  carries the value for forms. The menu is a Monday-first 6×7 grid (always
  42 cells) computed on the server: the «‹›» buttons are hx-get round-trips
  re-rendering the exported menu fragment for the target month — no client
  date math. Today is outlined, the selected day fills accent, out-month
  days are dimmed, and a preset row offers today / -1d / -7d / -30d with
  server-computed dates.

  Scenario: trigger shows the selected date and a hidden input carries the value
    When I render a date picker named "due" selecting "2026-06-12" with today "2026-06-12" via "/api/datepicker"
    Then exactly 1 element matches "span.dp > button.dp-trigger[type=button][aria-haspopup=dialog][aria-expanded=false]"
    And the element "button.dp-trigger > span.dp-value" has text "2026-06-12"
    And exactly 1 element matches "span.dp > input.dp-input[type=hidden][name=due][value=2026-06-12]"
    And exactly 1 element matches "span.dp > div.dp-menu[role=dialog]"

  Scenario: the closed picker embeds the menu fragment for the selected month
    When I render a date picker named "due" selecting "2026-06-12" with today "2026-06-12" via "/api/datepicker"
    Then exactly 1 element matches "div.dp-menu > div.dp-hd"
    And exactly 42 elements match "div.dp-menu > div.dp-grid > button.dp-day"
    And the element "span.dp-title" has text "Jun 2026"

  Scenario: unselected picker prompts and submits an empty value
    When I render a date picker named "until" with no selection and today "2026-06-12" via "/api/datepicker"
    Then the element "button.dp-trigger > span.dp-value" has text "pick date"
    And the element "input.dp-input" has attribute "value" equal to ""
    And no element matches "button.dp-day.sel"

  Scenario: June 2026 starts on a Monday — no leading out-month cells
    When I render the date picker menu for month "2026-06" selecting "2026-06-12" with today "2026-06-12" via "/api/datepicker"
    Then exactly 7 elements match "div.dp-grid > span.dp-dow"
    And the day-of-week header reads "Mo Tu We Th Fr Sa Su"
    And exactly 42 elements match "div.dp-grid > button.dp-day[type=button]"
    And the day cells run from "2026-06-01" to "2026-07-12"
    And exactly 12 elements match "button.dp-day.out"
    And the element "button.dp-day[data-date=2026-06-01]" has text "1"
    And the element "button.dp-day[data-date=2026-07-12]" has text "12"

  Scenario: February 2026 (non-leap, starts Sunday) pads six leading and eight trailing cells
    When I render the date picker menu for month "2026-02" with today "2026-06-12" via "/api/datepicker"
    Then exactly 42 elements match "div.dp-grid > button.dp-day"
    And the day cells run from "2026-01-26" to "2026-03-08"
    And exactly 14 elements match "button.dp-day.out"
    And the element "button.dp-day[data-date=2026-01-26]" has classes "out"
    And the element "button.dp-day[data-date=2026-03-08]" has classes "out"
    And no element matches "button.dp-day.today"
    And no element matches "button.dp-day.sel"
    And the element "span.dp-title" has text "Feb 2026"

  Scenario: March 2026 (starts Sunday) still renders six full weeks
    When I render the date picker menu for month "2026-03" with today "2026-06-12" via "/api/datepicker"
    Then exactly 42 elements match "div.dp-grid > button.dp-day"
    And the day cells run from "2026-02-23" to "2026-04-05"
    And exactly 11 elements match "button.dp-day.out"

  Scenario: today is outlined and announced, independent of the selection
    When I render the date picker menu for month "2026-06" selecting "2026-06-05" with today "2026-06-12" via "/api/datepicker"
    Then exactly 1 element matches "button.dp-day.today"
    And the element "button.dp-day.today" has attribute "data-date" equal to "2026-06-12"
    And the element "button.dp-day.today" has attribute "aria-current" equal to "date"
    And exactly 1 element matches "button.dp-day.sel"
    And the element "button.dp-day[data-date=2026-06-05]" has classes "sel"
    And no element matches "button.dp-day.sel.today"

  Scenario: a selection outside the viewed month marks no cell
    When I render the date picker menu for month "2026-06" selecting "2026-02-10" with today "2026-06-12" via "/api/datepicker"
    Then no element matches "button.dp-day.sel"
    And exactly 1 element matches "button.dp-day.today"

  Scenario: nav buttons carry hx-get round-trips for the target months
    When I render the date picker menu for month "2026-06" selecting "2026-06-12" with today "2026-06-12" via "/api/datepicker"
    Then exactly 4 elements match "div.dp-hd > button.dp-nav[type=button][hx-target=closest .dp-menu][hx-swap=innerHTML]"
    And the element "button.dp-nav[aria-label=previous year]" has attribute "hx-get" equal to "/api/datepicker?month=2025-06&selected=2026-06-12"
    And the element "button.dp-nav[aria-label=previous month]" has attribute "hx-get" equal to "/api/datepicker?month=2026-05&selected=2026-06-12"
    And the element "button.dp-nav[aria-label=next month]" has attribute "hx-get" equal to "/api/datepicker?month=2026-07&selected=2026-06-12"
    And the element "button.dp-nav[aria-label=next year]" has attribute "hx-get" equal to "/api/datepicker?month=2027-06&selected=2026-06-12"
    And the element "button.dp-nav[aria-label=previous year]" has text "«"
    And the element "button.dp-nav[aria-label=previous month]" has text "‹"
    And the element "button.dp-nav[aria-label=next month]" has text "›"
    And the element "button.dp-nav[aria-label=next year]" has text "»"

  Scenario: viewing January rolls the year over in nav URLs
    When I render the date picker menu for month "2026-01" with today "2026-06-12" via "/api/datepicker"
    Then the element "button.dp-nav[aria-label=previous month]" has attribute "hx-get" equal to "/api/datepicker?month=2025-12"
    And the element "button.dp-nav[aria-label=next month]" has attribute "hx-get" equal to "/api/datepicker?month=2026-02"
    And the element "button.dp-nav[aria-label=next year]" has attribute "hx-get" equal to "/api/datepicker?month=2027-01"

  Scenario: preset row computes its dates on the server from today
    When I render the date picker menu for month "2026-06" selecting "2026-06-12" with today "2026-06-12" via "/api/datepicker"
    Then exactly 4 elements match "div.dp-presets > button.dp-preset[type=button]"
    And the element "button.dp-preset[data-date=2026-06-12]" has text "today"
    And the element "button.dp-preset[data-date=2026-06-11]" has text "-1d"
    And the element "button.dp-preset[data-date=2026-06-05]" has text "-7d"
    And the element "button.dp-preset[data-date=2026-05-13]" has text "-30d"

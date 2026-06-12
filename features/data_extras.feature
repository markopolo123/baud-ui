Feature: Pagination & DiffViewer — data extras
  The pager footer bar reports the visible row range with thousands
  separators and steps pages via |‹ ‹ prev n/N next › ›| buttons — htmx
  round-trips (hx-get ?page=N into one target) or plain links (HrefFor) —
  plus an optional accent "load more ↓" htmx append (hx-swap beforeend).
  The diff viewer is a server-rendered unified diff: dual line-number
  gutters, +/− rows tinted ok/err, hunk headers tinted info — pure CSS,
  no client behaviour. ParseUnified turns raw unified-diff text into
  DiffLines with both gutters numbered from the hunk headers.

  Scenario: first page — range text, position and disabled lower bound
    When I render an htmx pager at page 1 of 12403 rows, 50 per page, getting "/demo/pagination" into "pager-demo"
    Then exactly 1 element matches "nav.pager"
    And the element "span.pager-info" has text "rows 1–50 of 12,403"
    And the element "span.pager-pos" has text "1/249"
    And exactly 1 element matches "button.pager-btn[aria-label='first page'][disabled]"
    And exactly 1 element matches "button.pager-btn[aria-label='previous page'][disabled]"
    And no element matches "button.pager-btn[aria-label='next page'][disabled]"
    And no element matches "button.pager-btn[aria-label='last page'][disabled]"
    And no element matches "button.pager-btn[aria-label='first page'][hx-get]"
    And no element matches "button.pager-btn[aria-label='previous page'][hx-get]"
    And the element "button.pager-btn[aria-label='next page']" has attribute "hx-get" equal to "/demo/pagination?page=2"
    And the element "button.pager-btn[aria-label='next page']" has attribute "hx-target" equal to "#pager-demo"
    And the element "button.pager-btn[aria-label='last page']" has attribute "hx-get" equal to "/demo/pagination?page=249"
    And no element matches "button.pager-btn.more"

  Scenario: middle page — every button live with the right page numbers
    When I render an htmx pager at page 5 of 12403 rows, 50 per page, getting "/demo/pagination" into "pager-demo"
    Then the element "span.pager-info" has text "rows 201–250 of 12,403"
    And the element "span.pager-pos" has text "5/249"
    And no element matches "button.pager-btn[disabled]"
    And the element "button.pager-btn[aria-label='first page']" has attribute "hx-get" equal to "/demo/pagination?page=1"
    And the element "button.pager-btn[aria-label='previous page']" has attribute "hx-get" equal to "/demo/pagination?page=4"
    And the element "button.pager-btn[aria-label='next page']" has attribute "hx-get" equal to "/demo/pagination?page=6"
    And the element "button.pager-btn[aria-label='last page']" has attribute "hx-get" equal to "/demo/pagination?page=249"

  Scenario: last page — partial range and disabled upper bound
    When I render an htmx pager at page 249 of 12403 rows, 50 per page, getting "/demo/pagination" into "pager-demo"
    Then the element "span.pager-info" has text "rows 12,401–12,403 of 12,403"
    And the element "span.pager-pos" has text "249/249"
    And exactly 1 element matches "button.pager-btn[aria-label='next page'][disabled]"
    And exactly 1 element matches "button.pager-btn[aria-label='last page'][disabled]"
    And no element matches "button.pager-btn[aria-label='next page'][hx-get]"
    And the element "button.pager-btn[aria-label='first page']" has attribute "hx-get" equal to "/demo/pagination?page=1"
    And the element "button.pager-btn[aria-label='previous page']" has attribute "hx-get" equal to "/demo/pagination?page=248"

  Scenario: a single page disables both bounds
    When I render an htmx pager at page 1 of 7 rows, 50 per page, getting "/demo/pagination" into "pager-demo"
    Then the element "span.pager-info" has text "rows 1–7 of 7"
    And the element "span.pager-pos" has text "1/1"
    And exactly 4 elements match "button.pager-btn[disabled]"

  Scenario: an empty result renders a zero range, everything disabled
    When I render an htmx pager at page 1 of 0 rows, 50 per page, getting "/demo/pagination" into "pager-demo"
    Then the element "span.pager-info" has text "rows 0–0 of 0"
    And the element "span.pager-pos" has text "1/1"
    And exactly 4 elements match "button.pager-btn[disabled]"

  Scenario: link mode — enabled steps are plain anchors, no htmx
    When I render an href pager at page 2 of 100 rows, 10 per page, linking "/rows"
    Then exactly 4 elements match "a.pager-btn"
    And the element "a.pager-btn[aria-label='first page']" has attribute "href" equal to "/rows?page=1"
    And the element "a.pager-btn[aria-label='previous page']" has attribute "href" equal to "/rows?page=1"
    And the element "a.pager-btn[aria-label='next page']" has attribute "href" equal to "/rows?page=3"
    And the element "a.pager-btn[aria-label='last page']" has attribute "href" equal to "/rows?page=10"
    And no element matches "a.pager-btn[hx-get]"
    And no element matches "button.pager-btn[disabled]"

  Scenario: link mode bounds — disabled ends are real disabled buttons, not dead links
    When I render an href pager at page 1 of 100 rows, 10 per page, linking "/rows"
    Then exactly 2 elements match "button.pager-btn[disabled]"
    And exactly 2 elements match "a.pager-btn"
    And exactly 1 element matches "button.pager-btn[aria-label='first page'][disabled]"
    And exactly 1 element matches "button.pager-btn[aria-label='previous page'][disabled]"
    And no element matches "a.pager-btn[aria-label='first page']"

  Scenario: load more — accent button appends via htmx
    When I render an htmx pager at page 1 of 12403 rows, 50 per page, getting "/demo/pagination" into "pager-demo" with load more "/demo/pagination?page=2&append=1" into "pager-list"
    Then exactly 1 element matches "button.pager-btn.more"
    And the element "button.pager-btn.more" has text "load more ↓"
    And the element "button.pager-btn.more" has attribute "hx-get" equal to "/demo/pagination?page=2&append=1"
    And the element "button.pager-btn.more" has attribute "hx-target" equal to "#pager-list"
    And the element "button.pager-btn.more" has attribute "hx-swap" equal to "beforeend"

  Scenario: each diff line kind renders its marker class, gutters and sign
    When I render a diff "cmd/main.go" from rows:
      """
      hunk|||@@ -10,4 +10,5 @@
      ctx|10|10|func main() {
      del|11||  run(1)
      add||11|  run(2)
      add||12|  audit()
      ctx|12|13|}
      """
    Then exactly 1 element matches "div.diff"
    And exactly 6 elements match "div.diff-line"
    And exactly 1 element matches "div.diff-line.hunk"
    And exactly 2 elements match "div.diff-line.add"
    And exactly 1 element matches "div.diff-line.del"
    And diff row 1 is a hunk with text "@@ -10,4 +10,5 @@"
    And diff row 2 is "ctx" with old "10" new "10" sign " " and text "func main() {"
    And diff row 3 is "del" with old "11" new "" sign "−" and text "  run(1)"
    And diff row 4 is "add" with old "" new "11" sign "+" and text "  run(2)"
    And diff row 5 is "add" with old "" new "12" sign "+" and text "  audit()"
    And diff row 6 is "ctx" with old "12" new "13" sign " " and text "}"

  Scenario: the diff header reports the file and add/del counts
    When I render a diff "cmd/main.go" from rows:
      """
      hunk|||@@ -10,4 +10,5 @@
      ctx|10|10|func main() {
      del|11||  run(1)
      add||11|  run(2)
      add||12|  audit()
      ctx|12|13|}
      """
    Then exactly 1 element matches "div.diff > div.diff-hd"
    And the element "span.diff-file" has text "cmd/main.go"
    And the element "span.diff-adds" has text "+2"
    And the element "span.diff-dels" has text "−1"

  Scenario: a parsed unified diff numbers both gutters from the hunk header
    When I render the parsed unified diff "main.go":
      """
      --- a/main.go
      +++ b/main.go
      @@ -1,3 +1,4 @@
       package main
      -var x = 1
      +var x = 2
      +var y = 3
       func main() {}
      """
    Then exactly 1 element matches "div.diff-line.hunk"
    And diff row 1 is a hunk with text "@@ -1,3 +1,4 @@"
    And diff row 2 is "ctx" with old "1" new "1" sign " " and text "package main"
    And diff row 3 is "del" with old "2" new "" sign "−" and text "var x = 1"
    And diff row 4 is "add" with old "" new "2" sign "+" and text "var x = 2"
    And diff row 5 is "add" with old "" new "3" sign "+" and text "var y = 3"
    And diff row 6 is "ctx" with old "3" new "4" sign " " and text "func main() {}"

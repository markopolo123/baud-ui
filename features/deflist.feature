Feature: DefList and Breadcrumb structure components
  DefList is a semantic <dl> on a max-content/1fr grid: UPPERCASE faint keys,
  tabular-nums values, optional hairline row rules (Lines), and a children
  slot (DefItem) for rich values. Breadcrumb is a labelled nav + list:
  non-current crumbs are faint links, the current crumb is bold with
  aria-current="page", and the › separators are presentational only.

  Scenario: deflist renders one dt/dd pair per row in a semantic dl
    When I render a deflist with pairs "host=db-04.prod,region=eu-west-1,uptime=41d 6h"
    Then exactly 1 element matches "dl.dl"
    And exactly 3 elements match "dl.dl > dt.dl-k"
    And exactly 3 elements match "dl.dl > dd.dl-v"
    And no element matches "dl.dl-lines"

  Scenario: deflist key and value text land in the right cells
    When I render a deflist with pairs "kernel=6.8.0-41"
    Then the element "dt.dl-k" has text "kernel"
    And the element "dd.dl-v" has text "6.8.0-41"

  Scenario: lines flag opts into hairline row rules
    When I render a lined deflist with pairs "cpu=72%,mem=18.2 GiB"
    Then exactly 1 element matches "dl.dl.dl-lines"
    And exactly 2 elements match "dl.dl-lines > dt.dl-k"
    And exactly 2 elements match "dl.dl-lines > dd.dl-v"

  Scenario: rich values render through the DefItem children slot
    When I render a deflist with a rich value keyed "status"
    Then exactly 1 element matches "dl.dl > dt.dl-k"
    And the element "dt.dl-k" has text "status"
    And exactly 1 element matches "dl.dl > dd.dl-v > span"
    And the element "dd.dl-v > span" has text "ok"

  Scenario: plain rows and rich rows mix in one deflist
    When I render a deflist with pairs "host=db-04.prod" and a rich value keyed "status"
    Then exactly 2 elements match "dl.dl > dt.dl-k"
    And exactly 2 elements match "dl.dl > dd.dl-v"
    And exactly 1 element matches "dl.dl > dd.dl-v > span"

  Scenario: breadcrumb renders a labelled nav wrapping a crumb list
    When I render a breadcrumb with trail "prod,core,ingest-gw"
    Then exactly 1 element matches "nav[aria-label=breadcrumb] > ol.crumbs"
    And exactly 3 elements match "ol.crumbs > li"

  Scenario: non-current crumbs are links, the current crumb is not
    When I render a breadcrumb with trail "prod,core,ingest-gw"
    Then exactly 2 elements match "li > a.crumb"
    And exactly 1 element matches "span.crumb.cur[aria-current=page]"
    And no element matches "a.crumb.cur"
    And the element "span.crumb.cur" has text "ingest-gw"

  Scenario: separators sit between crumbs and are presentational only
    When I render a breadcrumb with trail "prod,core,ingest-gw"
    Then exactly 2 elements match "li > span.crumb-sep[aria-hidden=true]"
    And no element matches "span.crumb-sep[aria-current]"

  Scenario: non-current crumbs carry their hrefs
    When I render a breadcrumb with trail "prod,core"
    Then exactly 1 element matches "a.crumb[href=/prod]"
    And the element "a.crumb" has text "prod"
    And the element "span.crumb.cur" has text "core"

  Scenario: a single-crumb breadcrumb is just the current page
    When I render a breadcrumb with trail "prod"
    Then exactly 1 element matches "span.crumb.cur[aria-current=page]"
    And no element matches "a.crumb"
    And no element matches "span.crumb-sep"

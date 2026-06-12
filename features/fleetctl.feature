Feature: fleetctl console — the acceptance-test composition
  The fleetctl demo console (design/README.md "Deliverables" item 3) replaces
  the placeholder at / and composes the shipped library components into one
  ops screen: app shell with topbar Tabs + global action Btns, Tree navigator
  in the nav slot, a metrics strip of Panels (DefList · Progress · Badge ·
  Dot), the sortable hosts DataTable with tone-coded threshold cells and a
  Pagination footer, a tone-coded log tail Panel, a load-bearing StatusBar,
  and the wired overlays: host drawer (DefList + DiffViewer + ConfirmInput
  kill guard → OOB toast), deploy Modal, and the ⌘K CommandPalette filtered
  by a console-specific endpoint. These scenarios assert the rendered
  composition and its htmx/hyperscript wiring attributes; the browser flows
  live in e2e/fleetctl_test.go.

  Scenario: the console fills the app shell with topbar, nav, main and statusbar
    When I render the fleetctl console
    Then exactly 1 element matches "div[data-shell]"
    And exactly 1 element matches "header[data-topbar]"
    And exactly 1 element matches "aside[data-nav]"
    And exactly 1 element matches "footer[data-statusbar]"
    And the element "span.brand-name" has text "baud/fleetctl"

  Scenario: the statusbar carries mode, env, counts, a spring and a clock cell
    When I render the fleetctl console
    Then exactly 1 element matches "footer[data-statusbar]>div.statusbar"
    And the element "span.sb-cell.sb-mode" has text "FLEET"
    And exactly 1 element matches "span.sb-cell.sb-spring"
    And the fleetctl console has a "span.sb-cell" with text "prod · us-east-1"
    And the fleetctl console has a "span.sb-cell" with text "14:32:41 utc"

  Scenario: topbar tabs hx-get their view pane into the shared tabpanel
    When I render the fleetctl console
    Then exactly 3 elements match "div.tabs[role=tablist]>button[role=tab]"
    And the element "button[id=fleet-tabs-tab-0]" has attribute "aria-selected" equal to "true"
    And the element "button[id=fleet-tabs-tab-0]" has attribute "hx-get" equal to "/fleet/tab?view=fleet"
    And the element "button[id=fleet-tabs-tab-1]" has attribute "hx-get" equal to "/fleet/tab?view=incidents"
    And the element "button[id=fleet-tabs-tab-2]" has attribute "hx-get" equal to "/fleet/tab?view=deploys"
    And the element "button[id=fleet-tabs-tab-1]" has attribute "hx-target" equal to "#fleet-view"
    And exactly 1 element matches "div.tab-panel[id=fleet-view][role=tabpanel]"

  Scenario: the deploy action and the ⌘K hint are global topbar buttons
    When I render the fleetctl console
    Then the element "button[id=fleet-deploy-btn]" has attribute "hx-get" equal to "/fleet/deploy"
    And the element "button[id=fleet-deploy-btn]" has attribute "hx-target" equal to "body"
    And the element "button[id=fleet-deploy-btn]" has attribute "hx-swap" equal to "beforeend"
    And the element "button[id=fleet-deploy-btn]" has classes "btn btn-primary"
    And the fleetctl element "button[id=fleet-deploy-btn]" attribute "_" contains "fleet-deploy"
    And the element "button[id=fleet-cmd-btn]>kbd.kbd" has text "⌘K"
    And the fleetctl element "button[id=fleet-cmd-btn]" attribute "_" contains "send baud:palette"

  Scenario: the nav hosts the region → cluster → host tree with a lazy branch
    When I render the fleetctl console
    Then exactly 1 element matches "aside[data-nav]>div.fleet-nav>ul.tree[id=fleet-tree]"
    And the element "li.tree-row.sel>span.tree-label" has text "ingest-gw"
    And the element "li.tree-row.sel" has attribute "aria-current" equal to "true"
    And exactly 1 element matches "details.tree-branch[hx-get=/fleet/tree?node=euw1/edge]"

  Scenario: the metrics strip composes panels of deflists, progress and badges
    When I render the fleetctl console
    Then exactly 3 elements match "div[id=fleet-metrics]>section[data-panel]"
    And the element "dd.dl-v>span.fl-err" has text "312 ms"
    And the element "dd.dl-v>span.fl-warn" has text "0.61%"
    And exactly 3 elements match "section[id=fleet-m-capacity]>div.panel-bd>div.fleet-pad>span.prog"
    And exactly 1 element matches "span.prog-bar.tone-warn"
    And exactly 1 element matches "span.badge.bd-tint.tone-err"
    And exactly 1 element matches "span.badge.bd-outline.tone-err"
    And exactly 1 element matches "span.dot.tone-ok.pulse"

  Scenario: the hosts table sorts via htmx, err desc first, ingest-gw selected
    When I render the fleetctl console
    Then exactly 1 element matches "table.dt[id=fleet-hosts]"
    And the element "th[data-key=err]" has attribute "aria-sort" equal to "descending"
    And the element "th[data-key=err]>span.sort-arrow" has text "▼"
    And the element "th[data-key=err]" has attribute "hx-get" equal to "/fleet/hosts?sort=err&dir=asc"
    And the element "th[data-key=cpu]" has attribute "hx-get" equal to "/fleet/hosts?sort=cpu&dir=asc"
    And the element "tr.sel>td.row-mark" has text "▌"
    And exactly 4 elements match "tr.sel>td.tone-err"
    And the element "tr.sel>td.tone-warn" has text "88.9"
    And the element "button[id=fleet-inspect-btn]" has attribute "hx-get" equal to "/fleet/host?id=ingest-gw"
    And the fleetctl element "button[id=fleet-inspect-btn]" attribute "_" contains "fleet-inspect"

  Scenario: the pagination footer reports the single fixture page honestly
    When I render the fleetctl console
    Then exactly 1 element matches "nav.pager"
    And the element "span.pager-info" has text "rows 1–12 of 12"
    And the element "span.pager-pos" has text "1/1"
    And exactly 4 elements match "nav.pager>button.pager-btn[disabled]"

  Scenario: the log tail renders tone-coded fixture lines
    When I render the fleetctl console
    Then exactly 10 elements match "div.fleet-log>div.fleet-log-line"
    And exactly 3 elements match "div.fleet-log-line>span.fl-lvl.fl-err"
    And exactly 2 elements match "div.fleet-log-line>span.fl-lvl.fl-warn"
    And exactly 5 elements match "div.fleet-log-line>span.fl-lvl.fl-info"

  Scenario: the host drawer composes deflist, diff and the guarded kill action
    When I render the fleetctl host drawer for "ingest-gw"
    Then exactly 1 element matches "div.overlay.drawer-wrap>div.drawer[role=dialog]"
    And the element "div.overlay" has attribute "id" equal to "fleet-drawer"
    And the element "span.modal-title" has text "ingest-gw · v2.14.0"
    And exactly 1 element matches "dl.dl"
    And exactly 1 element matches "div.diff"
    And the element "input[data-confirm=ingest-gw]" has attribute "id" equal to "fleet-confirm-input"
    And the element "div[id=fleet-kill]" has attribute "hx-get" equal to "/fleet/kill?host=ingest-gw"
    And the element "div[id=fleet-kill]" has attribute "hx-trigger" equal to "click from:find .btn-danger"
    And the element "div[id=fleet-kill]" has attribute "hx-swap" equal to "none"
    And exactly 1 element matches "div.confirm-row>button.btn.btn-danger[disabled]"

  Scenario: the deploy modal guards the rollout behind an htmx confirm
    When I render the fleetctl deploy modal
    Then exactly 1 element matches "div.overlay>div.modal[role=dialog]"
    And the element "div.overlay" has attribute "id" equal to "fleet-modal"
    And the element "button[id=fleet-deploy-run]" has attribute "hx-get" equal to "/fleet/deploy/run"
    And the element "button[id=fleet-deploy-run]" has attribute "hx-swap" equal to "none"
    And the fleetctl element "button[id=fleet-deploy-run]" attribute "_" contains "send baud:overlayClose"
    And exactly 1 element matches "footer.modal-ft>button[data-overlay-close]"

  Scenario: the command palette is wired to the console filter endpoint
    When I render the fleetctl console
    Then the element "div.palette-overlay" has attribute "id" equal to "fleet-palette"
    And the element "input[id=fleet-palette-input]" has attribute "hx-get" equal to "/fleet/palette"
    And exactly 1 element matches "button.palette-item[data-cmd=fleet-deploy]"
    And exactly 1 element matches "button.palette-item[data-cmd=fleet-inspect]"
    And exactly 1 element matches "a.palette-item[href=/sheet]"

  Scenario: the incidents tab pane renders server-side
    When I render the fleetctl "incidents" tab pane
    Then exactly 1 element matches "section[id=fleet-incidents][data-panel]"
    And exactly 1 element matches "span.badge.bd-solid.tone-err"

  Scenario: the deploys tab pane renders the empty panel state
    When I render the fleetctl "deploys" tab pane
    Then exactly 1 element matches "section[id=fleet-deploys][data-panel]"
    And exactly 1 element matches "div.pstate"

  Scenario: the static render keeps relative hrefs for GitHub Pages
    When I render the fleetctl console for the static site
    Then the element "a.brand-sub" has attribute "href" equal to "index.html"
    And exactly 1 element matches "a.palette-item[href=index.html]"

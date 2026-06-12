Feature: Tree — box-drawing branches, disclosure, lazy htmx children
  A Tree renders nested nodes as rows whose depth is drawn with box-drawing
  branch glyphs (├─ for a middle sibling, └─ for the last, │ continuing each
  still-open ancestor) in --fg-faint. Branch nodes are native
  <details>/<summary> disclosures (free keyboard + a11y) showing ▸ collapsed
  and ▾ expanded, with aria-expanded kept in sync; child lists are
  role=group. A node with a LazyURL fetches its child-list fragment via
  hx-get on first toggle only. The selected row carries --sel + the accent
  inset bar. Tree specs below: two spaces per depth level; suffixes mark a
  node `>` expanded branch, `+` collapsed branch, `~URL` lazy branch,
  `*` selected, `[meta]` right-aligned meta.

  Scenario: a flat tree is a list of plain leaf rows
    When I render a tree from:
      """
      docs
      cmd
      assets
      """
    Then exactly 1 element matches "ul.tree"
    And exactly 3 elements match "ul.tree > li.tree-row"
    And no element matches "details.tree-branch"
    And no element matches "span.tree-disclosure"

  Scenario: root nodes carry no branch glyphs
    When I render a tree from:
      """
      prod>
        core
      docs
      """
    Then the tree row labeled "prod" has glyph ""
    And the tree row labeled "docs" has glyph ""

  Scenario: middle siblings draw the tee, the last sibling the corner
    When I render a tree from:
      """
      prod>
        core
        edge
        batch
      """
    Then the tree row labeled "core" has glyph "├─"
    And the tree row labeled "edge" has glyph "├─"
    And the tree row labeled "batch" has glyph "└─"

  Scenario: an open ancestor continues as a pipe, a closed one as spaces
    When I render a tree from:
      """
      prod>
        core>
          ingest-gw
          api-core
        edge>
          cache
      """
    Then the tree row labeled "ingest-gw" has glyph "│  ├─"
    And the tree row labeled "api-core" has glyph "│  └─"
    And the tree row labeled "cache" has glyph "   └─"

  Scenario: branch rows are disclosure summaries with the ▸/▾ glyph slot
    When I render a tree from:
      """
      prod>
        core
      """
    Then exactly 1 element matches "li > details.tree-branch > summary.tree-row"
    And exactly 1 element matches "summary.tree-row > span.tree-glyph > span.tree-disclosure[aria-hidden=true]"
    And exactly 1 element matches "details.tree-branch > ul[role=group]"

  Scenario: an expanded branch is open with aria-expanded true
    When I render a tree from:
      """
      prod>
        core
      """
    Then exactly 1 element matches "details.tree-branch[open]"
    And the element "summary.tree-row" has attribute "aria-expanded" equal to "true"
    And exactly 1 element matches "details[open] > ul[role=group] > li.tree-row"

  Scenario: a collapsed branch keeps its loaded children in the DOM
    When I render a tree from:
      """
      staging+
        smoke
        soak
      """
    Then no element matches "details.tree-branch[open]"
    And the element "summary.tree-row" has attribute "aria-expanded" equal to "false"
    And exactly 2 elements match "details.tree-branch > ul[role=group] > li.tree-row"

  Scenario: every branch syncs aria-expanded on its native toggle
    When I render a tree from:
      """
      prod>
        core
      """
    Then the element "details.tree-branch" has attribute "_" equal to "on toggle set @aria-expanded of the first <summary/> in me to me.open as String"

  Scenario: the meta slot renders right-aligned per node and only when given
    When I render a tree from:
      """
      prod>
        ingest-gw [12 pods]
        api-core
      """
    Then exactly 1 element matches "li.tree-row > span.tree-meta"
    And the element "span.tree-meta" has text "12 pods"
    And the tree row labeled "api-core" has no meta

  Scenario: the selected node is marked sel and aria-current
    When I render a tree from:
      """
      prod>
        core>
          ingest-gw*
          api-core
      """
    Then exactly 1 element matches ".tree-row.sel"
    And the tree row labeled "ingest-gw" is the selected row
    And exactly 1 element matches "li.tree-row.sel[aria-current=true]"
    And no element matches "summary.tree-row.sel"

  Scenario: a lazy branch hx-gets its child fragment on first toggle only
    When I render a tree from:
      """
      prod>
        edge~/demo/tree?node=prod/edge
      """
    Then exactly 1 element matches "details.tree-branch[hx-get=/demo/tree?node=prod/edge]"
    And the element "details.tree-branch[hx-get]" has attribute "hx-trigger" equal to "toggle once"
    And the element "details.tree-branch[hx-get]" has attribute "hx-target" equal to "find ul"
    And the element "details.tree-branch[hx-get]" has attribute "hx-swap" equal to "innerHTML"
    And exactly 0 elements match "details.tree-branch[hx-get] > ul[role=group] > li"

  Scenario: lazy branches always start collapsed
    When I render a tree from:
      """
      edge>~/demo/tree?node=edge
      """
    Then no element matches "details.tree-branch[open]"
    And the element "summary.tree-row" has attribute "aria-expanded" equal to "false"

  Scenario: loaded branches carry no htmx wiring
    When I render a tree from:
      """
      staging+
        smoke
      """
    Then no element matches "details[hx-get]"
    And no element matches "details[hx-trigger]"

  Scenario: the exported child-list fragment renders rows under a prefix
    When I render a tree children fragment with prefix "   " from:
      """
      edge-cache-1 [warm]
      edge-cache-2
      """
    Then exactly 2 elements match "li.tree-row"
    And no element matches "ul.tree"
    And the tree row labeled "edge-cache-1" has glyph "   ├─"
    And the tree row labeled "edge-cache-2" has glyph "   └─"
    And the element "span.tree-meta" has text "warm"

  Scenario: leaf glyphs keep the two-space alignment pad
    When I render a tree from:
      """
      prod>
        core
        batch
      """
    Then the tree row labeled "core" has exact glyph text "├─  "
    And the tree row labeled "batch" has exact glyph text "└─  "

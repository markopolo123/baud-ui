Feature: TagInput — key=value chips with hidden form inputs and suggestions
  Chips render the key in a faint span and carry a real <input type=hidden>
  so the tags submit with forms; removing a chip removes its hidden input.
  The visible text input is nameless so it never pollutes the submission.
  Enter adds, Backspace pops and the suggestion menu filters — all local UI
  in the TagInput hyperscript behavior, dismissed via MenuDismiss.

  Scenario: root installs the behaviors and carries the form name
    When I render a tag input named "labels" with values "env=prod"
    Then exactly 1 element matches "span.tags[data-name=labels]"
    And the element "span.tags" has attribute "_" equal to "install TagInput install MenuDismiss"

  Scenario: key=value chip splits the key into a faint span
    When I render a tag input named "labels" with values "env=prod"
    Then exactly 1 element matches "span.tags-wrap > span.tag-chip"
    And the element "span.tag-chip > span.tag-k" has text "env="
    And the element "span.tag-chip > span.tag-v" has text "prod"

  Scenario: bare value chip renders no key span
    When I render a tag input named "tags" with values "canary"
    Then exactly 1 element matches "span.tag-chip"
    And no element matches "span.tag-k"
    And the element "span.tag-chip > span.tag-v" has text "canary"

  Scenario: every chip carries a hidden input with the shared form name
    When I render a tag input named "labels" with values "env=prod,region=eu-west,canary"
    Then exactly 3 elements match "span.tag-chip"
    And exactly 3 elements match "span.tag-chip > input[type=hidden][name=labels]"
    And exactly 1 element matches "span.tag-chip > input[type=hidden][value=env=prod]"
    And exactly 1 element matches "span.tag-chip > input[type=hidden][value=region=eu-west]"
    And exactly 1 element matches "span.tag-chip > input[type=hidden][value=canary]"

  Scenario: every chip has an accessible remove button
    When I render a tag input named "labels" with values "env=prod"
    Then exactly 1 element matches "span.tag-chip > button.x-btn[type=button]"
    And the element "button.x-btn" has attribute "aria-label" equal to "remove env=prod"
    And the element "button.x-btn" has text "✕"

  Scenario: the visible text input is nameless and placeholdered
    When I render a tag input named "labels" with values "env=prod" and placeholder "add label…"
    Then exactly 1 element matches "span.tags-wrap > input.tags-input[type=text][placeholder=add label…]"
    And no element matches "input.tags-input[name]"

  Scenario: id lands on the text input for field pairing
    When I render a tag input with id "ti-1" named "labels"
    Then exactly 1 element matches "input.tags-input[id=ti-1]"

  Scenario: suggestions render as menu items with data-tag payloads
    When I render a tag input named "labels" with suggestions "env=staging,team=core"
    Then exactly 1 element matches "span.tags > div.tags-menu"
    And exactly 2 elements match "div.tags-menu > button.tags-menu-item[type=button]"
    And exactly 1 element matches "button.tags-menu-item[data-tag=env=staging]"
    And exactly 1 element matches "button.tags-menu-item[data-tag=team=core]"
    And the element "button.tags-menu-item[data-tag=team=core] > span.tags-menu-tag" has text "team=core"
    And the element "button.tags-menu-item[data-tag=team=core] > span.tags-menu-meta" has text "add"

  Scenario: no suggestions renders no menu
    When I render a tag input named "labels" with values "env=prod"
    Then no element matches "div.tags-menu"

  Scenario: empty state renders just the nameless input
    When I render a tag input named "labels" with placeholder "add label…"
    Then no element matches "span.tag-chip"
    And no element matches "input[type=hidden]"
    And exactly 1 element matches "span.tags-wrap > input.tags-input"

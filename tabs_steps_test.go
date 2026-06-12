package baudui_test

import (
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// Step definitions for features/tabs.feature. Registered with one line in
// InitializeScenario (steps_test.go) — the only file this feature shares
// with other waves. Assertions reuse the shared exactlyNMatch /
// elementHasClasses / elementHasAttr / elementHasText steps.

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderUnderlineTabs(labels string, active int) error {
	return s.render(baud.Tabs(baud.TabsProps{
		Tabs:        tabsFromSpec(labels),
		ActiveIndex: active,
	}))
}

func (s *scenarioState) renderBoxedTabs(labels string, active int) error {
	return s.render(baud.Tabs(baud.TabsProps{
		Variant:     baud.TabsBoxed,
		Tabs:        tabsFromSpec(labels),
		ActiveIndex: active,
	}))
}

func (s *scenarioState) renderUnderlineTabsWithCounts(spec string) error {
	return s.render(baud.Tabs(baud.TabsProps{Tabs: tabsFromSpec(spec)}))
}

func (s *scenarioState) renderBoxedTabsWithCounts(spec string) error {
	return s.render(baud.Tabs(baud.TabsProps{Variant: baud.TabsBoxed, Tabs: tabsFromSpec(spec)}))
}

func (s *scenarioState) renderHTMXTabs(spec, target string) error {
	var tabs []baud.Tab
	for _, part := range strings.Split(spec, ",") {
		label, href, _ := strings.Cut(part, "=")
		tabs = append(tabs, baud.Tab{Label: label, Href: href})
	}
	return s.render(baud.Tabs(baud.TabsProps{
		ID:     "demo-tabs",
		Tabs:   tabs,
		Target: target,
	}))
}

func (s *scenarioState) renderLocalTabs(spec string, active int) error {
	var tabs []baud.Tab
	for _, part := range strings.Split(spec, ",") {
		label, panel, _ := strings.Cut(part, "=")
		tabs = append(tabs, baud.Tab{Label: label, Panel: panel})
	}
	return s.render(baud.Tabs(baud.TabsProps{
		ID:          "demo-tabs",
		Tabs:        tabs,
		ActiveIndex: active,
	}))
}

func (s *scenarioState) renderActiveTabPanel(id string) error {
	return s.render(withChild{baud.TabPanel(id, true), textPane("pane body")})
}

func (s *scenarioState) renderInactiveTabPanel(id string) error {
	return s.render(withChild{baud.TabPanel(id, false), textPane("pane body")})
}

// tabsFromSpec parses "pods,events" / "pods=39,events=7,logs" into Tabs —
// a bare label, or label=count for a count-badged tab.
func tabsFromSpec(spec string) []baud.Tab {
	var tabs []baud.Tab
	for _, part := range strings.Split(spec, ",") {
		label, count, ok := strings.Cut(part, "=")
		t := baud.Tab{Label: label}
		if ok {
			n, err := strconv.Atoi(count)
			if err == nil {
				t.Count = &n
			}
		}
		tabs = append(tabs, t)
	}
	return tabs
}

// registerTabsSteps wires the Tabs scenario steps onto the shared state.
func registerTabsSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render underline tabs "([^"]*)" with active index (\d+)$`, s.renderUnderlineTabs)
	sc.When(`^I render boxed tabs "([^"]*)" with active index (\d+)$`, s.renderBoxedTabs)
	sc.When(`^I render underline tabs with counts "([^"]*)"$`, s.renderUnderlineTabsWithCounts)
	sc.When(`^I render boxed tabs with counts "([^"]*)"$`, s.renderBoxedTabsWithCounts)
	sc.When(`^I render htmx tabs "([^"]*)" targeting "([^"]*)"$`, s.renderHTMXTabs)
	sc.When(`^I render local tabs "([^"]*)" with active index (\d+)$`, s.renderLocalTabs)
	sc.When(`^I render an active tab panel "([^"]*)"$`, s.renderActiveTabPanel)
	sc.When(`^I render an inactive tab panel "([^"]*)"$`, s.renderInactiveTabPanel)
}

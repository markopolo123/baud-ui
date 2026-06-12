package baudui_test

import (
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// deflistPairs parses a "key=value,key=value" spec into DefList rows.
func deflistPairs(spec string) []baud.KV {
	var rows []baud.KV
	for _, part := range strings.Split(spec, ",") {
		k, v, _ := strings.Cut(part, "=")
		rows = append(rows, baud.KV{Key: k, Value: v})
	}
	return rows
}

// deflistCrumbs turns a "prod,core,ingest-gw" trail into breadcrumb items;
// non-current hrefs derive from the label ("/prod").
func deflistCrumbs(trail string) []baud.Crumb {
	labels := strings.Split(trail, ",")
	items := make([]baud.Crumb, len(labels))
	for i, l := range labels {
		items[i] = baud.Crumb{Label: l, Href: "/" + l}
	}
	return items
}

// ---- render steps ---------------------------------------------------------

func (s *scenarioState) renderDefList(pairs string) error {
	return s.render(baud.DefList(baud.DefListProps{Rows: deflistPairs(pairs)}))
}

func (s *scenarioState) renderLinedDefList(pairs string) error {
	return s.render(baud.DefList(baud.DefListProps{Rows: deflistPairs(pairs), Lines: true}))
}

func (s *scenarioState) renderDefListRichValue(key string) error {
	return s.render(withChild{
		baud.DefList(baud.DefListProps{}),
		withChild{baud.DefItem(key), textPane("ok")},
	})
}

func (s *scenarioState) renderDefListMixed(pairs, key string) error {
	return s.render(withChild{
		baud.DefList(baud.DefListProps{Rows: deflistPairs(pairs)}),
		withChild{baud.DefItem(key), textPane("ok")},
	})
}

func (s *scenarioState) renderBreadcrumb(trail string) error {
	return s.render(baud.Breadcrumb(baud.BreadcrumbProps{Items: deflistCrumbs(trail)}))
}

// registerDefListSteps wires the DefList/Breadcrumb steps into the shared suite.
func registerDefListSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a deflist with pairs "([^"]*)"$`, s.renderDefList)
	sc.When(`^I render a lined deflist with pairs "([^"]*)"$`, s.renderLinedDefList)
	sc.When(`^I render a deflist with a rich value keyed "([^"]*)"$`, s.renderDefListRichValue)
	sc.When(`^I render a deflist with pairs "([^"]*)" and a rich value keyed "([^"]*)"$`, s.renderDefListMixed)
	sc.When(`^I render a breadcrumb with trail "([^"]*)"$`, s.renderBreadcrumb)
}

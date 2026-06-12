package baudui_test

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// registerPaletteSteps wires the CommandPalette scenario steps onto the
// shared scenario state. Registered with one line in InitializeScenario
// (steps_test.go) — the only file this feature shares with other waves.
func registerPaletteSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a palette with id "([^"]*)" searching "([^"]*)" with commands "([^"]*)"$`, s.renderPalette)
	sc.When(`^I render a static palette with id "([^"]*)" highlighting row (\d+) with commands "([^"]*)"$`, s.renderStaticPalette)
	sc.When(`^I render palette results for id "([^"]*)" highlighting row (\d+) from "([^"]*)"$`, s.renderPaletteResults)
	sc.When(`^I render palette results for id "([^"]*)" with no commands$`, s.renderPaletteResultsEmpty)

	sc.Then(`^each palette row has an id derived from "([^"]*)"$`, s.paletteRowIDsDerived)
}

// paletteParseCommands parses a command spec: comma-separated commands,
// each "cat|label|kbd|target". A target starting with "/" is an Href,
// any other non-empty target is an Action (a hyperscript handler), and
// an empty target leaves the row inert.
func paletteParseCommands(spec string) []baud.PaletteCommand {
	var cmds []baud.PaletteCommand
	for _, part := range strings.Split(spec, ",") {
		f := strings.SplitN(part, "|", 4)
		for len(f) < 4 {
			f = append(f, "")
		}
		c := baud.PaletteCommand{Category: f[0], Label: f[1], Kbd: f[2]}
		if strings.HasPrefix(f[3], "/") {
			c.Href = f[3]
		} else {
			c.Action = f[3]
		}
		cmds = append(cmds, c)
	}
	return cmds
}

// ---- render steps ----------------------------------------------------------

func (s *scenarioState) renderPalette(id, url, commands string) error {
	return s.render(baud.Palette(baud.PaletteProps{
		ID: id, SearchURL: url, Commands: paletteParseCommands(commands),
	}))
}

func (s *scenarioState) renderStaticPalette(id string, highlight int, commands string) error {
	return s.render(baud.Palette(baud.PaletteProps{
		ID: id, Static: true, Highlight: highlight, Commands: paletteParseCommands(commands),
	}))
}

func (s *scenarioState) renderPaletteResults(id string, highlight int, commands string) error {
	return s.render(baud.PaletteResults(baud.PaletteResultsProps{
		ID: id, Highlight: highlight, Commands: paletteParseCommands(commands),
	}))
}

func (s *scenarioState) renderPaletteResultsEmpty(id string) error {
	return s.render(baud.PaletteResults(baud.PaletteResultsProps{ID: id}))
}

// ---- assertion steps ---------------------------------------------------------

// paletteRowIDsDerived asserts every row id is prefix-cmd-<index>, in
// document order — the contract aria-activedescendant relies on.
func (s *scenarioState) paletteRowIDsDerived(prefix string) error {
	rows := s.matching(".palette-item")
	if len(rows) == 0 {
		return fmt.Errorf("no .palette-item rows found")
	}
	for i, n := range rows {
		want := fmt.Sprintf("%s-cmd-%d", prefix, i)
		if got := attrVal(n, "id"); got != want {
			return fmt.Errorf("row %d id = %q, want %q", i, got, want)
		}
	}
	return nil
}

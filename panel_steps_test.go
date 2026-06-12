package baudui_test

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// Step definitions for features/panel.feature (Panel / StatusBar / Toolbar).
// Registered with one line in InitializeScenario (steps_test.go) — the only
// file this feature shares with other waves. Assertions reuse the shared
// godog helpers (exactlyNMatch, elementHasText, first/lastElementChildIs…).

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderPanelTitled(title string) error {
	return s.render(withChildren(baud.Panel(baud.PanelProps{Title: title}), textPane("body")))
}

func (s *scenarioState) renderPanelWithKbdAction(title, kbd string) error {
	return s.render(withChildren(
		baud.Panel(baud.PanelProps{Title: title, Actions: baud.Kbd(kbd)}),
		textPane("body"),
	))
}

func (s *scenarioState) renderPanelBodyBlocks(title string, n int) error {
	return s.render(withChildren(baud.Panel(baud.PanelProps{Title: title}), templ.Join(panes(n)...)))
}

func (s *scenarioState) renderPanelIDClass(title, id, class string) error {
	return s.render(withChildren(
		baud.Panel(baud.PanelProps{Title: title, ID: id, Class: class}),
		textPane("body"),
	))
}

func (s *scenarioState) renderStatusBarCells(texts string) error {
	return s.render(baud.StatusBar(panelStatusCells(texts)))
}

func (s *scenarioState) renderStatusBarMode(mode, texts string) error {
	cells := append([]baud.StatusCell{{Text: mode, Mode: true}}, panelStatusCells(texts)...)
	return s.render(baud.StatusBar(cells))
}

func (s *scenarioState) renderStatusBarSpring(left, right string) error {
	cells := append(panelStatusCells(left), baud.StatusCell{Spring: true})
	cells = append(cells, panelStatusCells(right)...)
	return s.render(baud.StatusBar(cells))
}

func (s *scenarioState) renderToolbarBtnInput(label string) error {
	return s.render(withChildren(baud.Toolbar(), templ.Join(
		baud.Btn(baud.BtnProps{Label: label}),
		baud.Input(baud.InputProps{Placeholder: "filter"}),
	)))
}

// panelStatusCells maps a comma-separated list to plain StatusBar cells.
func panelStatusCells(texts string) []baud.StatusCell {
	var cells []baud.StatusCell
	for _, t := range strings.Split(texts, ",") {
		cells = append(cells, baud.StatusCell{Text: t})
	}
	return cells
}

// registerPanelSteps wires the Panel/StatusBar/Toolbar steps into the suite.
func registerPanelSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a panel titled "([^"]*)"$`, s.renderPanelTitled)
	sc.When(`^I render a panel titled "([^"]*)" with a kbd action "([^"]*)"$`, s.renderPanelWithKbdAction)
	sc.When(`^I render a panel titled "([^"]*)" with (\d+) body blocks$`, s.renderPanelBodyBlocks)
	sc.When(`^I render a panel titled "([^"]*)" with id "([^"]*)" and class "([^"]*)"$`, s.renderPanelIDClass)
	sc.When(`^I render a status bar with cells "([^"]*)"$`, s.renderStatusBarCells)
	sc.When(`^I render a status bar with mode "([^"]*)" and cells "([^"]*)"$`, s.renderStatusBarMode)
	sc.When(`^I render a status bar with cells "([^"]*)" then a spring then cells "([^"]*)"$`, s.renderStatusBarSpring)
	sc.When(`^I render a toolbar of a button "([^"]*)" and an input$`, s.renderToolbarBtnInput)
}

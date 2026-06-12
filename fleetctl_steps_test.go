package baudui_test

// fleetctl console step definitions — features/fleetctl.feature. The
// console is a demo-package composition, so these steps render demo
// templates (AppPage and the exported overlay/tab fragments) instead of
// bare baud components. Component-prefixed helpers only; generic
// assertions live in steps_test.go.

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/demo"
)

func (s *scenarioState) renderFleetctlConsole() error {
	return s.render(demo.AppPage(demo.ServerOpts()))
}

func (s *scenarioState) renderFleetctlConsoleStatic() error {
	return s.render(demo.AppPage(demo.StaticOpts()))
}

func (s *scenarioState) renderFleetctlHostDrawer(name string) error {
	c, ok := demo.FleetHostDrawer(name)
	if !ok {
		return fmt.Errorf("unknown fleet service %q", name)
	}
	return s.render(c)
}

func (s *scenarioState) renderFleetctlDeployModal() error {
	return s.render(demo.FleetDeployModal())
}

func (s *scenarioState) renderFleetctlTabPane(view string) error {
	c, ok := demo.FleetTabPane(view)
	if !ok {
		return fmt.Errorf("unknown fleet tab view %q", view)
	}
	return s.render(c)
}

// fleetctlAttrContains asserts a substring of one element's attribute —
// for the console's hyperscript listeners, where the full program text
// is an implementation detail but the wired command ids are contract.
func (s *scenarioState) fleetctlAttrContains(selector, name, substr string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	got, ok := attr(n, name)
	if !ok {
		return fmt.Errorf("element %q has no attribute %q", selector, name)
	}
	if !strings.Contains(got, substr) {
		return fmt.Errorf("element %q attribute %q = %q, want substring %q", selector, name, got, substr)
	}
	return nil
}

// fleetctlHasCellWithText asserts that at least one element matching the
// selector carries exactly the given text — for repeated cells (statusbar)
// where the generic unique-element text step cannot apply.
func (s *scenarioState) fleetctlHasCellWithText(selector, want string) error {
	nodes := s.matching(selector)
	if len(nodes) == 0 {
		return fmt.Errorf("no element matches %q", selector)
	}
	for _, n := range nodes {
		if strings.TrimSpace(textContent(n)) == want {
			return nil
		}
	}
	return fmt.Errorf("no element matching %q has text %q (%d candidates)", selector, want, len(nodes))
}

func registerFleetctlSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render the fleetctl console$`, s.renderFleetctlConsole)
	sc.When(`^I render the fleetctl console for the static site$`, s.renderFleetctlConsoleStatic)
	sc.When(`^I render the fleetctl host drawer for "([^"]*)"$`, s.renderFleetctlHostDrawer)
	sc.When(`^I render the fleetctl deploy modal$`, s.renderFleetctlDeployModal)
	sc.When(`^I render the fleetctl "([^"]*)" tab pane$`, s.renderFleetctlTabPane)
	sc.Then(`^the fleetctl element "([^"]*)" attribute "([^"]*)" contains "([^"]*)"$`, s.fleetctlAttrContains)
	sc.Then(`^the fleetctl console has a "([^"]*)" with text "([^"]*)"$`, s.fleetctlHasCellWithText)
}

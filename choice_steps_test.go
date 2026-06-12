package baudui_test

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"

	"github.com/markopolo123/baud-ui/baud"
)

// registerChoiceSteps wires the Checkbox/Radio/Toggle scenario steps onto
// the shared scenario state. Registered with one line in InitializeScenario
// (steps_test.go) — the only file this feature shares with other waves.
func registerChoiceSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a checkbox named "([^"]*)" with value "([^"]*)" and label "([^"]*)"$`, s.renderCheckbox)
	sc.When(`^I render a checked checkbox named "([^"]*)" with label "([^"]*)"$`, s.renderCheckedCheckbox)
	sc.When(`^I render a disabled checkbox named "([^"]*)" with label "([^"]*)"$`, s.renderDisabledCheckbox)
	sc.When(`^I render a checked disabled checkbox named "([^"]*)" with label "([^"]*)"$`, s.renderCheckedDisabledCheckbox)
	sc.When(`^I render a checkbox with id "([^"]*)" named "([^"]*)"$`, s.renderCheckboxWithID)
	sc.When(`^I render a radio named "([^"]*)" with value "([^"]*)" and label "([^"]*)"$`, s.renderRadio)
	sc.When(`^I render a disabled radio named "([^"]*)" with value "([^"]*)" and label "([^"]*)"$`, s.renderDisabledRadio)
	sc.When(`^I render a radio group named "([^"]*)" with options "([^"]*)" selecting "([^"]*)"$`, s.renderRadioGroup)
	sc.When(`^I render a toggle named "([^"]*)" with options "([^"]*)" selecting "([^"]*)"$`, s.renderToggle)
	sc.When(`^I render a toggle named "([^"]*)" with options "([^"]*)" selecting "([^"]*)" disabling "([^"]*)"$`, s.renderToggleDisabling)

	sc.Then(`^the element "([^"]*)" has text "([^"]*)"$`, s.elementHasText)
}

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderCheckbox(name, value, label string) error {
	return s.render(baud.Checkbox(baud.ChoiceProps{Name: name, Value: value, Label: label}))
}

func (s *scenarioState) renderCheckedCheckbox(name, label string) error {
	return s.render(baud.Checkbox(baud.ChoiceProps{Name: name, Label: label, Checked: true}))
}

func (s *scenarioState) renderDisabledCheckbox(name, label string) error {
	return s.render(baud.Checkbox(baud.ChoiceProps{Name: name, Label: label, Disabled: true}))
}

func (s *scenarioState) renderCheckedDisabledCheckbox(name, label string) error {
	return s.render(baud.Checkbox(baud.ChoiceProps{Name: name, Label: label, Checked: true, Disabled: true}))
}

func (s *scenarioState) renderCheckboxWithID(id, name string) error {
	return s.render(baud.Checkbox(baud.ChoiceProps{ID: id, Name: name, Label: name}))
}

func (s *scenarioState) renderRadio(name, value, label string) error {
	return s.render(baud.Radio(baud.ChoiceProps{Name: name, Value: value, Label: label}))
}

func (s *scenarioState) renderDisabledRadio(name, value, label string) error {
	return s.render(baud.Radio(baud.ChoiceProps{Name: name, Value: value, Label: label, Disabled: true}))
}

func (s *scenarioState) renderRadioGroup(name, options, selected string) error {
	var group []templ.Component
	for _, v := range strings.Split(options, ",") {
		group = append(group, baud.Radio(baud.ChoiceProps{
			Name:    name,
			Value:   v,
			Label:   v,
			Checked: v == selected,
		}))
	}
	return s.render(templ.Join(group...))
}

func (s *scenarioState) renderToggle(name, options, selected string) error {
	return s.renderToggleDisabling(name, options, selected, "")
}

func (s *scenarioState) renderToggleDisabling(name, options, selected, disabled string) error {
	var opts []baud.ToggleOption
	for _, v := range strings.Split(options, ",") {
		opts = append(opts, baud.ToggleOption{Value: v, Disabled: v == disabled && disabled != ""})
	}
	return s.render(baud.Toggle(baud.ToggleProps{Name: name, Value: selected, Options: opts}))
}

// ---- assertion steps -----------------------------------------------------

func (s *scenarioState) elementHasText(selector, want string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	var b strings.Builder
	walk(n, func(c *html.Node) {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	})
	if got := strings.TrimSpace(b.String()); got != want {
		return fmt.Errorf("element %q text = %q, want %q", selector, got, want)
	}
	return nil
}

package baudui_test

import (
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// registerTagInputSteps wires the TagInput scenario steps onto the shared
// scenario state. Registered with one line in InitializeScenario
// (steps_test.go) — the only file this feature shares with other waves.
func registerTagInputSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a tag input named "([^"]*)" with values "([^"]*)" and placeholder "([^"]*)"$`, s.renderTagInputValuesPlaceholder)
	sc.When(`^I render a tag input named "([^"]*)" with values "([^"]*)"$`, s.renderTagInputValues)
	sc.When(`^I render a tag input named "([^"]*)" with suggestions "([^"]*)"$`, s.renderTagInputSuggestions)
	sc.When(`^I render a tag input named "([^"]*)" with placeholder "([^"]*)"$`, s.renderTagInputPlaceholder)
	sc.When(`^I render a tag input with id "([^"]*)" named "([^"]*)"$`, s.renderTagInputWithID)
}

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderTagInputValues(name, values string) error {
	return s.render(baud.TagInput(baud.TagInputProps{
		Name:   name,
		Values: strings.Split(values, ","),
	}))
}

func (s *scenarioState) renderTagInputValuesPlaceholder(name, values, placeholder string) error {
	return s.render(baud.TagInput(baud.TagInputProps{
		Name:        name,
		Values:      strings.Split(values, ","),
		Placeholder: placeholder,
	}))
}

func (s *scenarioState) renderTagInputSuggestions(name, suggestions string) error {
	return s.render(baud.TagInput(baud.TagInputProps{
		Name:        name,
		Suggestions: strings.Split(suggestions, ","),
	}))
}

func (s *scenarioState) renderTagInputPlaceholder(name, placeholder string) error {
	return s.render(baud.TagInput(baud.TagInputProps{
		Name:        name,
		Placeholder: placeholder,
	}))
}

func (s *scenarioState) renderTagInputWithID(id, name string) error {
	return s.render(baud.TagInput(baud.TagInputProps{ID: id, Name: name}))
}

package baudui_test

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// registerSelectSteps wires the Select/SelectMenu/Combobox scenario steps
// onto the shared scenario state. Registered with one line in
// InitializeScenario (steps_test.go) — the only file this feature shares
// with other waves.
func registerSelectSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a select named "([^"]*)" with options "([^"]*)"$`, s.renderSelect)
	sc.When(`^I render a select named "([^"]*)" with options "([^"]*)" selecting "([^"]*)"$`, s.renderSelectSelecting)
	sc.When(`^I render a disabled select named "([^"]*)" with options "([^"]*)"$`, s.renderDisabledSelect)
	sc.When(`^I render a select named "([^"]*)" with options "([^"]*)" disabling "([^"]*)"$`, s.renderSelectDisabling)
	sc.When(`^I render a select with id "([^"]*)" named "([^"]*)" and options "([^"]*)"$`, s.renderSelectWithID)
	sc.When(`^I render a select named "([^"]*)" with options "([^"]*)" sized "([^"]*)"$`, s.renderSelectSized)
	sc.When(`^I render a select menu with id "([^"]*)" named "([^"]*)" and options "([^"]*)" selecting "([^"]*)"$`, s.renderSelectMenu)
	sc.When(`^I render a select menu with id "([^"]*)" and options "([^"]*)" with placeholder "([^"]*)"$`, s.renderSelectMenuPlaceholder)
	sc.When(`^I render a disabled select menu with id "([^"]*)" named "([^"]*)" and options "([^"]*)"$`, s.renderDisabledSelectMenu)
	sc.When(`^I render a right-aligned select menu with id "([^"]*)" named "([^"]*)" and options "([^"]*)"$`, s.renderRightAlignedSelectMenu)
	sc.When(`^I render a combobox with id "([^"]*)" named "([^"]*)" and options "([^"]*)"$`, s.renderCombobox)
	sc.When(`^I render a combobox with id "([^"]*)" named "([^"]*)" and options "([^"]*)" selecting "([^"]*)"$`, s.renderComboboxSelecting)
	sc.When(`^I render a server combobox with id "([^"]*)" named "([^"]*)" searching "([^"]*)"$`, s.renderServerCombobox)
	sc.When(`^I render combobox options for id "([^"]*)" matching "([^"]*)" from "([^"]*)"$`, s.renderComboboxOptions)

	sc.Then(`^each option has an id derived from "([^"]*)"$`, s.selectOptionIDsDerived)
}

// selectParseOptions parses an options spec: comma-separated values, each
// optionally carrying a right-aligned meta column after "=", e.g.
// "ingest-gw-1=eu-west,api-core=us-east,scratch".
func selectParseOptions(spec string) []baud.SelectOption {
	var opts []baud.SelectOption
	for _, part := range strings.Split(spec, ",") {
		value, meta, _ := strings.Cut(part, "=")
		opts = append(opts, baud.SelectOption{Value: value, Meta: meta})
	}
	return opts
}

// ---- render steps: Select (native baseline) ------------------------------

func (s *scenarioState) renderSelect(name, options string) error {
	return s.render(baud.Select(baud.SelectProps{Name: name, Options: selectParseOptions(options)}))
}

func (s *scenarioState) renderSelectSelecting(name, options, value string) error {
	return s.render(baud.Select(baud.SelectProps{Name: name, Value: value, Options: selectParseOptions(options)}))
}

func (s *scenarioState) renderDisabledSelect(name, options string) error {
	return s.render(baud.Select(baud.SelectProps{Name: name, Disabled: true, Options: selectParseOptions(options)}))
}

func (s *scenarioState) renderSelectDisabling(name, options, disabled string) error {
	opts := selectParseOptions(options)
	for i := range opts {
		opts[i].Disabled = opts[i].Value == disabled
	}
	return s.render(baud.Select(baud.SelectProps{Name: name, Options: opts}))
}

func (s *scenarioState) renderSelectWithID(id, name, options string) error {
	return s.render(baud.Select(baud.SelectProps{ID: id, Name: name, Options: selectParseOptions(options)}))
}

func (s *scenarioState) renderSelectSized(name, options, width string) error {
	return s.render(baud.Select(baud.SelectProps{Name: name, Width: width, Options: selectParseOptions(options)}))
}

// ---- render steps: SelectMenu (custom trigger + listbox) -----------------

func (s *scenarioState) renderSelectMenu(id, name, options, value string) error {
	return s.render(baud.SelectMenu(baud.SelectMenuProps{
		ID: id, Name: name, Value: value, Options: selectParseOptions(options),
	}))
}

func (s *scenarioState) renderSelectMenuPlaceholder(id, options, placeholder string) error {
	return s.render(baud.SelectMenu(baud.SelectMenuProps{
		ID: id, Placeholder: placeholder, Options: selectParseOptions(options),
	}))
}

func (s *scenarioState) renderDisabledSelectMenu(id, name, options string) error {
	return s.render(baud.SelectMenu(baud.SelectMenuProps{
		ID: id, Name: name, Disabled: true, Options: selectParseOptions(options),
	}))
}

func (s *scenarioState) renderRightAlignedSelectMenu(id, name, options string) error {
	return s.render(baud.SelectMenu(baud.SelectMenuProps{
		ID: id, Name: name, AlignRight: true, Options: selectParseOptions(options),
	}))
}

// ---- render steps: Combobox ----------------------------------------------

func (s *scenarioState) renderCombobox(id, name, options string) error {
	return s.render(baud.Combobox(baud.ComboboxProps{
		ID: id, Name: name, Options: selectParseOptions(options),
	}))
}

func (s *scenarioState) renderComboboxSelecting(id, name, options, value string) error {
	return s.render(baud.Combobox(baud.ComboboxProps{
		ID: id, Name: name, Value: value, Options: selectParseOptions(options),
	}))
}

func (s *scenarioState) renderServerCombobox(id, name, url string) error {
	return s.render(baud.Combobox(baud.ComboboxProps{
		ID: id, Name: name, SearchURL: url,
	}))
}

func (s *scenarioState) renderComboboxOptions(id, query, options string) error {
	return s.render(baud.ComboboxOptions(baud.ComboboxOptionsProps{
		ID: id, Query: query, Options: selectParseOptions(options),
	}))
}

// ---- assertion steps ------------------------------------------------------

// selectOptionIDsDerived asserts every option id is prefix-opt-<index>, in
// document order — the contract aria-activedescendant relies on.
func (s *scenarioState) selectOptionIDsDerived(prefix string) error {
	opts := s.matching("button.menu-item")
	if len(opts) == 0 {
		return fmt.Errorf("no button.menu-item options found")
	}
	for i, n := range opts {
		want := fmt.Sprintf("%s-opt-%d", prefix, i)
		if got := attrVal(n, "id"); got != want {
			return fmt.Errorf("option %d id = %q, want %q", i, got, want)
		}
	}
	return nil
}

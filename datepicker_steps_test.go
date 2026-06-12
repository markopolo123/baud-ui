package baudui_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// registerDatePickerSteps wires the DatePicker scenario steps onto the
// shared scenario state. Registered with one line in InitializeScenario
// (steps_test.go) — the only file this feature shares with other waves.
func registerDatePickerSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a date picker named "([^"]*)" selecting "([^"]*)" with today "([^"]*)" via "([^"]*)"$`, s.datepickerRender)
	sc.When(`^I render a date picker named "([^"]*)" with no selection and today "([^"]*)" via "([^"]*)"$`, s.datepickerRenderEmpty)
	sc.When(`^I render the date picker menu for month "([^"]*)" selecting "([^"]*)" with today "([^"]*)" via "([^"]*)"$`, s.datepickerRenderMenu)
	sc.When(`^I render the date picker menu for month "([^"]*)" with today "([^"]*)" via "([^"]*)"$`, s.datepickerRenderMenuNoSelection)

	sc.Then(`^the day cells run from "([^"]*)" to "([^"]*)"$`, s.datepickerDayCellsRun)
	sc.Then(`^the day-of-week header reads "([^"]*)"$`, s.datepickerDowHeader)
}

// ---- render steps --------------------------------------------------------

func (s *scenarioState) datepickerRender(name, selected, today, endpoint string) error {
	sel, err := time.Parse("2006-01-02", selected)
	if err != nil {
		return err
	}
	tod, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}
	return s.render(baud.DatePicker(baud.DatePickerProps{
		Name: name, Selected: sel, Today: tod, Endpoint: endpoint,
	}))
}

func (s *scenarioState) datepickerRenderEmpty(name, today, endpoint string) error {
	tod, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}
	return s.render(baud.DatePicker(baud.DatePickerProps{
		Name: name, Today: tod, Endpoint: endpoint,
	}))
}

func (s *scenarioState) datepickerRenderMenu(month, selected, today, endpoint string) error {
	m, err := time.Parse("2006-01", month)
	if err != nil {
		return err
	}
	sel, err := time.Parse("2006-01-02", selected)
	if err != nil {
		return err
	}
	tod, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}
	return s.render(baud.DatePickerMenu(baud.DatePickerProps{
		Month: m, Selected: sel, Today: tod, Endpoint: endpoint,
	}))
}

func (s *scenarioState) datepickerRenderMenuNoSelection(month, today, endpoint string) error {
	m, err := time.Parse("2006-01", month)
	if err != nil {
		return err
	}
	tod, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}
	return s.render(baud.DatePickerMenu(baud.DatePickerProps{
		Month: m, Today: tod, Endpoint: endpoint,
	}))
}

// ---- assertion steps -----------------------------------------------------

// datepickerDayCellsRun asserts the first and last day-cell dates — combined
// with the 42-cell count and the Mo-first header this pins the whole grid.
func (s *scenarioState) datepickerDayCellsRun(first, last string) error {
	cells := s.matching("button.dp-day")
	if len(cells) == 0 {
		return fmt.Errorf("no day cells rendered")
	}
	if got := attrVal(cells[0], "data-date"); got != first {
		return fmt.Errorf("first day cell is %q, want %q", got, first)
	}
	if got := attrVal(cells[len(cells)-1], "data-date"); got != last {
		return fmt.Errorf("last day cell is %q, want %q", got, last)
	}
	return nil
}

func (s *scenarioState) datepickerDowHeader(want string) error {
	var labels []string
	for _, n := range s.matching("span.dp-dow") {
		labels = append(labels, strings.TrimSpace(textContent(n)))
	}
	if got := strings.Join(labels, " "); got != want {
		return fmt.Errorf("day-of-week header reads %q, want %q", got, want)
	}
	return nil
}

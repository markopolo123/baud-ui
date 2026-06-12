package baudui_test

import (
	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// ---- render steps ----------------------------------------------------------

func (s *scenarioState) renderToastWithBody(tone, title, body string) error {
	return s.render(baud.Toast(baud.ToastProps{Tone: tone, Title: title, Body: body}))
}

func (s *scenarioState) renderToast(tone, title string) error {
	return s.render(baud.Toast(baud.ToastProps{Tone: tone, Title: title}))
}

func (s *scenarioState) renderToastNoTone(title string) error {
	return s.render(baud.Toast(baud.ToastProps{Title: title}))
}

func (s *scenarioState) renderToastDismissMs(tone, title string, ms int) error {
	return s.render(baud.Toast(baud.ToastProps{Tone: tone, Title: title, DismissMs: ms}))
}

func (s *scenarioState) renderStickyToast(tone, title string) error {
	return s.render(baud.Toast(baud.ToastProps{Tone: tone, Title: title, Sticky: true}))
}

func (s *scenarioState) renderToastOOB(tone, title string) error {
	return s.render(baud.ToastOOB(baud.ToastProps{Tone: tone, Title: title}))
}

// registerToastSteps wires the Toast steps into the shared suite.
func registerToastSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a toast with tone "([^"]*)" titled "([^"]*)" with body "([^"]*)"$`, s.renderToastWithBody)
	sc.When(`^I render a toast with tone "([^"]*)" titled "([^"]*)"$`, s.renderToast)
	sc.When(`^I render a toast with no tone titled "([^"]*)"$`, s.renderToastNoTone)
	sc.When(`^I render a toast with tone "([^"]*)" titled "([^"]*)" dismissing after (\d+) ms$`, s.renderToastDismissMs)
	sc.When(`^I render a sticky toast with tone "([^"]*)" titled "([^"]*)"$`, s.renderStickyToast)
	sc.When(`^I render an OOB toast with tone "([^"]*)" titled "([^"]*)"$`, s.renderToastOOB)
}

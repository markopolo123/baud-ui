package baudui_test

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// feedbackPanelState maps a scenario kind to canned PanelState props so the
// assertions have stable copy to match on.
func feedbackPanelState(kind string) baud.PanelStateProps {
	switch kind {
	case "skeleton":
		return baud.PanelStateProps{Kind: baud.StateSkeleton}
	case "empty":
		return baud.PanelStateProps{
			Kind:  baud.StateEmpty,
			Title: "no units",
			Sub:   "try widening the filter",
		}
	case "error":
		return baud.PanelStateProps{
			Kind:  baud.StateError,
			Title: "fetch failed",
			Sub:   "upstream timed out",
		}
	default:
		return baud.PanelStateProps{Kind: baud.PanelStateKind(kind)}
	}
}

// ---- render steps ---------------------------------------------------------

func (s *scenarioState) renderProgress(value int) error {
	return s.render(baud.Progress(baud.ProgressProps{Value: float64(value)}))
}

func (s *scenarioState) renderProgressForcedTone(value int, tone string) error {
	return s.render(baud.Progress(baud.ProgressProps{Value: float64(value), Tone: tone}))
}

func (s *scenarioState) renderProgressLabelled(value int, label string, chars int) error {
	return s.render(baud.Progress(baud.ProgressProps{
		Value: float64(value),
		Label: label,
		Chars: chars,
	}))
}

func (s *scenarioState) renderSpinner() error {
	return s.render(baud.Spinner())
}

func (s *scenarioState) renderPanelState(kind string) error {
	return s.render(baud.PanelState(feedbackPanelState(kind)))
}

func (s *scenarioState) renderPanelStateWithAction(kind string) error {
	p := feedbackPanelState(kind)
	p.Action = textPane("clear filters")
	return s.render(baud.PanelState(p))
}

func (s *scenarioState) renderLoadingState(hint string) error {
	return s.render(baud.PanelState(baud.PanelStateProps{
		Kind:  baud.StateLoading,
		Title: "loading units",
		Sub:   hint,
	}))
}

func (s *scenarioState) renderConfirmInput(expect, action string) error {
	return s.render(baud.ConfirmInput(baud.ConfirmInputProps{
		Expect:      expect,
		ActionLabel: action,
	}))
}

func (s *scenarioState) renderNamedConfirmInput(expect, name string) error {
	return s.render(baud.ConfirmInput(baud.ConfirmInputProps{
		Expect:      expect,
		ActionLabel: "decommission",
		Name:        name,
	}))
}

// ---- assertion steps ------------------------------------------------------

// confirmComparisonWired asserts the typed-value guard is purely-local
// hyperscript inline on the input: the _ script must react to input events,
// compare against the element's own data-confirm attribute, and toggle the
// action button's disabled attribute.
func (s *scenarioState) confirmComparisonWired() error {
	n, err := s.one("div.confirm input.input")
	if err != nil {
		return err
	}
	script, ok := attr(n, "_")
	if !ok {
		return fmt.Errorf("confirm input has no inline hyperscript (_ attribute)")
	}
	for _, want := range []string{"on input", "@data-confirm", "@disabled"} {
		if !strings.Contains(script, want) {
			return fmt.Errorf("confirm input hyperscript %q is missing %q", script, want)
		}
	}
	return nil
}

// registerFeedbackSteps wires the feedback-group steps into the shared suite.
func registerFeedbackSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a progress at (-?\d+)$`, s.renderProgress)
	sc.When(`^I render a progress at (-?\d+) with forced tone "([^"]*)"$`, s.renderProgressForcedTone)
	sc.When(`^I render a progress at (-?\d+) labelled "([^"]*)" with (\d+) chars$`, s.renderProgressLabelled)
	sc.When(`^I render a spinner$`, s.renderSpinner)
	sc.When(`^I render an? "([^"]*)" panel state$`, s.renderPanelState)
	sc.When(`^I render an? "([^"]*)" panel state with an action$`, s.renderPanelStateWithAction)
	sc.When(`^I render a loading panel state hinting "([^"]*)"$`, s.renderLoadingState)
	sc.When(`^I render a confirm input expecting "([^"]*)" with action "([^"]*)"$`, s.renderConfirmInput)
	sc.When(`^I render a confirm input expecting "([^"]*)" named "([^"]*)"$`, s.renderNamedConfirmInput)

	sc.Then(`^the confirm input compares its value against data-confirm via hyperscript$`, s.confirmComparisonWired)
}

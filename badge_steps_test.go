package baudui_test

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// withChild renders a component with a templ children block injected —
// how badge labels reach { children... } from plain Go.
type withChild struct {
	c, child templ.Component
}

func (w withChild) Render(ctx context.Context, wr io.Writer) error {
	return w.c.Render(templ.WithChildren(ctx, w.child), wr)
}

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderBadge(tone, variant, label string) error {
	return s.render(withChild{baud.Badge(baud.BadgeProps{Tone: tone, Variant: variant}), textPane(label)})
}

func (s *scenarioState) renderDefaultBadge(label string) error {
	return s.render(withChild{baud.Badge(baud.BadgeProps{}), textPane(label)})
}

func (s *scenarioState) renderBadgeWithDot(tone, variant, label string) error {
	return s.render(withChild{baud.Badge(baud.BadgeProps{Tone: tone, Variant: variant, Dot: true}), textPane(label)})
}

func (s *scenarioState) renderDot(tone string) error {
	return s.render(baud.Dot(baud.DotProps{Tone: tone}))
}

func (s *scenarioState) renderDefaultDot() error {
	return s.render(baud.Dot(baud.DotProps{}))
}

func (s *scenarioState) renderPulsingDot(tone string) error {
	return s.render(baud.Dot(baud.DotProps{Tone: tone, Pulse: true}))
}

// ---- assertion steps -----------------------------------------------------

// registerBadgeSteps wires the Badge/Dot steps into the shared suite.
func registerBadgeSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a badge with tone "([^"]*)" and variant "([^"]*)" labelled "([^"]*)"$`, s.renderBadge)
	sc.When(`^I render a badge with tone "([^"]*)" and variant "([^"]*)" and a dot labelled "([^"]*)"$`, s.renderBadgeWithDot)
	sc.When(`^I render a default badge labelled "([^"]*)"$`, s.renderDefaultBadge)
	sc.When(`^I render a dot with tone "([^"]*)"$`, s.renderDot)
	sc.When(`^I render a default dot$`, s.renderDefaultDot)
	sc.When(`^I render a pulsing dot with tone "([^"]*)"$`, s.renderPulsingDot)

}

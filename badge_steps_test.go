package baudui_test

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"

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

// elementHasText asserts the element's collapsed text content.
func (s *scenarioState) elementHasText(selector, want string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	var sb strings.Builder
	walk(n, func(c *html.Node) {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	})
	if got := strings.TrimSpace(sb.String()); got != want {
		return fmt.Errorf("element %q text = %q, want %q", selector, got, want)
	}
	return nil
}

// registerBadgeSteps wires the Badge/Dot steps into the shared suite.
func registerBadgeSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a badge with tone "([^"]*)" and variant "([^"]*)" labelled "([^"]*)"$`, s.renderBadge)
	sc.When(`^I render a badge with tone "([^"]*)" and variant "([^"]*)" and a dot labelled "([^"]*)"$`, s.renderBadgeWithDot)
	sc.When(`^I render a default badge labelled "([^"]*)"$`, s.renderDefaultBadge)
	sc.When(`^I render a dot with tone "([^"]*)"$`, s.renderDot)
	sc.When(`^I render a default dot$`, s.renderDefaultDot)
	sc.When(`^I render a pulsing dot with tone "([^"]*)"$`, s.renderPulsingDot)

	sc.Then(`^the element "([^"]*)" has text "([^"]*)"$`, s.elementHasText)
}

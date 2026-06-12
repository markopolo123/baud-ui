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

// Step definitions for features/btn.feature (Btn / BtnGroup / Kbd).
// Registered from InitializeScenario in steps_test.go.

// ---- render steps -------------------------------------------------------

func (s *scenarioState) renderBtn(label string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label}))
}

func (s *scenarioState) renderVariantBtn(variant, label string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label, Variant: btnVariant(variant)}))
}

func (s *scenarioState) renderDisabledBtn(label string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label, Disabled: true}))
}

func (s *scenarioState) renderActiveBtn(variant, label string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label, Variant: btnVariant(variant), Active: true}))
}

func (s *scenarioState) renderGlyphBtn(label, glyph string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label, Glyph: glyph}))
}

func (s *scenarioState) renderKbdBtn(label, kbd string) error {
	return s.render(baud.Btn(baud.BtnProps{Label: label, Kbd: kbd}))
}

func (s *scenarioState) renderFullBtn(variant, label, glyph, kbd string) error {
	return s.render(baud.Btn(baud.BtnProps{
		Label:   label,
		Variant: btnVariant(variant),
		Glyph:   glyph,
		Kbd:     kbd,
	}))
}

func (s *scenarioState) renderKbdChip(label string) error {
	return s.render(baud.Kbd(label))
}

func (s *scenarioState) renderBtnGroup(labels, active string) error {
	var btns []templ.Component
	for _, l := range strings.Split(labels, ",") {
		btns = append(btns, baud.Btn(baud.BtnProps{Label: l, Active: l == active}))
	}
	return s.render(withChildren(baud.BtnGroup(), templ.Join(btns...)))
}

// btnVariant maps the feature-file wording to the typed variant;
// "default" means the zero value.
func btnVariant(v string) baud.BtnVariant {
	if v == "default" {
		return baud.BtnDefault
	}
	return baud.BtnVariant(v)
}

// withChildren renders a component with the given templ children block.
func withChildren(c, children templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return c.Render(templ.WithChildren(ctx, children), w)
	})
}

// ---- assertion steps ----------------------------------------------------

// elementHasText asserts the element's full text content (trimmed).
func (s *scenarioState) elementHasText(selector, want string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(textContent(n)); got != want {
		return fmt.Errorf("element %q text = %q, want %q", selector, got, want)
	}
	return nil
}

// elementHasOwnText asserts the element's direct text nodes (trimmed),
// ignoring text inside child elements (glyph/kbd chips).
func (s *scenarioState) elementHasOwnText(selector, want string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	if got := strings.TrimSpace(b.String()); got != want {
		return fmt.Errorf("element %q own text = %q, want %q", selector, got, want)
	}
	return nil
}

func (s *scenarioState) firstElementChildIs(selector, childSel string) error {
	return s.edgeElementChildIs(selector, childSel, false)
}

func (s *scenarioState) lastElementChildIs(selector, childSel string) error {
	return s.edgeElementChildIs(selector, childSel, true)
}

func (s *scenarioState) edgeElementChildIs(selector, childSel string, last bool) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	var edge *html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		edge = c
		if !last {
			break
		}
	}
	which := "first"
	if last {
		which = "last"
	}
	if edge == nil {
		return fmt.Errorf("element %q has no element children, want %s child %q", selector, which, childSel)
	}
	sel := parseSimple(childSel)
	if !nodeMatches(edge, sel) {
		return fmt.Errorf("%s element child of %q is <%s class=%q>, want %q", which, selector, edge.Data, attrVal(edge, "class"), childSel)
	}
	return nil
}

func textContent(n *html.Node) string {
	var b strings.Builder
	walk(n, func(c *html.Node) {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	})
	return b.String()
}

// registerBtnSteps wires the Btn/BtnGroup/Kbd steps into the suite.
func registerBtnSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a button "([^"]*)"$`, s.renderBtn)
	sc.When(`^I render a "([^"]*)" button "([^"]*)"$`, s.renderVariantBtn)
	sc.When(`^I render a disabled button "([^"]*)"$`, s.renderDisabledBtn)
	sc.When(`^I render an active "([^"]*)" button "([^"]*)"$`, s.renderActiveBtn)
	sc.When(`^I render a button "([^"]*)" with glyph "([^"]*)"$`, s.renderGlyphBtn)
	sc.When(`^I render a button "([^"]*)" with kbd hint "([^"]*)"$`, s.renderKbdBtn)
	sc.When(`^I render a "([^"]*)" button "([^"]*)" with glyph "([^"]*)" and kbd hint "([^"]*)"$`, s.renderFullBtn)
	sc.When(`^I render a kbd chip "([^"]*)"$`, s.renderKbdChip)
	sc.When(`^I render a button group of buttons "([^"]*)" with active "([^"]*)"$`, s.renderBtnGroup)

	sc.Then(`^the element "([^"]*)" has text "([^"]*)"$`, s.elementHasText)
	sc.Then(`^the element "([^"]*)" has own text "([^"]*)"$`, s.elementHasOwnText)
	sc.Then(`^the first element child of "([^"]*)" is "([^"]*)"$`, s.firstElementChildIs)
	sc.Then(`^the last element child of "([^"]*)" is "([^"]*)"$`, s.lastElementChildIs)
}

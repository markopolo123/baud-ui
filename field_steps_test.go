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

// Step definitions for features/field.feature (Field + Input primitives).
// Registered from InitializeScenario via registerFieldSteps.

// ---- render steps -------------------------------------------------------

func (s *scenarioState) renderInputIDPlaceholder(id, placeholder string) error {
	return s.render(baud.Input(baud.InputProps{ID: id, Placeholder: placeholder}))
}

func (s *scenarioState) renderInputPrefix(prefix string) error {
	return s.render(baud.Input(baud.InputProps{Prefix: prefix}))
}

func (s *scenarioState) renderInputSuffix(suffix string) error {
	return s.render(baud.Input(baud.InputProps{Suffix: suffix}))
}

func (s *scenarioState) renderInputAffixes(prefix, suffix string) error {
	return s.render(baud.Input(baud.InputProps{Prefix: prefix, Suffix: suffix}))
}

func (s *scenarioState) renderDisabledInput() error {
	return s.render(baud.Input(baud.InputProps{Disabled: true}))
}

func (s *scenarioState) renderErrorInput() error {
	return s.render(baud.Input(baud.InputProps{Error: true}))
}

func (s *scenarioState) renderFieldWithHint(label, id, hint string) error {
	return s.render(fieldWith(
		baud.FieldProps{Label: label, For: id, Hint: hint},
		baud.InputProps{ID: id},
	))
}

func (s *scenarioState) renderFieldWithError(label, id, msg string) error {
	return s.render(fieldWith(
		baud.FieldProps{Label: label, For: id, Error: msg},
		baud.InputProps{ID: id, Error: true},
	))
}

// fieldWith composes a Field around an Input child, the canonical pairing.
func fieldWith(fp baud.FieldProps, ip baud.InputProps) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		ctx = templ.WithChildren(ctx, baud.Input(ip))
		return baud.Field(fp).Render(ctx, w)
	})
}

// ---- assertion steps ----------------------------------------------------

func (s *scenarioState) inputPrecededByAffix(text string) error {
	return s.inputAdjacentAffix(text, true)
}

func (s *scenarioState) inputFollowedByAffix(text string) error {
	return s.inputAdjacentAffix(text, false)
}

// inputAdjacentAffix asserts the element sibling immediately before/after
// the .input is a span.affix with the given text.
func (s *scenarioState) inputAdjacentAffix(text string, before bool) error {
	in, err := s.one("input.input")
	if err != nil {
		return err
	}
	sib := elementSibling(in, before)
	dir := "after"
	if before {
		dir = "before"
	}
	if sib == nil {
		return fmt.Errorf("no element sibling %s the input", dir)
	}
	if sib.Data != "span" || !hasClass(sib, "affix") {
		return fmt.Errorf("sibling %s the input is <%s class=%q>, want span.affix", dir, sib.Data, attrVal(sib, "class"))
	}
	if got := strings.TrimSpace(textContent(sib)); got != text {
		return fmt.Errorf("affix %s the input has text %q, want %q", dir, got, text)
	}
	return nil
}

func elementSibling(n *html.Node, before bool) *html.Node {
	step := func(m *html.Node) *html.Node { return m.NextSibling }
	if before {
		step = func(m *html.Node) *html.Node { return m.PrevSibling }
	}
	for m := step(n); m != nil; m = step(m) {
		if m.Type == html.ElementNode {
			return m
		}
	}
	return nil
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attrVal(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// registerFieldSteps wires the Field/Input steps onto the shared state.
func registerFieldSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render an input with id "([^"]*)" and placeholder "([^"]*)"$`, s.renderInputIDPlaceholder)
	sc.When(`^I render an input with prefix "([^"]*)" and suffix "([^"]*)"$`, s.renderInputAffixes)
	sc.When(`^I render an input with prefix "([^"]*)"$`, s.renderInputPrefix)
	sc.When(`^I render an input with suffix "([^"]*)"$`, s.renderInputSuffix)
	sc.When(`^I render a disabled input$`, s.renderDisabledInput)
	sc.When(`^I render an input in the error state$`, s.renderErrorInput)
	sc.When(`^I render a field labelled "([^"]*)" for id "([^"]*)" with hint "([^"]*)"$`, s.renderFieldWithHint)
	sc.When(`^I render a field labelled "([^"]*)" for id "([^"]*)" with error "([^"]*)"$`, s.renderFieldWithError)

	sc.Then(`^the input is immediately preceded by an affix with text "([^"]*)"$`, s.inputPrecededByAffix)
	sc.Then(`^the input is immediately followed by an affix with text "([^"]*)"$`, s.inputFollowedByAffix)
}

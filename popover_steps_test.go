package baudui_test

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// registerPopoverSteps wires the Popover/Tooltip scenario steps onto the
// shared scenario state. Registered with one line in InitializeScenario
// (steps_test.go) — the only file this feature shares with other waves.
func registerPopoverSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a popover with id "([^"]*)" labelled "([^"]*)"$`, s.renderPopover)
	sc.When(`^I render a popover with id "([^"]*)" labelled "([^"]*)" containing "([^"]*)"$`, s.renderPopoverContaining)
	sc.When(`^I render a popover with id "([^"]*)" labelled "([^"]*)" titled "([^"]*)"$`, s.renderPopoverTitled)
	sc.When(`^I render a popover with id "([^"]*)" labelled "([^"]*)" with glyph "([^"]*)"$`, s.renderPopoverGlyph)
	sc.When(`^I render an open popover with id "([^"]*)" labelled "([^"]*)"$`, s.renderOpenPopover)
	sc.When(`^I render a disabled popover with id "([^"]*)" labelled "([^"]*)"$`, s.renderDisabledPopover)
	sc.When(`^I render a tooltip host with tip "([^"]*)"$`, s.renderTooltipHost)
	sc.When(`^I render an under-styled tooltip host with tip "([^"]*)"$`, s.renderTooltipUnderHost)

	sc.Then(`^the popover wrap installs the MenuDismiss behavior$`, s.popoverWrapInstallsMenuDismiss)
	sc.Then(`^the popover wrap syncs aria-expanded when MenuDismiss closes it$`, s.popoverWrapSyncsAria)
	sc.Then(`^the popover trigger toggles the open class with aria-expanded in step$`, s.popoverTriggerToggles)
	sc.Then(`^the tooltip host preserves the multi-line tip "([^"]*)"$`, s.tooltipMultilinePreserved)
}

// ---- render steps: Popover ------------------------------------------------

func (s *scenarioState) renderPopoverProps(p baud.PopoverProps, body templ.Component) error {
	c := baud.Popover(p)
	if body != nil {
		return s.render(templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return c.Render(templ.WithChildren(ctx, body), w)
		}))
	}
	return s.render(c)
}

func (s *scenarioState) renderPopover(id, label string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label}, nil)
}

func (s *scenarioState) renderPopoverContaining(id, label, content string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label}, textPane(content))
}

func (s *scenarioState) renderPopoverTitled(id, label, title string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label, Title: title}, nil)
}

func (s *scenarioState) renderPopoverGlyph(id, label, glyph string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label, Glyph: glyph}, nil)
}

func (s *scenarioState) renderOpenPopover(id, label string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label, Open: true}, nil)
}

func (s *scenarioState) renderDisabledPopover(id, label string) error {
	return s.renderPopoverProps(baud.PopoverProps{ID: id, Label: label, Disabled: true}, nil)
}

// ---- render steps: Tooltip ------------------------------------------------

// tooltipHost spreads the attribute-helper output onto a plain span via
// templ's real attribute renderer, so escaping (incl. multi-line tips)
// is exercised exactly as a .templ spread would.
func tooltipHost(attrs templ.Attributes) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "<span"); err != nil {
			return err
		}
		if err := templ.RenderAttributes(ctx, w, attrs); err != nil {
			return err
		}
		_, err := io.WriteString(w, ">tip host</span>")
		return err
	})
}

// tooltipUnescape turns the feature file's literal \n escapes into real
// newlines (Gherkin step arguments cannot carry raw newlines).
func tooltipUnescape(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}

func (s *scenarioState) renderTooltipHost(tip string) error {
	return s.render(tooltipHost(baud.Tip(tooltipUnescape(tip))))
}

func (s *scenarioState) renderTooltipUnderHost(tip string) error {
	return s.render(tooltipHost(baud.TipUnder(tooltipUnescape(tip))))
}

// ---- assertion steps --------------------------------------------------------

func (s *scenarioState) popoverWrapHSContains(want, why string) error {
	n, err := s.one("span.pop-wrap")
	if err != nil {
		return err
	}
	if hs := attrVal(n, "_"); !strings.Contains(hs, want) {
		return fmt.Errorf("popover wrap _ attribute %q does not contain %q (%s)", hs, want, why)
	}
	return nil
}

func (s *scenarioState) popoverWrapInstallsMenuDismiss() error {
	return s.popoverWrapHSContains("install MenuDismiss", "outside click / Esc dismissal is the shared behavior")
}

// popoverWrapSyncsAria asserts the wrap carries the two MenuDismiss-mirroring
// handlers that drop aria-expanded when .open is dropped for it.
func (s *scenarioState) popoverWrapSyncsAria() error {
	for _, want := range []string{
		"on keyup[key == 'Escape'] from window",
		"on click from elsewhere",
		"call ctl.setAttribute('aria-expanded', 'false')",
	} {
		if err := s.popoverWrapHSContains(want, "aria-expanded must follow MenuDismiss closes"); err != nil {
			return err
		}
	}
	return nil
}

func (s *scenarioState) popoverTriggerToggles() error {
	n, err := s.one("button.pop-trigger")
	if err != nil {
		return err
	}
	hs := attrVal(n, "_")
	for _, want := range []string{
		"toggle .open on root",
		"set @aria-expanded to 'true'",
		"set @aria-expanded to 'false'",
	} {
		if !strings.Contains(hs, want) {
			return fmt.Errorf("popover trigger _ attribute %q does not contain %q", hs, want)
		}
	}
	return nil
}

// tooltipMultilinePreserved asserts data-tip survives rendering + parsing
// with its newlines intact — the contract white-space: pre relies on.
func (s *scenarioState) tooltipMultilinePreserved(tip string) error {
	want := tooltipUnescape(tip)
	if !strings.Contains(want, "\n") {
		return fmt.Errorf("scenario bug: %q has no newline, nothing multi-line to prove", tip)
	}
	n, err := s.one("span[data-tip]")
	if err != nil {
		return err
	}
	if got := attrVal(n, "data-tip"); got != want {
		return fmt.Errorf("data-tip = %q, want %q with newlines preserved", got, want)
	}
	return nil
}

package baudui_test

// Overlay (Modal/Drawer) step definitions — features/overlay.feature.
// Component-prefixed helpers only; generic assertions live in steps_test.go.

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"

	"github.com/markopolo123/baud-ui/baud"
)

// overlayWithBody renders an overlay component with the fixed "body content"
// child so scenarios can assert the body slot.
func overlayWithBody(c templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return c.Render(templ.WithChildren(ctx, textPane("body content")), w)
	})
}

func (s *scenarioState) renderOverlayModal(id, title string) error {
	return s.render(overlayWithBody(baud.Modal(baud.ModalProps{ID: id, Title: title})))
}

func (s *scenarioState) renderOverlayStaticModal(id, title string) error {
	return s.render(overlayWithBody(baud.Modal(baud.ModalProps{ID: id, Title: title, Static: true})))
}

func (s *scenarioState) renderOverlayModalWithFooter(id, title string) error {
	return s.render(overlayWithBody(baud.Modal(baud.ModalProps{
		ID:     id,
		Title:  title,
		Footer: textPane("apply"),
	})))
}

func (s *scenarioState) renderOverlayDrawer(id, title string) error {
	return s.render(overlayWithBody(baud.Drawer(baud.DrawerProps{ID: id, Title: title})))
}

func (s *scenarioState) renderOverlayStaticDrawer(id, title string) error {
	return s.render(overlayWithBody(baud.Drawer(baud.DrawerProps{ID: id, Title: title, Static: true})))
}

func registerOverlaySteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a modal with id "([^"]*)" titled "([^"]*)"$`, s.renderOverlayModal)
	sc.When(`^I render a static modal with id "([^"]*)" titled "([^"]*)"$`, s.renderOverlayStaticModal)
	sc.When(`^I render a modal with footer actions with id "([^"]*)" titled "([^"]*)"$`, s.renderOverlayModalWithFooter)
	sc.When(`^I render a drawer with id "([^"]*)" titled "([^"]*)"$`, s.renderOverlayDrawer)
	sc.When(`^I render a static drawer with id "([^"]*)" titled "([^"]*)"$`, s.renderOverlayStaticDrawer)
}

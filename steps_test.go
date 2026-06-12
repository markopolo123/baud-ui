package baudui_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"

	"github.com/markopolo123/baud-ui/baud"
)

// scenarioState holds the parsed HTML document under test.
type scenarioState struct {
	doc *html.Node
}

func textPane(s string) templComponent {
	return templComponent{s}
}

// templComponent is a trivial templ.Component for filling slots in tests.
type templComponent struct{ text string }

func (c templComponent) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "<span>"+c.text+"</span>")
	return err
}

func (s *scenarioState) render(c interface {
	Render(context.Context, io.Writer) error
}) error {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		return err
	}
	doc, err := html.Parse(&buf)
	if err != nil {
		return err
	}
	s.doc = doc
	return nil
}

// ---- render steps -------------------------------------------------------

func (s *scenarioState) renderShellWithNav() error {
	return s.render(baud.Shell(baud.ShellProps{
		TopBar:    textPane("topbar"),
		Nav:       textPane("nav"),
		StatusBar: textPane("status"),
	}))
}

func (s *scenarioState) renderShellWithoutNav() error {
	return s.render(baud.Shell(baud.ShellProps{
		TopBar:    textPane("topbar"),
		StatusBar: textPane("status"),
	}))
}

func (s *scenarioState) renderDefaultPage() error {
	return s.render(baud.Page(baud.PageProps{Title: "test"}))
}

func (s *scenarioState) renderPanes(template string, n int) error {
	return s.render(baud.Panes(baud.PanesProps{Template: template}, panes(n)...))
}

func (s *scenarioState) renderRowPanes(template string, n int) error {
	return s.render(baud.PanesRows(baud.PanesProps{Template: template}, panes(n)...))
}

func (s *scenarioState) renderResizablePanes(template, id string, n int) error {
	return s.render(baud.Panes(baud.PanesProps{
		Template:  template,
		Resizable: true,
		ID:        id,
	}, panes(n)...))
}

func panes(n int) []templ.Component {
	out := make([]templ.Component, n)
	for i := range out {
		out[i] = textPane(fmt.Sprintf("pane %d", i))
	}
	return out
}

// ---- assertion steps ----------------------------------------------------

func (s *scenarioState) exactlyNMatch(n int, selector string) error {
	got := len(s.matching(selector))
	if got != n {
		return fmt.Errorf("expected %d element(s) matching %q, got %d", n, selector, got)
	}
	return nil
}

func (s *scenarioState) noneMatch(selector string) error {
	return s.exactlyNMatch(0, selector)
}

func (s *scenarioState) elementHasClasses(selector, classes string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	have := strings.Fields(attrVal(n, "class"))
	for _, want := range strings.Fields(classes) {
		found := false
		for _, c := range have {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("element %q class list %v is missing %q", selector, have, want)
		}
	}
	return nil
}

func (s *scenarioState) elementHasAttr(selector, name, value string) error {
	n, err := s.one(selector)
	if err != nil {
		return err
	}
	if got, ok := attr(n, name); !ok || got != value {
		return fmt.Errorf("element %q attribute %q = %q (present=%v), want %q", selector, name, got, ok, value)
	}
	return nil
}

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

// behaviorsScriptFirst asserts the text/hyperscript behaviors script appears
// in the head before the _hyperscript library script (a documented
// requirement for remotely-loaded behaviors).
func (s *scenarioState) behaviorsScriptFirst() error {
	behaviorsIdx, libIdx := -1, -1
	idx := 0
	walk(s.doc, func(n *html.Node) {
		if n.Type != html.ElementNode || n.Data != "script" {
			return
		}
		idx++
		if attrVal(n, "type") == "text/hyperscript" {
			behaviorsIdx = idx
		}
		if attrVal(n, "src") == baud.HyperscriptSrc {
			libIdx = idx
		}
	})
	if behaviorsIdx == -1 {
		return fmt.Errorf("no script[type=text/hyperscript] found")
	}
	if libIdx == -1 {
		return fmt.Errorf("no _hyperscript library script (src=%s) found", baud.HyperscriptSrc)
	}
	if behaviorsIdx >= libIdx {
		return fmt.Errorf("behaviors script (#%d) must come before the hyperscript library (#%d)", behaviorsIdx, libIdx)
	}
	return nil
}

// ---- tiny selector engine (tag, .class, [attr], [attr=val], ">") --------

type simpleSel struct {
	tag     string
	classes []string
	attrs   []attrSel
}

type attrSel struct {
	key, val string
	hasVal   bool
}

func parseSelector(s string) []simpleSel {
	var chain []simpleSel
	for _, part := range strings.Split(s, ">") {
		chain = append(chain, parseSimple(strings.TrimSpace(part)))
	}
	return chain
}

func parseSimple(s string) simpleSel {
	var sel simpleSel
	i := 0
	for i < len(s) && s[i] != '.' && s[i] != '[' {
		i++
	}
	sel.tag = s[:i]
	for i < len(s) {
		switch s[i] {
		case '.':
			j := i + 1
			for j < len(s) && s[j] != '.' && s[j] != '[' {
				j++
			}
			sel.classes = append(sel.classes, s[i+1:j])
			i = j
		case '[':
			j := strings.IndexByte(s[i:], ']') + i
			body := s[i+1 : j]
			if k := strings.IndexByte(body, '='); k >= 0 {
				sel.attrs = append(sel.attrs, attrSel{key: body[:k], val: strings.Trim(body[k+1:], `"'`), hasVal: true})
			} else {
				sel.attrs = append(sel.attrs, attrSel{key: body})
			}
			i = j + 1
		default:
			i++
		}
	}
	return sel
}

func nodeMatches(n *html.Node, sel simpleSel) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if sel.tag != "" && n.Data != sel.tag {
		return false
	}
	have := strings.Fields(attrVal(n, "class"))
	for _, want := range sel.classes {
		found := false
		for _, c := range have {
			if c == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, a := range sel.attrs {
		got, ok := attr(n, a.key)
		if !ok {
			return false
		}
		if a.hasVal && got != a.val {
			return false
		}
	}
	return true
}

func chainMatches(n *html.Node, chain []simpleSel) bool {
	if !nodeMatches(n, chain[len(chain)-1]) {
		return false
	}
	if len(chain) == 1 {
		return true
	}
	if n.Parent == nil {
		return false
	}
	return chainMatches(n.Parent, chain[:len(chain)-1])
}

func (s *scenarioState) matching(selector string) []*html.Node {
	chain := parseSelector(selector)
	var out []*html.Node
	walk(s.doc, func(n *html.Node) {
		if chainMatches(n, chain) {
			out = append(out, n)
		}
	})
	return out
}

func (s *scenarioState) one(selector string) (*html.Node, error) {
	nodes := s.matching(selector)
	if len(nodes) != 1 {
		return nil, fmt.Errorf("expected exactly 1 element matching %q, got %d", selector, len(nodes))
	}
	return nodes[0], nil
}

func walk(n *html.Node, fn func(*html.Node)) {
	fn(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, fn)
	}
}

func attr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func attrVal(n *html.Node, key string) string {
	v, _ := attr(n, key)
	return v
}

// textContent concatenates all text nodes under n.
func textContent(n *html.Node) string {
	var b strings.Builder
	walk(n, func(m *html.Node) {
		if m.Type == html.TextNode {
			b.WriteString(m.Data)
		}
	})
	return b.String()
}

// InitializeScenario registers the step definitions.
func InitializeScenario(sc *godog.ScenarioContext) {
	s := &scenarioState{}

	sc.When(`^I render a shell with a nav$`, s.renderShellWithNav)
	sc.When(`^I render a shell without a nav$`, s.renderShellWithoutNav)
	sc.When(`^I render the default page$`, s.renderDefaultPage)
	sc.When(`^I render panes "([^"]*)" with (\d+) panes$`, s.renderPanes)
	sc.When(`^I render row panes "([^"]*)" with (\d+) panes$`, s.renderRowPanes)
	sc.When(`^I render resizable panes "([^"]*)" with id "([^"]*)" and (\d+) panes$`, s.renderResizablePanes)
	registerFieldSteps(sc, s)

	sc.Then(`^exactly (\d+) elements? match(?:es)? "([^"]*)"$`, s.exactlyNMatch)
	sc.Then(`^no element matches "([^"]*)"$`, s.noneMatch)
	sc.Then(`^the element "([^"]*)" has classes "([^"]*)"$`, s.elementHasClasses)
	sc.Then(`^the element "([^"]*)" has attribute "([^"]*)" equal to "([^"]*)"$`, s.elementHasAttr)
	sc.Then(`^the element "([^"]*)" has text "([^"]*)"$`, s.elementHasText)
	sc.Then(`^the behaviors script comes before the hyperscript library script$`, s.behaviorsScriptFirst)

	registerBtnSteps(sc, s)
	registerBadgeSteps(sc, s)
	registerChoiceSteps(sc, s)
	registerDefListSteps(sc, s)
	registerTagInputSteps(sc, s)
	registerPanelSteps(sc, s)
	registerDatePickerSteps(sc, s)
	registerTabsSteps(sc, s)
	registerSelectSteps(sc, s)
	registerPaginationSteps(sc, s)
	registerDataTableSteps(sc, s)
}

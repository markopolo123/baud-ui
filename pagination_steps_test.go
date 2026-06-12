package baudui_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"golang.org/x/net/html"

	"github.com/markopolo123/baud-ui/baud"
)

// Step definitions for features/data_extras.feature (Pagination +
// DiffViewer). Registered with one line in InitializeScenario
// (steps_test.go) — the only file this wave shares with others.
// Generic assertions reuse the shared exactlyNMatch / elementHasAttr /
// elementHasText steps; everything here is pagination-/diff-prefixed.

// ---- pagination render steps ---------------------------------------------

func (s *scenarioState) renderHTMXPager(page, total, perPage int, base, target string) error {
	return s.render(baud.Pagination(baud.PaginationProps{
		Page:    page,
		PerPage: perPage,
		Total:   total,
		HxGet:   base,
		Target:  target,
	}))
}

func (s *scenarioState) renderHTMXPagerWithMore(page, total, perPage int, base, target, moreURL, moreTarget string) error {
	return s.render(baud.Pagination(baud.PaginationProps{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		HxGet:      base,
		Target:     target,
		MoreURL:    moreURL,
		MoreTarget: moreTarget,
	}))
}

func (s *scenarioState) renderHrefPager(page, total, perPage int, base string) error {
	return s.render(baud.Pagination(baud.PaginationProps{
		Page:    page,
		PerPage: perPage,
		Total:   total,
		HrefFor: func(p int) string { return base + "?page=" + strconv.Itoa(p) },
	}))
}

// ---- diff render steps ----------------------------------------------------

// renderDiffFromRows builds DiffLines from "kind|old|new|text" docstring
// rows (empty old/new = nil gutter) and renders the viewer.
func (s *scenarioState) renderDiffFromRows(file string, doc *godog.DocString) error {
	var lines []baud.DiffLine
	for _, row := range strings.Split(doc.Content, "\n") {
		parts := strings.SplitN(row, "|", 4)
		if len(parts) != 4 {
			return fmt.Errorf("diff row spec %q: want kind|old|new|text", row)
		}
		l := baud.DiffLine{Kind: baud.DiffKind(parts[0]), Text: parts[3]}
		var err error
		if l.OldN, err = diffSpecNum(parts[1]); err != nil {
			return err
		}
		if l.NewN, err = diffSpecNum(parts[2]); err != nil {
			return err
		}
		lines = append(lines, l)
	}
	return s.render(baud.DiffViewer(baud.DiffProps{File: file, Lines: lines}))
}

func (s *scenarioState) renderParsedUnifiedDiff(file string, doc *godog.DocString) error {
	lines, err := baud.ParseUnified(doc.Content)
	if err != nil {
		return fmt.Errorf("ParseUnified: %w", err)
	}
	return s.render(baud.DiffViewer(baud.DiffProps{File: file, Lines: lines}))
}

func diffSpecNum(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("diff row gutter %q: %w", s, err)
	}
	return &n, nil
}

// ---- diff assertion steps -------------------------------------------------

// diffRow returns the 1-based nth .diff-line in document order.
func (s *scenarioState) diffRow(row int) (*html.Node, error) {
	rows := s.matching("div.diff-line")
	if row < 1 || row > len(rows) {
		return nil, fmt.Errorf("diff row %d out of range (have %d rows)", row, len(rows))
	}
	return rows[row-1], nil
}

// diffSpans collects the descendant spans of n carrying the given class,
// in document order.
func diffSpans(n *html.Node, class string) []*html.Node {
	var out []*html.Node
	walk(n, func(m *html.Node) {
		if m.Type != html.ElementNode || m.Data != "span" {
			return
		}
		for _, c := range strings.Fields(attrVal(m, "class")) {
			if c == class {
				out = append(out, m)
				return
			}
		}
	})
	return out
}

func diffHasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attrVal(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// diffRowIs asserts a non-hunk row's kind class, both gutter numbers
// (empty = blank gutter), the sign glyph (exact, untrimmed) and the text.
func (s *scenarioState) diffRowIs(row int, kind, oldN, newN, sign, text string) error {
	n, err := s.diffRow(row)
	if err != nil {
		return err
	}
	for _, marker := range []string{"add", "del", "hunk"} {
		want := marker == kind
		if diffHasClass(n, marker) != want {
			return fmt.Errorf("diff row %d class list %q: kind %q expected", row, attrVal(n, "class"), kind)
		}
	}
	gutters := diffSpans(n, "diff-no")
	if len(gutters) != 2 {
		return fmt.Errorf("diff row %d has %d gutters, want 2 (old + new)", row, len(gutters))
	}
	if got := strings.TrimSpace(textContent(gutters[0])); got != oldN {
		return fmt.Errorf("diff row %d old gutter = %q, want %q", row, got, oldN)
	}
	if got := strings.TrimSpace(textContent(gutters[1])); got != newN {
		return fmt.Errorf("diff row %d new gutter = %q, want %q", row, got, newN)
	}
	signs := diffSpans(n, "diff-sign")
	if len(signs) != 1 {
		return fmt.Errorf("diff row %d has %d sign cells, want 1", row, len(signs))
	}
	if got := textContent(signs[0]); got != sign {
		return fmt.Errorf("diff row %d sign = %q, want %q", row, got, sign)
	}
	texts := diffSpans(n, "diff-text")
	if len(texts) != 1 {
		return fmt.Errorf("diff row %d has %d text cells, want 1", row, len(texts))
	}
	if got := textContent(texts[0]); got != text {
		return fmt.Errorf("diff row %d text = %q, want %q", row, got, text)
	}
	return nil
}

// diffRowIsHunk asserts a hunk header row: marker class, raw header text,
// and no gutter/sign cells (the design indents it with padding instead).
func (s *scenarioState) diffRowIsHunk(row int, text string) error {
	n, err := s.diffRow(row)
	if err != nil {
		return err
	}
	if !diffHasClass(n, "hunk") {
		return fmt.Errorf("diff row %d class list %q is missing %q", row, attrVal(n, "class"), "hunk")
	}
	if got := strings.TrimSpace(textContent(n)); got != text {
		return fmt.Errorf("diff row %d text = %q, want %q", row, got, text)
	}
	if g := diffSpans(n, "diff-no"); len(g) != 0 {
		return fmt.Errorf("diff hunk row %d has %d gutter cells, want 0", row, len(g))
	}
	return nil
}

// registerPaginationSteps wires the Pagination + DiffViewer scenario steps
// onto the shared state.
func registerPaginationSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render an htmx pager at page (\d+) of (\d+) rows, (\d+) per page, getting "([^"]*)" into "([^"]*)"$`, s.renderHTMXPager)
	sc.When(`^I render an htmx pager at page (\d+) of (\d+) rows, (\d+) per page, getting "([^"]*)" into "([^"]*)" with load more "([^"]*)" into "([^"]*)"$`, s.renderHTMXPagerWithMore)
	sc.When(`^I render an href pager at page (\d+) of (\d+) rows, (\d+) per page, linking "([^"]*)"$`, s.renderHrefPager)
	sc.When(`^I render a diff "([^"]*)" from rows:$`, s.renderDiffFromRows)
	sc.When(`^I render the parsed unified diff "([^"]*)":$`, s.renderParsedUnifiedDiff)

	sc.Then(`^diff row (\d+) is a hunk with text "([^"]*)"$`, s.diffRowIsHunk)
	sc.Then(`^diff row (\d+) is "([^"]*)" with old "([^"]*)" new "([^"]*)" sign "([^"]*)" and text "([^"]*)"$`, s.diffRowIs)
}

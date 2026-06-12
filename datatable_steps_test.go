package baudui_test

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"

	"github.com/markopolo123/baud-ui/baud"
)

// Step definitions for features/datatable.feature. Registered with one line
// in InitializeScenario (steps_test.go) — the only file this feature shares
// with other waves. Assertions reuse the shared exactlyNMatch /
// elementHasAttr / elementHasText steps; everything here is
// datatable-prefixed per the helper rule.

// datatableColumns is the feature fixture: a sortable text column, a plain
// column, a sortable numeric column and a non-sortable numeric column —
// every Column flag combination the matrix needs.
func datatableColumns() []baud.Column {
	return []baud.Column{
		{Key: "host", Label: "Host", Sortable: true},
		{Key: "region", Label: "Reg"},
		{Key: "cpu", Label: "CPU%", Numeric: true, Sortable: true},
		{Key: "mem", Label: "Mem%", Numeric: true},
	}
}

func datatableRows() []baud.Row {
	return []baud.Row{
		{Key: "h1", Cells: []string{"alpha", "use1", "12.5", "40.0"}},
		{Key: "h2", Cells: []string{"bravo", "euw1", "75.2", "60.1"}},
		{Key: "h3", Cells: []string{"charlie", "use1", "96.7", "88.9"}},
	}
}

func datatableProps() baud.DataTableProps {
	return baud.DataTableProps{
		ID:       "dt-fleet",
		Columns:  datatableColumns(),
		Rows:     datatableRows(),
		Endpoint: "/demo/datatable",
	}
}

// ---- render steps --------------------------------------------------------

func (s *scenarioState) renderDataTable() error {
	return s.render(baud.DataTable(datatableProps()))
}

func (s *scenarioState) renderZebraDataTable() error {
	p := datatableProps()
	p.Zebra = true
	return s.render(baud.DataTable(p))
}

func (s *scenarioState) renderLinesDataTable() error {
	p := datatableProps()
	p.Lines = true
	return s.render(baud.DataTable(p))
}

func (s *scenarioState) renderDataTableSelected(key string) error {
	p := datatableProps()
	p.Selected = key
	return s.render(baud.DataTable(p))
}

func (s *scenarioState) renderDataTableSorted(key, dir string) error {
	p := datatableProps()
	p.SortKey = key
	p.SortDir = dir
	return s.render(baud.DataTable(p))
}

// renderDataTableToneHook renders with the threshold hook: cpu > 90 ⇒ err,
// cpu > 70 ⇒ warn. The hook keys off the column, so the mem cells (88.9,
// 60.1) must stay untoned — the scenario counts prove that.
func (s *scenarioState) renderDataTableToneHook() error {
	p := datatableProps()
	p.CellTone = func(c baud.Column, v string) baud.Tone {
		if c.Key != "cpu" {
			return baud.ToneNone
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return baud.ToneNone
		}
		switch {
		case f > 90:
			return baud.ToneErr
		case f > 70:
			return baud.ToneWarn
		}
		return baud.ToneNone
	}
	return s.render(baud.DataTable(p))
}

func (s *scenarioState) renderDataTableBodyOnly() error {
	return s.datatableRenderFragment(baud.DataTableBody(datatableProps()))
}

func (s *scenarioState) renderDataTableHeadOOB(key, dir string) error {
	p := datatableProps()
	p.SortKey = key
	p.SortDir = dir
	return s.datatableRenderFragment(baud.DataTableHead(p, true))
}

// datatableRenderFragment parses a thead/tbody fragment wrapped in a host
// <table> — html.Parse would otherwise foster-parent table innards rendered
// outside a table and drop the element under test.
func (s *scenarioState) datatableRenderFragment(c templ.Component) error {
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		return err
	}
	doc, err := html.Parse(strings.NewReader("<table>" + buf.String() + "</table>"))
	if err != nil {
		return err
	}
	s.doc = doc
	return nil
}

// ---- assertion steps -----------------------------------------------------

// datatableMarkOnSelectedOnly asserts the ▌ glyph shows in exactly the
// row-mark cells of .sel rows — every other mark cell stays blank.
func (s *scenarioState) datatableMarkOnSelectedOnly() error {
	marked := 0
	for _, n := range s.matching("tbody > tr > td.row-mark") {
		if strings.TrimSpace(textContent(n)) == "▌" {
			marked++
		}
	}
	sel := len(s.matching("tbody > tr.sel"))
	if marked != sel {
		return fmt.Errorf("▌ marks on %d row(s), but %d row(s) are selected", marked, sel)
	}
	return nil
}

// registerDataTableSteps wires the DataTable scenario steps onto the shared
// state.
func registerDataTableSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a data table$`, s.renderDataTable)
	sc.When(`^I render a zebra data table$`, s.renderZebraDataTable)
	sc.When(`^I render a lines data table$`, s.renderLinesDataTable)
	sc.When(`^I render a data table with row "([^"]*)" selected$`, s.renderDataTableSelected)
	sc.When(`^I render a data table sorted by "([^"]*)" "([^"]*)"$`, s.renderDataTableSorted)
	sc.When(`^I render a data table with a cpu threshold tone hook$`, s.renderDataTableToneHook)
	sc.When(`^I render only the data table body$`, s.renderDataTableBodyOnly)
	sc.When(`^I render the data table head as an out-of-band fragment sorted by "([^"]*)" "([^"]*)"$`, s.renderDataTableHeadOOB)

	sc.Then(`^the ▌ row mark appears on selected rows only$`, s.datatableMarkOnSelectedOnly)
}

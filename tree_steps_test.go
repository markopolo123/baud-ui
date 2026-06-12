package baudui_test

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"golang.org/x/net/html"

	"github.com/markopolo123/baud-ui/baud"
)

// treeSpecItem is one parsed line of the feature's tree DSL.
type treeSpecItem struct {
	depth int
	node  baud.TreeNode
}

// treeParseSpec turns the feature DSL into TreeNodes: two leading spaces
// per depth level; suffixes mark a node `>` expanded branch, `+` collapsed
// branch, `~URL` lazy branch, `*` selected, ` [meta]` right-aligned meta.
func treeParseSpec(spec string) ([]baud.TreeNode, error) {
	var items []treeSpecItem
	for _, raw := range strings.Split(spec, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent%2 != 0 {
			return nil, fmt.Errorf("tree spec line %q: odd indent %d", line, indent)
		}
		items = append(items, treeSpecItem{depth: indent / 2, node: treeParseNode(strings.TrimSpace(line))})
	}
	pos := 0
	nodes, err := treeBuildLevel(items, 0, &pos)
	if err != nil {
		return nil, err
	}
	if pos != len(items) {
		return nil, fmt.Errorf("tree spec: line %d nests deeper than its parent allows", pos+1)
	}
	return nodes, nil
}

// treeParseNode decodes one trimmed DSL line into a node (children attach
// later from indentation).
func treeParseNode(s string) baud.TreeNode {
	var n baud.TreeNode
	if i := strings.IndexByte(s, '~'); i >= 0 {
		n.LazyURL = s[i+1:]
		s = s[:i]
	}
markers:
	for len(s) > 0 {
		switch s[len(s)-1] {
		case '>':
			n.Expanded = true
		case '+':
			// collapsed branch — the default; children attach by indent.
		case '*':
			n.Selected = true
		default:
			break markers
		}
		s = s[:len(s)-1]
	}
	if strings.HasSuffix(s, "]") {
		if i := strings.LastIndex(s, " ["); i >= 0 {
			n.Meta = s[i+2 : len(s)-1]
			s = s[:i]
		}
	}
	n.Label = strings.TrimSpace(s)
	return n
}

// treeBuildLevel consumes items at exactly depth, recursing for each
// node's deeper children, and returns at the first shallower line.
func treeBuildLevel(items []treeSpecItem, depth int, pos *int) ([]baud.TreeNode, error) {
	var out []baud.TreeNode
	for *pos < len(items) {
		it := items[*pos]
		if it.depth < depth {
			return out, nil
		}
		if it.depth > depth {
			if len(out) == 0 {
				return out, nil // parent caller reports the skip
			}
			return nil, fmt.Errorf("tree spec: %q skips a depth level", it.node.Label)
		}
		*pos++
		n := it.node
		children, err := treeBuildLevel(items, depth+1, pos)
		if err != nil {
			return nil, err
		}
		n.Children = children
		out = append(out, n)
	}
	return out, nil
}

// ---- render steps ---------------------------------------------------------

func (s *scenarioState) renderTree(spec *godog.DocString) error {
	nodes, err := treeParseSpec(spec.Content)
	if err != nil {
		return err
	}
	return s.render(baud.Tree(baud.TreeProps{Nodes: nodes}))
}

func (s *scenarioState) renderTreeChildren(prefix string, spec *godog.DocString) error {
	nodes, err := treeParseSpec(spec.Content)
	if err != nil {
		return err
	}
	return s.render(baud.TreeChildren(prefix, nodes))
}

// ---- assertion steps ------------------------------------------------------

// treeFirstByClass returns the first descendant of root carrying class.
func treeFirstByClass(root *html.Node, class string) *html.Node {
	var found *html.Node
	walk(root, func(n *html.Node) {
		if found != nil || n.Type != html.ElementNode {
			return
		}
		for _, c := range strings.Fields(attrVal(n, "class")) {
			if c == class {
				found = n
				return
			}
		}
	})
	return found
}

// treeRow finds the unique .tree-row (leaf li or branch summary) whose
// .tree-label text equals label.
func (s *scenarioState) treeRow(label string) (*html.Node, error) {
	var rows []*html.Node
	for _, row := range s.matching(".tree-row") {
		lbl := treeFirstByClass(row, "tree-label")
		if lbl != nil && strings.TrimSpace(textContent(lbl)) == label {
			rows = append(rows, row)
		}
	}
	if len(rows) != 1 {
		return nil, fmt.Errorf("expected exactly 1 tree row labeled %q, got %d", label, len(rows))
	}
	return rows[0], nil
}

// treeRowHasGlyph compares the row's branch glyph text, ignoring the
// trailing alignment pad leaves carry in place of the ▸/▾ slot (the
// disclosure glyph itself is CSS content, so it never appears here).
func (s *scenarioState) treeRowHasGlyph(label, glyph string) error {
	row, err := s.treeRow(label)
	if err != nil {
		return err
	}
	g := treeFirstByClass(row, "tree-glyph")
	if g == nil {
		return fmt.Errorf("tree row %q has no .tree-glyph", label)
	}
	if got := strings.TrimRight(textContent(g), " "); got != glyph {
		return fmt.Errorf("tree row %q glyph = %q, want %q", label, got, glyph)
	}
	return nil
}

func (s *scenarioState) treeRowHasNoMeta(label string) error {
	row, err := s.treeRow(label)
	if err != nil {
		return err
	}
	if m := treeFirstByClass(row, "tree-meta"); m != nil {
		return fmt.Errorf("tree row %q has meta %q, want none", label, strings.TrimSpace(textContent(m)))
	}
	return nil
}

func (s *scenarioState) treeRowIsSelected(label string) error {
	row, err := s.treeRow(label)
	if err != nil {
		return err
	}
	for _, c := range strings.Fields(attrVal(row, "class")) {
		if c == "sel" {
			return nil
		}
	}
	return fmt.Errorf("tree row %q class = %q, missing sel", label, attrVal(row, "class"))
}

// registerTreeSteps wires the Tree scenario steps onto the shared scenario
// state. Registered with one line in InitializeScenario (steps_test.go).
func registerTreeSteps(sc *godog.ScenarioContext, s *scenarioState) {
	sc.When(`^I render a tree from:$`, s.renderTree)
	sc.When(`^I render a tree children fragment with prefix "([^"]*)" from:$`, s.renderTreeChildren)
	sc.Then(`^the tree row labeled "([^"]*)" has glyph "([^"]*)"$`, s.treeRowHasGlyph)
	sc.Then(`^the tree row labeled "([^"]*)" has no meta$`, s.treeRowHasNoMeta)
	sc.Then(`^the tree row labeled "([^"]*)" is the selected row$`, s.treeRowIsSelected)
}

package baud

// TreeNode is one node of a Tree. A node with Children or a LazyURL is a
// branch rendered as a native <details>/<summary> disclosure; anything
// else is a plain leaf row. Depth is drawn with box-drawing glyphs in the
// row itself — never with indentation.
type TreeNode struct {
	// Label is the row text.
	Label string
	// Meta renders right-aligned in --fg-faint (pod counts, sizes, …)
	// and only when non-empty.
	Meta string
	// Children render eagerly into the branch's ul[role=group].
	Children []TreeNode
	// Selected marks the row: --sel fill, accent inset bar, accent
	// glyph, aria-current.
	Selected bool
	// Expanded opens the branch on first render. Ignored for lazy
	// branches, which must start collapsed so their first toggle stays
	// the fetch trigger.
	Expanded bool
	// LazyURL, when set, hx-gets the TreeChildren fragment into the
	// branch's empty group on first expand (hx-trigger="toggle once").
	LazyURL string
}

// TreeProps configures the Tree root (see components/tree.css).
type TreeProps struct {
	// ID is the optional ul id.
	ID string
	// Nodes are the root nodes — depth 0 carries no branch glyphs.
	Nodes []TreeNode
}

// branch reports whether the node renders as a disclosure.
func (n TreeNode) branch() bool { return len(n.Children) > 0 || n.LazyURL != "" }

// open reports whether the details starts open. Lazy branches never do:
// the first toggle must be the hx-get trigger.
func (n TreeNode) open() bool { return n.Expanded && n.LazyURL == "" }

func (n TreeNode) ariaExpanded() string {
	if n.open() {
		return "true"
	}
	return "false"
}

// treeRowGlyph draws one row's branch glyph under the accumulated
// ancestor prefix: ├─ for a middle sibling, └─ for the last. Root rows
// carry none.
func treeRowGlyph(prefix string, depth int, last bool) string {
	if depth == 0 {
		return ""
	}
	if last {
		return prefix + "└─"
	}
	return prefix + "├─"
}

// treeChildPrefix extends the ancestor prefix for one row's children:
// │ continues past a middle sibling, spaces past the last.
func treeChildPrefix(prefix string, depth int, last bool) string {
	if depth == 0 {
		return ""
	}
	if last {
		return prefix + "   "
	}
	return prefix + "│  "
}

package demo

import (
	"net/http"

	"github.com/markopolo123/baud-ui/baud"
)

// treeFragment is one lazy branch's server-side children: the rows plus
// the box-drawing prefix the branch's position dictates (│/spaces per
// still-continuing ancestor) — the server owns glyph continuity.
type treeFragment struct {
	Prefix string
	Nodes  []baud.TreeNode
}

// treeLazyNodes fixes the demo's lazy tree content by node path. The
// prod/edge branch is the last child of root-level prod, so its children
// sit under a three-space prefix.
var treeLazyNodes = map[string]treeFragment{
	"prod/edge": {
		Prefix: "   ",
		Nodes: []baud.TreeNode{
			{Label: "edge-cache-1", Meta: "warm"},
			{Label: "edge-cache-2", Meta: "warm"},
			{Label: "edge-lb", Meta: "v2"},
		},
	},
}

// treeChildren serves GET /demo/tree?node=… — the lazy-branch round-trip.
// A lazy branch hx-gets this TreeChildren fragment into its ul[role=group]
// on first toggle; unknown nodes are a 400.
func treeChildren(w http.ResponseWriter, r *http.Request) {
	frag, ok := treeLazyNodes[r.URL.Query().Get("node")]
	if !ok {
		http.Error(w, "unknown node", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	baud.TreeChildren(frag.Prefix, frag.Nodes).Render(r.Context(), w)
}

package baud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DiffKind classifies one DiffViewer row.
type DiffKind string

const (
	DiffCtx  DiffKind = "ctx"  // unchanged context line — plain row
	DiffAdd  DiffKind = "add"  // added line — 10% --ok tint, + sign
	DiffDel  DiffKind = "del"  // deleted line — 10% --err tint, − sign
	DiffHunk DiffKind = "hunk" // @@ … @@ header — --info tint, no gutters
)

// DiffLine is one row of a unified diff. Gutter numbers are pointers so a
// side that has no number (the old side of an add, the new side of a del,
// hunk headers) renders a blank gutter cell.
type DiffLine struct {
	Kind DiffKind
	OldN *int // old-file line number; nil = blank gutter
	NewN *int // new-file line number; nil = blank gutter
	Text string
}

// DiffProps configures the DiffViewer (design/README.md "Data"): a
// server-rendered unified diff — the component is pure CSS, no client
// behaviour. File labels the header, which also reports +adds/−dels
// counts derived from Lines.
type DiffProps struct {
	// File names the diff in the header bar; empty omits the header.
	File  string
	Lines []DiffLine
}

func (p DiffProps) count(k DiffKind) int {
	n := 0
	for _, l := range p.Lines {
		if l.Kind == k {
			n++
		}
	}
	return n
}

func (p DiffProps) addsText() string { return "+" + strconv.Itoa(p.count(DiffAdd)) }
func (p DiffProps) delsText() string { return "−" + strconv.Itoa(p.count(DiffDel)) }

// sign is the marker cell glyph: + for adds, − (minus sign) for dels, a
// plain space for context.
func (l DiffLine) sign() string {
	switch l.Kind {
	case DiffAdd:
		return "+"
	case DiffDel:
		return "−"
	}
	return " "
}

// diffGutter renders one line-number cell; nil = blank.
func diffGutter(n *int) string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(*n)
}

// diffHunkRE matches "@@ -old[,len] +new[,len] @@" (optional trailing
// section heading).
var diffHunkRE = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

// ParseUnified parses single-file unified-diff text into DiffLines,
// numbering both gutters from the hunk headers. File headers before the
// first hunk (---/+++/diff/index/…) are skipped. Malformed input — a bad
// @@ header, an unmarked line, content before any hunk, or a diff with no
// hunks at all — returns an error; ParseUnified never panics.
func ParseUnified(src string) ([]DiffLine, error) {
	var out []DiffLine
	oldN, newN := 0, 0
	inHunk := false
	for i, raw := range strings.Split(strings.TrimSuffix(src, "\n"), "\n") {
		switch {
		case strings.HasPrefix(raw, "@@"):
			m := diffHunkRE.FindStringSubmatch(raw)
			if m == nil {
				return nil, fmt.Errorf("line %d: malformed hunk header %q", i+1, raw)
			}
			// The regexp guarantees plain digits; range errors surface as 0.
			oldN, _ = strconv.Atoi(m[1])
			newN, _ = strconv.Atoi(m[2])
			inHunk = true
			out = append(out, DiffLine{Kind: DiffHunk, Text: raw})
		case !inHunk:
			if diffFileHeader(raw) {
				continue
			}
			return nil, fmt.Errorf("line %d: %q before any @@ hunk header", i+1, raw)
		case strings.HasPrefix(raw, "+"):
			n := newN
			newN++
			out = append(out, DiffLine{Kind: DiffAdd, NewN: &n, Text: raw[1:]})
		case strings.HasPrefix(raw, "-"):
			n := oldN
			oldN++
			out = append(out, DiffLine{Kind: DiffDel, OldN: &n, Text: raw[1:]})
		case strings.HasPrefix(raw, " ") || raw == "":
			o, n := oldN, newN
			oldN++
			newN++
			out = append(out, DiffLine{Kind: DiffCtx, OldN: &o, NewN: &n, Text: strings.TrimPrefix(raw, " ")})
		case strings.HasPrefix(raw, `\`):
			// "\ No newline at end of file" — metadata, no line numbers.
			out = append(out, DiffLine{Kind: DiffCtx, Text: raw})
		default:
			return nil, fmt.Errorf("line %d: unrecognised diff line %q", i+1, raw)
		}
	}
	if !inHunk {
		return nil, fmt.Errorf("no @@ hunk headers found")
	}
	return out, nil
}

// diffFileHeader reports whether a pre-hunk line is a recognised
// unified/git file header.
func diffFileHeader(s string) bool {
	for _, prefix := range []string{
		"--- ", "+++ ", "diff ", "index ", "new file", "deleted file",
		"old mode", "new mode", "similarity", "rename ", "copy ", "Binary files",
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return s == ""
}

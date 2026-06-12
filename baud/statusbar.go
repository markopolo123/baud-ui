package baud

// StatusCell is one cell of the StatusBar strip.
type StatusCell struct {
	Text string
	// Mode marks the vim-style mode cell — accent bg, on-accent bold text.
	// By convention it is the first cell.
	Mode bool
	// Spring makes the cell flex, pushing later cells to the far edge.
	// Exactly one cell should spring.
	Spring bool
}

// classes builds the cell marker list: sb-cell [sb-mode] [sb-spring].
func (c StatusCell) classes() string {
	cls := "sb-cell"
	if c.Mode {
		cls += " sb-mode"
	}
	if c.Spring {
		cls += " sb-spring"
	}
	return cls
}

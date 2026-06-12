package baud

// BtnVariant selects the button's visual variant.
type BtnVariant string

const (
	BtnDefault BtnVariant = ""        // raised fill, hairline border
	BtnPrimary BtnVariant = "primary" // accent fill, on-accent text
	BtnDanger  BtnVariant = "danger"  // transparent with err border; fills err on hover
	BtnGhost   BtnVariant = "ghost"   // borderless, muted until hover
)

// BtnProps configures the Btn primitive (design/README.md "Primitives").
// Label casing/letter-spacing comes from the type rules in CSS — author
// labels in lowercase.
type BtnProps struct {
	Label    string
	Variant  BtnVariant
	Glyph    string // mono char prefix, e.g. "▸"
	Kbd      string // inline shortcut hint chip, e.g. "⌘K"
	Active   bool   // pressed/selected look (accent fill)
	Disabled bool
}

// classes builds the marker class list: btn [btn-<variant>] [is-active].
func (p BtnProps) classes() string {
	cls := "btn"
	if p.Variant != BtnDefault {
		cls += " btn-" + string(p.Variant)
	}
	if p.Active {
		cls += " is-active"
	}
	return cls
}

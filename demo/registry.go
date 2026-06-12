package demo

import "github.com/a-h/templ"

// SheetSection is one section of the component sheet.
type SheetSection struct {
	Title string
	Body  templ.Component
}

// Sections is the component-sheet registry, rendered in order.
// Each component wave adds a demo/sheet_<name>.templ file and registers it
// here — this one line is the only expected merge-conflict point.
func Sections() []SheetSection {
	return []SheetSection{
		{Title: "Tokens", Body: SheetTokens()},
		{Title: "Layout — panes", Body: SheetLayout()},
		{Title: "Primitives — buttons & kbd", Body: SheetBtns()},
		{Title: "Primitives — badge & dot", Body: SheetBadge()},
		{Title: "Field / Input", Body: SheetField()},
		{Title: "Choice — checkbox · radio · toggle", Body: SheetChoice()},
		{Title: "DatePicker", Body: SheetDatePicker()},
		{Title: "Structure — deflist & breadcrumb", Body: SheetDefList()},
		{Title: "TagInput", Body: SheetTagInput()},
		{Title: "Structure — panel · statusbar · toolbar", Body: SheetPanel()},
		{Title: "Tabs — underline · boxed", Body: SheetTabs()},
		{Title: "Select — native · menu · combobox", Body: SheetSelect()},
		{Title: "Data — pagination & diff", Body: SheetDataExtras()},
		{Title: "Data — table (htmx sort)", Body: SheetDataTable()},
	}
}

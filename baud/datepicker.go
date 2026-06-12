package baud

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DatePickerProps configures a DatePicker. The trigger shows the Selected
// date as YYYY-MM-DD and a real hidden input carries the value for forms.
// The menu grid is always server-computed: the «‹›» nav buttons hx-get the
// Endpoint, whose handler re-renders the exported DatePickerMenu fragment
// for the target month — no client date math anywhere. Picking a day or a
// preset is purely-local hyperscript (copy the button's server-computed
// data-date into the input + trigger, close the menu).
type DatePickerProps struct {
	Name     string    // form field name for the hidden value input
	ID       string    // optional: lands on the root for external hooks
	Selected time.Time // zero = nothing picked yet
	Month    time.Time // view month; zero falls back to Selected, then Today
	Today    time.Time // zero = time.Now(); injectable for deterministic tests
	Endpoint string    // hx-get base URL of the menu-fragment handler
}

func (p DatePickerProps) today() time.Time {
	if p.Today.IsZero() {
		return time.Now()
	}
	return p.Today
}

// month resolves the viewed month, normalised to its first day.
func (p DatePickerProps) month() time.Time {
	m := p.Month
	if m.IsZero() {
		m = p.Selected
	}
	if m.IsZero() {
		m = p.today()
	}
	return time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, m.Location())
}

func (p DatePickerProps) triggerText() string {
	if p.Selected.IsZero() {
		return "pick date"
	}
	return p.Selected.Format("2006-01-02")
}

func (p DatePickerProps) value() string {
	if p.Selected.IsZero() {
		return ""
	}
	return p.Selected.Format("2006-01-02")
}

func (p DatePickerProps) title() string { return p.month().Format("Jan 2006") }

// navURL is the hx-get URL that re-renders the menu fragment shifted by the
// given number of months (±1 month, ±12 year). The selected date rides
// along so the re-rendered grid keeps its accent cell.
func (p DatePickerProps) navURL(months int) string {
	q := url.Values{}
	q.Set("month", p.month().AddDate(0, months, 0).Format("2006-01"))
	if !p.Selected.IsZero() {
		q.Set("selected", p.Selected.Format("2006-01-02"))
	}
	sep := "?"
	if strings.Contains(p.Endpoint, "?") {
		sep = "&"
	}
	return p.Endpoint + sep + q.Encode()
}

func (p DatePickerProps) dayIsOut(d time.Time) bool { return d.Month() != p.month().Month() }

// presetDate is the server-computed date a preset button carries: today
// minus the given number of days, formatted for data-date.
func (p DatePickerProps) presetDate(days int) string {
	return p.today().AddDate(0, 0, -days).Format("2006-01-02")
}

// MonthGrid returns the 42 days (Monday-first 6×7, always six full weeks)
// shown for the month containing m. Exported so server handlers can reuse
// the exact grid computation the templ fragment renders.
func MonthGrid(m time.Time) []time.Time {
	first := time.Date(m.Year(), m.Month(), 1, 0, 0, 0, 0, m.Location())
	offset := (int(first.Weekday()) + 6) % 7 // Monday-first
	start := first.AddDate(0, 0, -offset)
	days := make([]time.Time, 42)
	for i := range days {
		days[i] = start.AddDate(0, 0, i)
	}
	return days
}

// dpDows are the Monday-first day-of-week column headers.
var dpDows = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

// dpPresets is the preset row: label + days back from today.
var dpPresets = []struct {
	label string
	days  int
}{
	{"today", 0}, {"-1d", 1}, {"-7d", 7}, {"-30d", 30},
}

func dpSameDay(a, b time.Time) bool {
	return !a.IsZero() && !b.IsZero() &&
		a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func dpDayNum(d time.Time) string { return strconv.Itoa(d.Day()) }

func dpDayDate(d time.Time) string { return d.Format("2006-01-02") }

// dpRootScript keeps the trigger's aria-expanded in sync with the .open
// class, however it changes (trigger toggle, MenuDismiss outside-click/Esc,
// a pick closing the menu).
const dpRootScript = `install MenuDismiss
on mutation of @class
  set t to the first <.dp-trigger/> in me
  if I match .open
    set t's @aria-expanded to 'true'
  else
    set t's @aria-expanded to 'false'
  end
end`

// dpPickScript is the delegated day/preset click handler on the persistent
// menu container (it survives htmx swaps of the fragment inside). Every
// pickable button carries a server-computed data-date; picking only copies
// that string into the hidden input + trigger text, moves the accent cell
// and closes the menu — purely-local UI, no date math.
const dpPickScript = `on click(target)
  set b to target.closest('button[data-date]')
  if no b then exit end
  set v to b's @data-date
  set dp to closest .dp
  set inp to the first <input.dp-input/> in dp
  set inp's value to v
  put v into the first <.dp-value/> in dp
  repeat for c in <button.dp-day/> in dp
    if c's @data-date is v
      add .sel to c
    else
      remove .sel from c
    end
  end
  remove .open from dp
end`

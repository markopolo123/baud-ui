package baud

import (
	"testing"
	"time"
)

func datepickerDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestMonthGrid(t *testing.T) {
	cases := []struct {
		name        string
		month       time.Time
		first, last time.Time
		out         int // out-month cell count
	}{
		{
			// 2026-06-01 is a Monday: zero leading out-month cells.
			name:  "month starting Monday",
			month: datepickerDate(2026, time.June, 12),
			first: datepickerDate(2026, time.June, 1),
			last:  datepickerDate(2026, time.July, 12),
			out:   12,
		},
		{
			// 2026-02-01 is a Sunday and Feb 2026 is non-leap (28 days).
			name:  "non-leap February starting Sunday",
			month: datepickerDate(2026, time.February, 1),
			first: datepickerDate(2026, time.January, 26),
			last:  datepickerDate(2026, time.March, 8),
			out:   14,
		},
		{
			// 2024-02-01: leap February.
			name:  "leap February",
			month: datepickerDate(2024, time.February, 15),
			first: datepickerDate(2024, time.January, 29),
			last:  datepickerDate(2024, time.March, 10),
			out:   13,
		},
		{
			// Year boundary: December grid runs into January.
			name:  "December crossing the year",
			month: datepickerDate(2026, time.December, 31),
			first: datepickerDate(2026, time.November, 30),
			last:  datepickerDate(2027, time.January, 10),
			out:   11,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			grid := MonthGrid(c.month)
			if len(grid) != 42 {
				t.Fatalf("len(grid) = %d, want 42 (6×7, always)", len(grid))
			}
			if grid[0].Weekday() != time.Monday {
				t.Errorf("grid starts on %v, want Monday", grid[0].Weekday())
			}
			if !grid[0].Equal(c.first) {
				t.Errorf("grid[0] = %v, want %v", grid[0], c.first)
			}
			if !grid[41].Equal(c.last) {
				t.Errorf("grid[41] = %v, want %v", grid[41], c.last)
			}
			out := 0
			for i, d := range grid {
				if i > 0 && !d.Equal(grid[i-1].AddDate(0, 0, 1)) {
					t.Errorf("grid[%d] = %v does not follow %v", i, d, grid[i-1])
				}
				if d.Month() != c.month.Month() {
					out++
				}
			}
			if out != c.out {
				t.Errorf("out-month cells = %d, want %d", out, c.out)
			}
		})
	}
}

func TestDatePickerPropsMonthFallback(t *testing.T) {
	sel := datepickerDate(2026, time.June, 12)
	p := DatePickerProps{Selected: sel}
	if got := p.month(); !got.Equal(datepickerDate(2026, time.June, 1)) {
		t.Errorf("month() = %v, want first of the selected month", got)
	}
	p = DatePickerProps{Selected: sel, Month: datepickerDate(2025, time.December, 9)}
	if got := p.month(); !got.Equal(datepickerDate(2025, time.December, 1)) {
		t.Errorf("month() = %v, want first of the explicit view month", got)
	}
	p = DatePickerProps{Today: datepickerDate(2026, time.March, 3)}
	if got := p.month(); !got.Equal(datepickerDate(2026, time.March, 1)) {
		t.Errorf("month() = %v, want first of today's month", got)
	}
}

func TestDatePickerNavURL(t *testing.T) {
	p := DatePickerProps{
		Month:    datepickerDate(2026, time.January, 1),
		Selected: datepickerDate(2026, time.January, 15),
		Endpoint: "/api/datepicker",
	}
	if got, want := p.navURL(-1), "/api/datepicker?month=2025-12&selected=2026-01-15"; got != want {
		t.Errorf("navURL(-1) = %q, want %q (year rollover)", got, want)
	}
	if got, want := p.navURL(12), "/api/datepicker?month=2027-01&selected=2026-01-15"; got != want {
		t.Errorf("navURL(12) = %q, want %q", got, want)
	}
	// An endpoint with its own query string gets '&', not a second '?'.
	p = DatePickerProps{Month: datepickerDate(2026, time.June, 1), Endpoint: "/x?picker=a"}
	if got, want := p.navURL(1), "/x?picker=a&month=2026-07"; got != want {
		t.Errorf("navURL(1) = %q, want %q", got, want)
	}
}

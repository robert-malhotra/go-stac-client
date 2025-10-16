package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type datePicker struct {
	*tview.Flex
	monthLabel *tview.TextView
	calendar   *tview.Table
	info       *tview.TextView
	buttons    *tview.Form

	currentDate  time.Time
	currentMonth time.Time

	start    time.Time
	end      time.Time
	hasStart bool
	hasEnd   bool

	done func(confirmed bool, start, end time.Time)
}

func newDatePicker() *datePicker {
	p := &datePicker{}
	p.monthLabel = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	p.calendar = tview.NewTable().
		SetSelectable(true, true).
		SetFixed(1, 0)

	p.info = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)

	p.buttons = tview.NewForm()
	p.buttons.AddButton("Confirm", func() {
		p.confirmSelection()
	})
	p.buttons.AddButton("Clear", func() {
		p.clearSelection()
	})
	p.buttons.AddButton("Cancel", func() {
		p.finish(false)
	})
	p.buttons.SetButtonsAlign(tview.AlignCenter)

	p.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p.monthLabel, 1, 0, false).
		AddItem(p.calendar, 0, 1, true).
		AddItem(p.info, 3, 0, false).
		AddItem(p.buttons, 3, 0, false)
	p.SetBorder(true).SetTitle("Select Date Range")

	p.currentDate = normalizeDate(time.Now().UTC())
	p.currentMonth = monthStart(p.currentDate)

	p.calendar.SetInputCapture(p.onCalendarInput)
	p.calendar.SetSelectionChangedFunc(func(row, column int) {
		cell := p.calendar.GetCell(row, column)
		if cell == nil {
			return
		}
		if ref, ok := cell.GetReference().(time.Time); ok {
			p.currentDate = ref
		}
	})
	p.calendar.SetSelectedFunc(func(row, column int) {
		cell := p.calendar.GetCell(row, column)
		if cell == nil {
			return
		}
		if ref, ok := cell.GetReference().(time.Time); ok {
			p.selectDate(ref)
		}
	})

	p.updateInfo()
	p.renderCalendar()
	return p
}

func (p *datePicker) FocusTarget() tview.Primitive {
	return p.calendar
}

func (p *datePicker) SetDoneFunc(fn func(bool, time.Time, time.Time)) {
	p.done = fn
}

func (p *datePicker) SetRange(start, end *time.Time) {
	if start != nil {
		p.start = normalizeDate(start.UTC())
		p.hasStart = true
		p.currentDate = p.start
	} else {
		p.start = time.Time{}
		p.hasStart = false
	}

	if end != nil {
		p.end = normalizeDate(end.UTC())
		p.hasEnd = true
		if !p.hasStart {
			p.currentDate = p.end
		}
	} else {
		p.end = time.Time{}
		p.hasEnd = false
	}

	if p.hasStart && p.hasEnd && p.end.Before(p.start) {
		p.start, p.end = p.end, p.start
	}

	if !p.hasStart && !p.hasEnd {
		p.currentDate = normalizeDate(time.Now().UTC())
	}

	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
	p.updateInfo()
}

func (p *datePicker) onCalendarInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Modifiers()&tcell.ModCtrl != 0 {
		switch event.Key() {
		case tcell.KeyLeft:
			p.moveMonths(-1)
			return nil
		case tcell.KeyRight:
			p.moveMonths(1)
			return nil
		case tcell.KeyUp:
			p.moveYears(-1)
			return nil
		case tcell.KeyDown:
			p.moveYears(1)
			return nil
		case tcell.KeyEnter:
			if p.hasStart {
				p.confirmSelection()
				return nil
			}
		}
	}

	switch event.Key() {
	case tcell.KeyLeft:
		p.moveDays(-1)
		return nil
	case tcell.KeyRight:
		p.moveDays(1)
		return nil
	case tcell.KeyUp:
		p.moveDays(-7)
		return nil
	case tcell.KeyDown:
		p.moveDays(7)
		return nil
	case tcell.KeyPgUp:
		p.moveMonths(-1)
		return nil
	case tcell.KeyPgDn:
		p.moveMonths(1)
		return nil
	case tcell.KeyHome:
		p.goToMonthBoundary(true)
		return nil
	case tcell.KeyEnd:
		p.goToMonthBoundary(false)
		return nil
	case tcell.KeyEnter:
		if p.hasStart && p.hasEnd && (p.currentDate.Equal(p.start) || p.currentDate.Equal(p.end)) {
			p.confirmSelection()
			return nil
		}
	case tcell.KeyEscape:
		p.finish(false)
		return nil
	}
	return event
}

func (p *datePicker) moveDays(delta int) {
	p.currentDate = normalizeDate(p.currentDate.AddDate(0, 0, delta))
	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
}

func (p *datePicker) moveMonths(delta int) {
	p.currentDate = normalizeDate(p.currentDate.AddDate(0, delta, 0))
	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
}

func (p *datePicker) moveYears(delta int) {
	p.currentDate = normalizeDate(p.currentDate.AddDate(delta, 0, 0))
	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
}

func (p *datePicker) goToMonthBoundary(start bool) {
	if start {
		p.currentDate = monthStart(p.currentMonth)
	} else {
		p.currentDate = monthEnd(p.currentMonth)
	}
	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
}

func (p *datePicker) selectDate(date time.Time) {
	date = normalizeDate(date)
	if !p.hasStart || (p.hasStart && p.hasEnd) {
		p.start = date
		p.hasStart = true
		p.end = time.Time{}
		p.hasEnd = false
	} else {
		if date.Before(p.start) {
			p.end = p.start
			p.start = date
		} else {
			p.end = date
		}
		p.hasEnd = true
	}
	p.currentDate = date
	p.currentMonth = monthStart(p.currentDate)
	p.renderCalendar()
	p.updateInfo()
}

func (p *datePicker) confirmSelection() {
	if p.done == nil {
		return
	}
	if !p.hasStart {
		p.done(true, time.Time{}, time.Time{})
		return
	}
	start := p.start
	end := p.end
	if !p.hasEnd {
		end = start
	}
	if end.Before(start) {
		start, end = end, start
	}
	p.done(true, start, end)
}

func (p *datePicker) clearSelection() {
	p.hasStart = false
	p.hasEnd = false
	p.start = time.Time{}
	p.end = time.Time{}
	p.currentDate = normalizeDate(time.Now().UTC())
	p.currentMonth = monthStart(p.currentDate)
	p.updateInfo()
	p.renderCalendar()
}

func (p *datePicker) finish(confirmed bool) {
	if p.done != nil {
		var start, end time.Time
		if confirmed && p.hasStart {
			start = p.start
			if p.hasEnd {
				end = p.end
			} else {
				end = p.start
			}
		}
		p.done(confirmed, start, end)
	}
}

func (p *datePicker) updateInfo() {
	var b strings.Builder
	b.WriteString("[yellow]Arrows move days; PgUp/PgDn or Ctrl+←/→ jump months; Ctrl+↑/↓ jump years. Esc cancels.[-]\n")

	switch {
	case !p.hasStart:
		b.WriteString("[yellow]Press Enter to set the start date.[-]")
	case !p.hasEnd:
		b.WriteString("[yellow]Start selected. Enter picks the end date or Confirm uses a single day.[-]")
	default:
		b.WriteString("[yellow]Range selected. Confirm to accept or Clear to start over.[-]")
	}

	p.info.SetText(b.String())
}

func (p *datePicker) renderCalendar() {
	p.calendar.Clear()
	weekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for col, label := range weekdays {
		cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", label)).
			SetAlign(tview.AlignCenter).
			SetSelectable(false)
		p.calendar.SetCell(0, col, cell)
	}

	monthStartDate := monthStart(p.currentMonth)
	p.monthLabel.SetText(fmt.Sprintf("%s  [gray](Ctrl+←/→ month, Ctrl+↑/↓ year)[-]", monthStartDate.Format("January 2006")))

	firstDayColumn := int(monthStartDate.Weekday())
	daysInMonth := daysIn(monthStartDate)
	row, col := 1, firstDayColumn
	var selectedRow, selectedCol int

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(monthStartDate.Year(), monthStartDate.Month(), day, 0, 0, 0, 0, time.UTC)
		text := fmt.Sprintf("%2d", day)
		cell := tview.NewTableCell(text).
			SetAlign(tview.AlignCenter).
			SetSelectable(true).
			SetReference(date)
		cell.SetTextColor(tcell.ColorWhite)

		inRange := p.hasStart && ((p.hasEnd && !date.Before(p.start) && !date.After(p.end)) || (!p.hasEnd && date.Equal(p.start)))
		isStart := p.hasStart && date.Equal(p.start)
		isEnd := p.hasEnd && date.Equal(p.end)

		if inRange {
			cell.SetTextColor(tcell.ColorBlack)
			cell.SetBackgroundColor(tcell.ColorDarkGreen)
		}
		if isStart || (isEnd && (!p.hasEnd || p.start.Equal(p.end))) {
			cell.SetTextColor(tcell.ColorBlack)
			cell.SetBackgroundColor(tcell.ColorYellow)
		} else if isEnd {
			cell.SetTextColor(tcell.ColorBlack)
			cell.SetBackgroundColor(tcell.ColorDarkCyan)
		}

		cell.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorLightCyan).Foreground(tcell.ColorBlack))

		tableRow, tableCol := row, col
		p.calendar.SetCell(tableRow, tableCol, cell)

		if date.Equal(p.currentDate) {
			selectedRow, selectedCol = tableRow, tableCol
		}

		col++
		if col > 6 {
			col = 0
			row++
		}
	}

	if selectedRow > 0 || selectedCol > 0 {
		p.calendar.Select(selectedRow, selectedCol)
	}
}

func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func monthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), daysIn(t), 0, 0, 0, 0, time.UTC)
}

func daysIn(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

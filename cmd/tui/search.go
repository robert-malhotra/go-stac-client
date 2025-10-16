package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/pkg/client"
)

const (
	searchFormPageID    = "searchForm"
	datePickerPageID    = "datePicker"
	datetimeMetadataKey = "datetime"
	collectionsMetaKey  = "collections"
)

func (t *TUI) setupSearchFormPage() {
	t.searchCollectionsField = tview.NewInputField().
		SetLabel("Collections").
		SetFieldWidth(40)

	t.searchDatetimeField = tview.NewInputField()
	t.searchDatetimeField.SetLabel("Datetime (ISO interval)")
	t.searchDatetimeField.SetFieldWidth(40)
	t.searchDatetimeField.SetDisabled(true)

	t.searchForm = tview.NewForm()
	t.searchForm.AddFormItem(t.searchCollectionsField)
	t.searchForm.AddFormItem(t.searchDatetimeField)
	t.searchForm.AddButton("Pick Date Range", func() {
		t.openDateRangePicker()
	})
	t.searchForm.AddButton("Search", func() {
		t.runBasicSearch()
	})
	t.searchForm.AddButton("Cancel", func() {
		t.pages.SwitchToPage("collections")
		t.app.SetFocus(t.collectionsList)
	})
	t.searchForm.SetButtonsAlign(tview.AlignCenter)
	t.searchForm.SetBorder(true).SetTitle("Basic Search")

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.searchForm, 0, 1, true)

	t.pages.AddPage(searchFormPageID, layout, true, false)

	if t.datePicker == nil {
		t.datePicker = newDatePicker()
		t.datePicker.SetDoneFunc(func(confirmed bool, start, end time.Time) {
			t.pages.HidePage(datePickerPageID)
			t.pages.RemovePage(datePickerPageID)

			if confirmed {
				if start.IsZero() {
					t.hasStagedStart = false
					t.hasStagedEnd = false
				} else {
					t.hasStagedStart = true
					t.stagedStart = start
					if end.IsZero() {
						t.hasStagedEnd = true
						t.stagedEnd = start
					} else {
						t.hasStagedEnd = true
						t.stagedEnd = end
					}
				}
				t.syncDatetimeField()
			}

			t.app.SetFocus(t.searchForm)
		})
	}
}

func (t *TUI) openDateRangePicker() {
	if t.datePicker == nil {
		t.datePicker = newDatePicker()
	}

	var startPtr, endPtr *time.Time
	if t.hasStagedStart {
		start := t.stagedStart
		startPtr = &start
	}
	if t.hasStagedEnd {
		end := t.stagedEnd
		endPtr = &end
	}
	t.datePicker.SetRange(startPtr, endPtr)

	t.pages.RemovePage(datePickerPageID)
	t.pages.AddPage(datePickerPageID, t.datePicker, true, true)
	t.pages.ShowPage(datePickerPageID)
	if focus := t.datePicker.FocusTarget(); focus != nil {
		t.app.SetFocus(focus)
	}
}

func (t *TUI) openBasicSearchForm() {
	if t.searchForm == nil {
		return
	}

	if t.lastSearchMetadata != nil {
		if collections, ok := t.lastSearchMetadata[collectionsMetaKey]; ok {
			t.searchCollectionsField.SetText(collections)
		}
		if raw, ok := t.lastSearchMetadata[datetimeMetadataKey]; ok {
			start, end := parseISOInterval(raw)
			if !start.IsZero() {
				t.hasStagedStart = true
				t.stagedStart = start
			} else {
				t.hasStagedStart = false
			}
			if !end.IsZero() {
				t.hasStagedEnd = true
				t.stagedEnd = end
			} else if t.hasStagedStart {
				t.hasStagedEnd = true
				t.stagedEnd = t.stagedStart
			} else {
				t.hasStagedEnd = false
			}
		}
	}

	t.syncDatetimeField()

	t.pages.SwitchToPage(searchFormPageID)
	t.pages.ShowPage(searchFormPageID)
	t.app.SetFocus(t.searchForm)
}

func (t *TUI) syncDatetimeField() {
	if t.searchDatetimeField == nil {
		return
	}
	if !t.hasStagedStart {
		t.searchDatetimeField.SetText("")
		return
	}
	start := t.stagedStart.UTC()
	end := start
	if t.hasStagedEnd {
		end = t.stagedEnd.UTC()
		if end.Before(start) {
			start, end = end, start
		}
	}
	iso := fmt.Sprintf("%s/%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	t.searchDatetimeField.SetText(iso)
}

func (t *TUI) runBasicSearch() {
	if t.client == nil {
		t.showError("No STAC API client loaded yet")
		return
	}

	collectionsText := strings.TrimSpace(t.searchCollectionsField.GetText())
	var collections []string
	if collectionsText != "" {
		parts := strings.Split(collectionsText, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				collections = append(collections, trimmed)
			}
		}
	}

	datetime := strings.TrimSpace(t.searchDatetimeField.GetText())

	params := client.SearchParams{
		Collections: collections,
		Datetime:    datetime,
	}

	label := "Basic search"
	if len(collections) > 0 {
		label = fmt.Sprintf("Search: %s", strings.Join(collections, ", "))
	}

	metadata := map[string]string{}
	if len(collections) > 0 {
		metadata[collectionsMetaKey] = strings.Join(collections, ",")
	}
	if datetime != "" {
		metadata[datetimeMetadataKey] = datetime
	}

	t.app.QueueUpdateDraw(func() {
		t.itemsList.Clear()
		t.itemSummary.Clear()
		t.itemsList.AddItem("Loading items…", "", 0, nil)
		t.itemsList.SetTitle(t.itemsListTitle(true))
		t.updateItemsHelp()
		t.pages.HidePage(searchFormPageID)
		t.pages.SwitchToPage("items")
		t.app.SetFocus(t.itemsList)
	})

	ctx, cancel := context.WithTimeout(t.baseCtx, 300*time.Second)
	seq := t.client.SearchSimple(ctx, params)
	t.startItemStream(label, metadata, seq, cancel)
}

func parseISOInterval(raw string) (time.Time, time.Time) {
	parts := strings.Split(raw, "/")
	switch len(parts) {
	case 2:
		start := parseISODate(strings.TrimSpace(parts[0]))
		end := parseISODate(strings.TrimSpace(parts[1]))
		return start, end
	case 1:
		start := parseISODate(strings.TrimSpace(parts[0]))
		return start, start
	default:
		return time.Time{}, time.Time{}
	}
}

func parseISODate(raw string) time.Time {
	if raw == "" || raw == ".." {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t
	}
	return time.Time{}
}

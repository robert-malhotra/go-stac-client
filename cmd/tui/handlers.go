package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func (t *TUI) onInputDone(key tcell.Key) {
	if key == tcell.KeyEnter {
		url := t.input.GetText()
		go t.fetchCollections(url)
	}
}

func (t *TUI) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlC {
		t.Stop()
		return nil
	}

	currentPage, _ := t.pages.GetFrontPage()

	// Handle 'j' key for JSON view
	if event.Key() == tcell.KeyRune {
		switch r := event.Rune(); {
		case r == 'j' || r == 'J':
			switch currentPage {
			case "collections":
				index := t.collectionsList.GetCurrentItem()
				if index >= 0 && index < len(t.cols) {
					col := t.cols[index]
					t.showJSON(fmt.Sprintf("Collection %s", col.Id), col)
				}
				return nil
			case "items":
				index := t.itemsList.GetCurrentItem()
				if index >= 0 && index < len(t.items) {
					item := t.items[index]
					t.showJSON(fmt.Sprintf("Item %s", item.Id), item)
				}
				return nil
			case "itemDetail":
				if t.currentItem != nil {
					t.showJSON(fmt.Sprintf("Item %s", t.currentItem.Id), t.currentItem)
				}
				return nil
			}
		case r == 's' || r == 'S':
			switch currentPage {
			case "collections", "items":
				t.openBasicSearchForm()
				return nil
			}
		}
	}

	// Item detail pane navigation
	if currentPage == "itemDetail" {
		if event.Key() == tcell.KeyTab {
			t.itemDetailFocus = (t.itemDetailFocus + 1) % len(t.itemDetailPanes)
			t.app.SetFocus(t.itemDetailPanes[t.itemDetailFocus])
			return nil
		} else if event.Key() == tcell.KeyBacktab {
			t.itemDetailFocus = (t.itemDetailFocus - 1 + len(t.itemDetailPanes)) % len(t.itemDetailPanes)
			t.app.SetFocus(t.itemDetailPanes[t.itemDetailFocus])
			return nil
		}
	}

	// Collections page focus toggle
	if currentPage == "collections" {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			if t.app.GetFocus() == t.collectionsList {
				t.app.SetFocus(t.colDetail)
			} else {
				t.app.SetFocus(t.collectionsList)
			}
			return nil
		}
	}

	if currentPage == searchPageID {
		switch event.Key() {
		case tcell.KeyTab:
			if t.searchCollectionsList != nil && t.searchCollectionsList.HasFocus() {
				t.focusSearchFormFirstField()
				return nil
			}
			if t.searchForm != nil && t.searchForm.HasFocus() {
				if t.searchCollectionsList != nil {
					t.app.SetFocus(t.searchCollectionsList)
				}
				return nil
			}
		case tcell.KeyBacktab:
			if t.searchForm != nil && t.searchForm.HasFocus() {
				if t.searchCollectionsList != nil {
					t.app.SetFocus(t.searchCollectionsList)
				}
				return nil
			}
			if t.searchCollectionsList != nil && t.searchCollectionsList.HasFocus() {
				t.focusSearchFormLastElement()
				return nil
			}
		}
	}

	// Escape key navigation
	if event.Key() == tcell.KeyEscape {
		// If JSON view is active, let its own handler deal with Escape.
		if currentPage == jsonPageID {
			return event
		}

		switch currentPage {
		case "download":
			t.cancelActiveDownload()
			t.restoreFocusAfterModal()
			return nil
		case "error", "info":
			t.pages.HidePage(currentPage)
			t.restoreFocusAfterModal()
			return nil
		case "itemDetail":
			t.pages.SwitchToPage("items")
			t.app.SetFocus(t.itemsList)
			return nil
		case "items":
			t.pages.SwitchToPage("collections")
			t.app.SetFocus(t.collectionsList)
			return nil
		case "collections":
			t.pages.SwitchToPage("input")
			t.app.SetFocus(t.input)
			return nil
		case searchPageID:
			t.closeSearchForm()
			return nil
		}
	}

	return event
}

func (t *TUI) restoreFocusAfterModal() {
	if t.app == nil || t.pages == nil {
		return
	}

	currentPage, primitive := t.pages.GetFrontPage()
	if primitive == nil {
		return
	}

	switch currentPage {
	case "items":
		if t.itemsList != nil {
			t.app.SetFocus(t.itemsList)
		}
	case "itemDetail":
		if len(t.itemDetailPanes) > 0 {
			if t.itemDetailFocus < 0 || t.itemDetailFocus >= len(t.itemDetailPanes) {
				t.itemDetailFocus = 0
			}
			t.app.SetFocus(t.itemDetailPanes[t.itemDetailFocus])
		}
	case "collections":
		if t.collectionsList != nil {
			t.app.SetFocus(t.collectionsList)
		}
	case searchPageID:
		if t.searchCollectionsList != nil {
			t.app.SetFocus(t.searchCollectionsList)
		}
	case "input":
		if t.input != nil {
			t.app.SetFocus(t.input)
		}
	}
}

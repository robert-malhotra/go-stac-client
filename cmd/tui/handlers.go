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
		r := event.Rune()
		if (r == 's' || r == 'S') && currentPage != jsonPageID {
			switch currentPage {
			case "collections", "items":
				t.openBasicSearchForm()
				return nil
			}
		}
		if r == 'j' || r == 'J' {
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

	// Escape key navigation
	if event.Key() == tcell.KeyEscape {
		// If JSON view is active, let its own handler deal with Escape.
		if currentPage == jsonPageID {
			return event
		}

		switch currentPage {
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
		}
	}

	return event
}

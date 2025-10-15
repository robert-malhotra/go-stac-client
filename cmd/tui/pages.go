package main

import (
	"context"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/cmd/tui/formatting"
	"github.com/robert-malhotra/go-stac-client/pkg/client"
)

func (t *TUI) setupPages() {
	t.setupInputPage()
	t.setupCollectionsPage()
	t.setupItemsPage()
	t.setupItemDetailPage()
	t.setupSearchFormPage()
}

const itemsHelpControls = "[yellow]↑/↓[white] select  [yellow]Enter[white] view detail  [yellow]j[white] raw JSON  [yellow]s[white] search  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit"

func (t *TUI) setupInputPage() {
	t.input = tview.NewInputField().
		SetLabel("STAC API URL: ").
		SetFieldWidth(60).
		SetText("https://earth-search.aws.element84.com/v1")
	t.input.SetBorder(true).SetTitle("Enter STAC API Root URL")
	t.input.SetDoneFunc(t.onInputDone)

	inputHelp := formatting.MakeHelpText("[yellow]Enter[white] load collections  [yellow]Ctrl+C[white] quit")
	inputPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.input, 0, 1, true).
		AddItem(inputHelp, 3, 0, false)

	t.pages.AddPage("input", inputPage, true, true)
}

func (t *TUI) setupCollectionsPage() {
	t.collectionsList = tview.NewList()
	t.collectionsList.SetBorder(true).SetTitle("Collections")
	t.collectionsList.ShowSecondaryText(false)

	t.colDetail = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true).SetScrollable(true)
	t.colDetail.SetBorder(true).SetTitle("Collection Details")

	collectionsContent := tview.NewFlex().
		AddItem(t.collectionsList, 0, 1, true).
		AddItem(t.colDetail, 0, 2, false)

	collectionsHelp := formatting.MakeHelpText("[yellow]↑/↓[white] select  [yellow]Enter[white] load items  [yellow]j[white] raw JSON  [yellow]s[white] search  [yellow]Tab[white] toggle focus  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit")
	collectionsPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(collectionsContent, 0, 1, true).
		AddItem(collectionsHelp, 3, 0, false)

	t.collectionsList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < len(t.cols) {
			col := t.cols[index]
			t.colDetail.SetText(formatting.FormatCollectionDetails(col))
			t.colDetail.ScrollToBeginning()
		} else {
			t.colDetail.Clear()
		}
	})

	t.pages.AddPage("collections", collectionsPage, true, false)
}

func (t *TUI) setupItemsPage() {
	t.itemsList = tview.NewList()
	t.itemsList.SetBorder(true)
	t.itemsList.SetTitle(t.itemsListTitle(false))
	t.itemsList.ShowSecondaryText(false)
	t.itemsList.SetWrapAround(false)

	t.itemSummary = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	t.itemSummary.SetBorder(true).SetTitle("Item Summary")

	itemsContent := tview.NewFlex().
		AddItem(t.itemsList, 0, 1, true).
		AddItem(t.itemSummary, 0, 1, false)

	t.itemsHelp = formatting.MakeHelpText("")
	t.updateItemsHelp()
	itemsPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(itemsContent, 0, 1, true).
		AddItem(t.itemsHelp, 3, 0, false)

	t.itemsList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		// Update summary
		if index < len(t.items) {
			item := t.items[index]
			t.itemSummary.SetText(formatting.FormatItemSummary(item))
			t.itemSummary.ScrollToBeginning()
		} else {
			t.itemSummary.Clear()
		}

		// Pagination
		if index >= t.itemsList.GetItemCount()-2 {
			lastItem, _ := t.itemsList.GetItemText(t.itemsList.GetItemCount() - 1)
			if lastItem == "Load more" {
				go t.loadNextPage()
			}
		}
	})

	t.pages.AddPage("items", itemsPage, true, false)
}

func (t *TUI) itemsListTitle(loading bool) string {
	title := "Items"
	if label := t.activeResultLabel; label != "" {
		title = fmt.Sprintf("%s – %s", title, label)
	}
	if loading {
		title += " (loading...)"
	}
	return title
}

func (t *TUI) itemsHelpText() string {
	if label := t.activeResultLabel; label != "" {
		return fmt.Sprintf("%s\n[white]Source: [green]%s[white]", itemsHelpControls, label)
	}
	return itemsHelpControls
}

func (t *TUI) updateItemsHelp() {
	if t.itemsHelp != nil {
		t.itemsHelp.SetText(t.itemsHelpText())
	}
}

func (t *TUI) setupItemDetailPage() {
	t.itemDetail = tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0)
	t.itemDetail.SetBorder(true).SetTitle("Item Detail")

	t.itemProperties = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	t.itemProperties.SetBorder(true).SetTitle("Properties")

	t.itemAssets = tview.NewList()
	t.itemAssets.SetBorder(true).SetTitle("Assets")
	t.itemAssets.ShowSecondaryText(false)

	t.itemLinks = tview.NewList()
	t.itemLinks.SetBorder(true).SetTitle("Links")
	t.itemLinks.ShowSecondaryText(false)

	t.itemDetailPanes = []tview.Primitive{t.itemProperties, t.itemAssets, t.itemLinks}

	itemDetailHelp := formatting.MakeHelpText("[yellow]Tab[white] next pane  [yellow]Shift+Tab[white] previous pane  [yellow]Enter[white] download asset  [yellow]j[white] raw JSON  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit")
	itemDetailPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.itemDetail, 0, 1, true).
		AddItem(itemDetailHelp, 3, 0, false)

	t.pages.AddPage("itemDetail", itemDetailPage, true, false)
}

func (t *TUI) setupSearchFormPage() {
	t.searchCollectionsField = tview.NewInputField().
		SetLabel("Collections (comma-separated): ").
		SetFieldWidth(40)
	t.searchDatetimeField = tview.NewInputField().
		SetLabel("Datetime (ISO8601 or range): ").
		SetFieldWidth(40)
	t.searchBboxField = tview.NewInputField().
		SetLabel("BBox (minLon,minLat,maxLon,maxLat): ").
		SetFieldWidth(40)
	t.searchLimitField = tview.NewInputField().
		SetLabel("Limit: ").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger)
	t.searchSortField = tview.NewInputField().
		SetLabel("Sort (field:dir or -field): ").
		SetFieldWidth(40)

	t.searchForm = tview.NewForm().
		AddFormItem(t.searchCollectionsField).
		AddFormItem(t.searchDatetimeField).
		AddFormItem(t.searchBboxField).
		AddFormItem(t.searchLimitField).
		AddFormItem(t.searchSortField)
	t.searchForm.AddButton("Submit", func() {
		t.runBasicSearch()
	})
	t.searchForm.AddButton("Cancel", func() {
		t.closeBasicSearchForm()
	})
	t.searchForm.SetCancelFunc(func() {
		t.closeBasicSearchForm()
	})
	t.searchForm.SetBorder(true).SetTitle("Basic Search")

	help := formatting.MakeHelpText("[yellow]Tab[white] next field  [yellow]Enter[white] activate button  [yellow]Esc[white] cancel  [yellow]Ctrl+C[white] quit")
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.searchForm, 0, 1, true).
		AddItem(help, 3, 0, false)

	t.pages.AddPage("searchForm", page, true, false)
}

func (t *TUI) fetchCollections(url string) {
	t.app.QueueUpdateDraw(func() {
		t.collectionsList.Clear()
		t.collectionsList.AddItem("Loading collections...", "", 0, nil)
		t.pages.SwitchToPage("collections")
		t.app.SetFocus(t.collectionsList)
	})

	go func() {
		cli, err := client.NewClient(url)
		if err != nil {
			t.showError(err.Error())
			return
		}
		t.client = cli

		collectionsChan := make(chan []*stac.Collection, 1)
		errorChan := make(chan error, 1)

		go func() {
			var collections []*stac.Collection
			ctx, cancel := context.WithTimeout(t.baseCtx, 30*time.Second)
			defer cancel()

			var fetchErr error
			t.client.GetCollections(ctx)(func(col *stac.Collection, err error) bool {
				if err != nil {
					fetchErr = err
					return false
				}
				collections = append(collections, col)
				return true
			})

			if fetchErr != nil {
				errorChan <- fetchErr
			} else {
				collectionsChan <- collections
			}
		}()

		select {
		case <-t.baseCtx.Done():
			return
		case collections := <-collectionsChan:
			t.cols = collections
			t.app.QueueUpdateDraw(func() {
				t.collectionsList.Clear()
				for _, col := range t.cols {
					collection := col
					t.collectionsList.AddItem(col.Title, "", 0, func() {
						go t.fetchItems(collection.Id)
					})
				}
			})
		case err := <-errorChan:
			t.showError(err.Error())
		case <-time.After(31 * time.Second):
			t.showError("Timeout fetching collections")
		}
	}()
}

func (t *TUI) fetchItems(collectionID string) {
	label := fmt.Sprintf("Collection: %s", collectionID)
	metadata := map[string]string{"collection_id": collectionID}

	t.activeResultLabel = label
	t.lastSearchMetadata = metadata

	t.app.QueueUpdateDraw(func() {
		t.itemsList.Clear()
		t.itemSummary.Clear()
		t.itemsList.AddItem("Loading items…", "", 0, nil)
		t.itemsList.SetTitle(t.itemsListTitle(true))
		t.updateItemsHelp()
		t.pages.SwitchToPage("items")
		t.app.SetFocus(t.itemsList)
	})

	ctx, cancel := context.WithTimeout(t.baseCtx, 300*time.Second)
	seq := t.client.GetItems(ctx, collectionID)
	t.startItemStream(label, metadata, seq, cancel)
}

func (t *TUI) openBasicSearchForm() {
	if t.client == nil || t.searchForm == nil {
		t.showError("Load collections before running a search.")
		return
	}

	currentPage, _ := t.pages.GetFrontPage()
	if currentPage != "collections" && currentPage != "items" {
		return
	}

	t.searchReturnPage = currentPage

	var collectionsValue string
	if currentPage == "collections" {
		index := t.collectionsList.GetCurrentItem()
		if index >= 0 && index < len(t.cols) {
			collectionsValue = t.cols[index].Id
		}
	}
	if collectionsValue == "" && t.lastSearchMetadata != nil {
		if v, ok := t.lastSearchMetadata["collections"]; ok && v != "" {
			collectionsValue = v
		} else if v, ok := t.lastSearchMetadata["collection_id"]; ok && v != "" {
			collectionsValue = v
		}
	}
	if collectionsValue == "" {
		collectionsValue = strings.TrimSpace(t.searchCollectionsField.GetText())
	}

	datetimeValue := strings.TrimSpace(t.searchDatetimeField.GetText())
	bboxValue := strings.TrimSpace(t.searchBboxField.GetText())
	limitValue := strings.TrimSpace(t.searchLimitField.GetText())
	sortValue := strings.TrimSpace(t.searchSortField.GetText())
	if t.lastSearchMetadata != nil {
		if v, ok := t.lastSearchMetadata["datetime"]; ok && v != "" {
			datetimeValue = v
		}
		if v, ok := t.lastSearchMetadata["bbox"]; ok && v != "" {
			bboxValue = v
		}
		if v, ok := t.lastSearchMetadata["limit"]; ok && v != "" {
			limitValue = v
		}
		if v, ok := t.lastSearchMetadata["sort"]; ok && v != "" {
			sortValue = v
		}
	}

	t.searchCollectionsField.SetText(collectionsValue)
	t.searchDatetimeField.SetText(datetimeValue)
	t.searchBboxField.SetText(bboxValue)
	t.searchLimitField.SetText(limitValue)
	t.searchSortField.SetText(sortValue)

	t.pages.SwitchToPage("searchForm")
	t.app.SetFocus(t.searchCollectionsField)
}

func (t *TUI) closeBasicSearchForm() {
	returnPage := t.searchReturnPage
	t.searchReturnPage = ""
	if returnPage == "" {
		returnPage = "collections"
	}
	t.pages.SwitchToPage(returnPage)
	switch returnPage {
	case "collections":
		t.app.SetFocus(t.collectionsList)
	case "items":
		t.app.SetFocus(t.itemsList)
	default:
		t.app.SetFocus(t.collectionsList)
	}
}

func (t *TUI) runBasicSearch() {
	if t.client == nil {
		t.showError("Load collections before running a search.")
		return
	}

	splitList := func(input string) []string {
		fields := strings.FieldsFunc(input, func(r rune) bool {
			switch r {
			case ',', ';', '\n', '\t', ' ':
				return true
			default:
				return false
			}
		})
		var result []string
		for _, f := range fields {
			value := strings.TrimSpace(f)
			if value != "" {
				result = append(result, value)
			}
		}
		return result
	}

	params := client.SearchParams{}
	metadata := map[string]string{}

	collectionsText := strings.TrimSpace(t.searchCollectionsField.GetText())
	if collectionsText != "" {
		metadata["collections"] = collectionsText
		params.Collections = splitList(collectionsText)
	}

	datetimeText := strings.TrimSpace(t.searchDatetimeField.GetText())
	if datetimeText != "" {
		params.Datetime = datetimeText
		metadata["datetime"] = datetimeText
	}

	bboxText := strings.TrimSpace(t.searchBboxField.GetText())
	if bboxText != "" {
		metadata["bbox"] = bboxText
		bboxParts := splitList(bboxText)
		for _, part := range bboxParts {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				t.showError(fmt.Sprintf("invalid bbox coordinate %q", part))
				return
			}
			params.Bbox = append(params.Bbox, value)
		}
		if len(params.Bbox) < 4 || len(params.Bbox)%2 != 0 {
			t.showError("bbox must contain an even number of coordinates (minimum four)")
			return
		}
	}

	limitText := strings.TrimSpace(t.searchLimitField.GetText())
	if limitText != "" {
		metadata["limit"] = limitText
		limitValue, err := strconv.Atoi(limitText)
		if err != nil || limitValue <= 0 {
			t.showError("limit must be a positive integer")
			return
		}
		params.Limit = limitValue
	}

	sortText := strings.TrimSpace(t.searchSortField.GetText())
	if sortText != "" {
		metadata["sort"] = sortText
		sortParts := strings.Split(sortText, ",")
		for _, raw := range sortParts {
			part := strings.TrimSpace(raw)
			if part == "" {
				continue
			}

			direction := "asc"
			field := part
			if strings.Contains(part, ":") {
				pieces := strings.SplitN(part, ":", 2)
				field = strings.TrimSpace(pieces[0])
				if len(pieces) > 1 {
					dir := strings.TrimSpace(pieces[1])
					if strings.EqualFold(dir, "desc") {
						direction = "desc"
					}
				}
			} else if strings.HasPrefix(part, "-") {
				field = strings.TrimSpace(strings.TrimPrefix(part, "-"))
				direction = "desc"
			} else if strings.HasPrefix(part, "+") {
				field = strings.TrimSpace(strings.TrimPrefix(part, "+"))
			}

			if field == "" {
				t.showError("sort field cannot be empty")
				return
			}

			params.SortBy = append(params.SortBy, client.SortField{Field: field, Direction: direction})
		}
	}

	label := "Search"
	switch {
	case len(params.Collections) == 1:
		label = fmt.Sprintf("Search – %s", params.Collections[0])
	case len(params.Collections) > 1:
		label = fmt.Sprintf("Search – %d collections", len(params.Collections))
	case datetimeText != "":
		label = fmt.Sprintf("Search – %s", datetimeText)
	}

	t.searchReturnPage = ""

	t.app.QueueUpdateDraw(func() {
		t.itemsList.Clear()
		t.itemSummary.Clear()
		t.itemsList.AddItem("Loading items…", "", 0, nil)
		t.itemsList.SetTitle(t.itemsListTitle(true))
		t.updateItemsHelp()
		t.pages.SwitchToPage("items")
		t.app.SetFocus(t.itemsList)
	})

	ctx, cancel := context.WithTimeout(t.baseCtx, 300*time.Second)
	seq := t.client.SearchSimple(ctx, params)
	t.startItemStream(label, metadata, seq, cancel)
}

func (t *TUI) startItemStream(label string, metadata map[string]string, seq iter.Seq2[*stac.Item, error], cancel context.CancelFunc) {
	t.cancelItemIteration()

	t.items = nil
	t.currentItem = nil
	t.activeResultLabel = label
	t.lastSearchMetadata = metadata

	t.itemLoadingMutex.Lock()
	t.isLoadingItems = false
	t.isExhausted = false
	t.itemLoadingMutex.Unlock()

	t.stacItemsIteratorCancel = cancel
	next, stop := iter.Pull2(seq)
	t.stacItemsIterator = next
	t.stacItemsIteratorStop = stop

	t.app.QueueUpdateDraw(func() {
		t.updateItemsHelp()
	})

	t.loadNextPage()
}

func (t *TUI) loadNextPage() {
	t.itemLoadingMutex.Lock()
	if t.isLoadingItems || t.isExhausted {
		t.itemLoadingMutex.Unlock()
		return
	}
	if err := t.baseCtx.Err(); err != nil {
		t.itemLoadingMutex.Unlock()
		return
	}
	t.isLoadingItems = true
	t.itemLoadingMutex.Unlock()

	t.app.QueueUpdateDraw(func() {
		t.itemsList.SetTitle(t.itemsListTitle(true))
		if c := t.itemsList.GetItemCount(); c > 0 {
			main, _ := t.itemsList.GetItemText(c - 1)
			if main == "Load more" || main == "Loading items…" {
				t.itemsList.RemoveItem(c - 1)
			}
		}
	})

	go func() {
		var batch []*stac.Item
		exhausted := false
		var pullErr error

		if err := t.baseCtx.Err(); err != nil {
			pullErr = err
			exhausted = true
		} else {
			for i := 0; i < t.pageSize; i++ {
				if t.stacItemsIterator == nil {
					pullErr = fmt.Errorf("no iterator initialized")
					break
				}
				item, err, ok := t.stacItemsIterator()
				if err != nil {
					pullErr = err
					exhausted = true
					break
				}
				if !ok {
					exhausted = true
					break
				}
				batch = append(batch, item)
			}
		}

		t.app.QueueUpdateDraw(func() {
			t.itemsList.SetTitle(t.itemsListTitle(false))
			if c := t.itemsList.GetItemCount(); c > 0 {
				main, _ := t.itemsList.GetItemText(c - 1)
				if main == "Loading items…" {
					t.itemsList.RemoveItem(c - 1)
				}
			}

			if pullErr != nil {
				t.showError(pullErr.Error())
			}

			t.items = append(t.items, batch...)

			for _, it := range batch {
				t.itemsList.AddItem(it.Id, "", 0, func() {
					t.showItemDetail(it)
				})
			}

			if exhausted || pullErr != nil {
				t.isExhausted = true
				if len(batch) == 0 && t.itemsList.GetItemCount() == 0 {
					t.itemsList.AddItem("No items found.", "", 0, nil)
				} else {
					t.itemsList.AddItem("No more items.", "", 0, nil)
				}
				if t.stacItemsIteratorStop != nil {
					t.stacItemsIteratorStop()
					t.stacItemsIteratorStop = nil
				}
				if t.stacItemsIteratorCancel != nil {
					t.stacItemsIteratorCancel()
					t.stacItemsIteratorCancel = nil
				}
			} else {
				t.itemsList.AddItem("Load more", "", 0, nil)
			}

			if t.itemsList.GetItemCount() > 0 && t.itemsList.GetCurrentItem() < 0 {
				t.itemsList.SetCurrentItem(0)
			}
		})

		t.itemLoadingMutex.Lock()
		t.isLoadingItems = false
		t.itemLoadingMutex.Unlock()
	}()
}

func (t *TUI) showInfo(message string) {
	t.app.QueueUpdateDraw(func() {
		modal := tview.NewModal().
			SetText(message).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				t.pages.HidePage("info")
			})
		t.pages.RemovePage("info")
		t.pages.AddPage("info", modal, false, true)
		t.pages.ShowPage("info")
	})
}

func (t *TUI) showError(message string) {
	t.app.QueueUpdateDraw(func() {
		modal := tview.NewModal().
			SetText(message).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				t.pages.HidePage("error")
			})
		t.pages.RemovePage("error")
		t.pages.AddPage("error", modal, false, true)
		t.pages.ShowPage("error")
	})
}

package main

import (
	"context"
	"fmt"
	"iter"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/cmd/tui/formatting"
	"github.com/robert-malhotra/go-stac-client/pkg/client"
)

func (t *TUI) setupPages() {
	t.setupInputPage()
	t.setupCollectionsPage()
	t.setupSearchFormPage()
	t.setupItemsPage()
	t.setupItemDetailPage()
}

const (
	searchPageID       = "search"
	searchHelpControls = "[yellow]↑/↓[white] navigate  [yellow]Enter/Space[white] toggle selection  [yellow]Tab[white] switch focus  [yellow]Esc[white] cancel  [yellow]Ctrl+C[white] quit"
	itemsHelpControls  = "[yellow]↑/↓[white] select  [yellow]Enter[white] view detail  [yellow]s[white] search (↑/↓ move, Space toggle)  [yellow]j[white] raw JSON  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit"
)

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

	collectionsHelp := formatting.MakeHelpText("[yellow]↑/↓[white] select  [yellow]Enter[white] load items  [yellow]s[white] search (↑/↓ move, Space toggle)  [yellow]j[white] raw JSON  [yellow]Tab[white] toggle focus  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit")
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

func (t *TUI) setupSearchFormPage() {
	t.searchForm = tview.NewForm()
	t.searchForm.SetBorder(true).SetTitle("Search Parameters")
	t.searchForm.SetButtonsAlign(tview.AlignRight)

	t.searchDatetime = tview.NewInputField().
		SetLabel("Datetime: ").
		SetFieldWidth(60).
		SetPlaceholder("YYYY-MM-DD/YYYY-MM-DD or open range")
	t.searchForm.AddFormItem(t.searchDatetime)

	t.searchBbox = tview.NewInputField().
		SetLabel("BBox: ").
		SetFieldWidth(60).
		SetPlaceholder("minLon,minLat,maxLon,maxLat")
	t.searchForm.AddFormItem(t.searchBbox)

	limitField := tview.NewInputField().
		SetLabel("Limit: ").
		SetFieldWidth(10)
	limitField.SetAcceptanceFunc(func(text string, ch rune) bool {
		if ch == 0 { // allow deletions
			return true
		}
		return ch >= '0' && ch <= '9'
	})
	t.searchLimit = limitField
	t.searchForm.AddFormItem(limitField)

	summaryField := tview.NewInputField().
		SetLabel("Selected collections: ").
		SetFieldWidth(60)
	summaryField.SetDisabled(true)
	summaryField.SetText("All collections")
	t.searchSummary = summaryField
	t.searchForm.AddFormItem(summaryField)

	t.searchForm.AddButton("Run search", func() {
		go t.runBasicSearch()
	})
	t.searchForm.AddButton("Cancel", func() {
		t.closeSearchForm()
	})
	t.searchForm.SetCancelFunc(func() {
		t.closeSearchForm()
	})
	t.searchForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event == nil || t.app == nil {
			return event
		}

		switch event.Key() {
		case tcell.KeyUp:
			if t.moveSearchFormFocus(-1) {
				return nil
			}
		case tcell.KeyDown:
			if t.moveSearchFormFocus(1) {
				return nil
			}
		case tcell.KeyTab, tcell.KeyBacktab:
			if t.searchCollectionsList != nil {
				t.app.SetFocus(t.searchCollectionsList)
			}
			return nil
		}

		return event
	})

	t.searchCollectionsList = tview.NewList()
	t.searchCollectionsList.SetBorder(true).SetTitle("Collections")
	t.searchCollectionsList.ShowSecondaryText(false)
	t.searchCollectionsList.SetWrapAround(false)
	t.searchCollectionsList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		t.toggleSearchCollection(index)
	})
	t.searchCollectionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			index := t.searchCollectionsList.GetCurrentItem()
			t.toggleSearchCollection(index)
			return nil
		case tcell.KeyRune:
			if event.Rune() == ' ' {
				index := t.searchCollectionsList.GetCurrentItem()
				t.toggleSearchCollection(index)
				return nil
			}
		}
		return event
	})

	formLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(t.searchCollectionsList, 0, 1, true).
		AddItem(t.searchForm, 0, 1, false)

	help := formatting.MakeHelpText(searchHelpControls)
	searchPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(formLayout, 0, 1, true).
		AddItem(help, 3, 0, false)

	t.pages.AddPage(searchPageID, searchPage, true, false)
}

func (t *TUI) runBasicSearch() {
	if t.client == nil {
		t.showError("No STAC API client is loaded yet")
		return
	}

	ids := t.selectedSearchCollectionIDs()
	params := client.SearchParams{Collections: ids}
	metadata := map[string]string{}
	if len(ids) > 0 {
		metadata["collections"] = strings.Join(ids, ",")
	}

	if t.searchDatetime != nil {
		if datetime := strings.TrimSpace(t.searchDatetime.GetText()); datetime != "" {
			params.Datetime = datetime
			metadata["datetime"] = datetime
		}
	}

	if t.searchBbox != nil {
		if bboxText := strings.TrimSpace(t.searchBbox.GetText()); bboxText != "" {
			bbox, normalized, err := parseBBoxInput(bboxText)
			if err != nil {
				t.showError(err.Error())
				return
			}
			params.Bbox = bbox
			metadata["bbox"] = normalized
		}
	}

	if t.searchLimit != nil {
		if limitText := strings.TrimSpace(t.searchLimit.GetText()); limitText != "" {
			limit, err := strconv.Atoi(limitText)
			if err != nil {
				t.showError("Limit must be a positive integer")
				return
			}
			if limit <= 0 {
				t.showError("Limit must be greater than zero")
				return
			}
			params.Limit = limit
			metadata["limit"] = limitText
		}
	}

	summary := t.searchSummaryText(ids)
	label := fmt.Sprintf("Search – %s", summary)

	t.app.QueueUpdateDraw(func() {
		t.pages.HidePage(searchPageID)
		t.pages.SwitchToPage("items")
		t.itemsList.Clear()
		t.itemSummary.Clear()
		t.itemsList.AddItem("Loading items…", "", 0, nil)
		t.itemsList.SetTitle(t.itemsListTitle(true))
		t.updateItemsHelp()
		t.app.SetFocus(t.itemsList)
	})

	ctx, cancel := context.WithTimeout(t.baseCtx, 300*time.Second)
	seq := t.client.SearchSimple(ctx, params)
	t.searchReturnPage = ""
	t.startItemStream(label, metadata, seq, cancel)
}

func (t *TUI) openBasicSearchForm() {
	if len(t.cols) == 0 {
		return
	}

	if t.searchSelectedCollections == nil {
		t.searchSelectedCollections = make(map[string]bool)
	}

	t.ensureSearchSelectionsValid()
	t.rebuildSearchCollectionsList()
	t.populateSearchFormFields()
	t.updateSearchCollectionsSummary()

	currentPage, _ := t.pages.GetFrontPage()
	t.searchReturnPage = currentPage

	highlight := t.searchDefaultCollectionIndex()
	if highlight >= 0 && highlight < t.searchCollectionsList.GetItemCount() {
		t.searchCollectionsList.SetCurrentItem(highlight)
	} else if t.searchCollectionsList.GetItemCount() > 0 {
		t.searchCollectionsList.SetCurrentItem(0)
	}

	t.pages.ShowPage(searchPageID)
	t.pages.SwitchToPage(searchPageID)
	t.app.SetFocus(t.searchCollectionsList)
}

func (t *TUI) focusSearchFormFirstField() {
	if t.searchForm == nil {
		return
	}

	if index := t.searchFormItemIndex(t.searchDatetime); index >= 0 && t.isSearchFormIndexFocusable(index) {
		t.setSearchFormFocus(index)
		return
	}

	total := t.searchForm.GetFormItemCount() + t.searchForm.GetButtonCount()
	for index := 0; index < total; index++ {
		if t.isSearchFormIndexFocusable(index) {
			t.setSearchFormFocus(index)
			return
		}
	}
}

func (t *TUI) focusSearchFormLastElement() {
	if t.searchForm == nil {
		return
	}

	total := t.searchForm.GetFormItemCount() + t.searchForm.GetButtonCount()
	for index := total - 1; index >= 0; index-- {
		if t.isSearchFormIndexFocusable(index) {
			t.setSearchFormFocus(index)
			return
		}
	}
}

func (t *TUI) moveSearchFormFocus(delta int) bool {
	if t.searchForm == nil || delta == 0 {
		return false
	}

	total := t.searchForm.GetFormItemCount() + t.searchForm.GetButtonCount()
	if total == 0 {
		return false
	}

	current := t.searchFormFocusIndex()
	if current < 0 {
		if delta > 0 {
			t.focusSearchFormFirstField()
		} else {
			t.focusSearchFormLastElement()
		}
		return true
	}

	next := current + delta
	for next >= 0 && next < total {
		if t.isSearchFormIndexFocusable(next) {
			t.setSearchFormFocus(next)
			return true
		}
		next += delta
	}

	return false
}

func (t *TUI) searchFormFocusIndex() int {
	if t.searchForm == nil {
		return -1
	}

	itemCount := t.searchForm.GetFormItemCount()
	for index := 0; index < itemCount; index++ {
		if item := t.searchForm.GetFormItem(index); item != nil && item.HasFocus() {
			return index
		}
	}

	buttonCount := t.searchForm.GetButtonCount()
	for index := 0; index < buttonCount; index++ {
		if button := t.searchForm.GetButton(index); button != nil && button.HasFocus() {
			return itemCount + index
		}
	}

	return -1
}

func (t *TUI) isSearchFormIndexFocusable(index int) bool {
	if t.searchForm == nil {
		return false
	}

	itemCount := t.searchForm.GetFormItemCount()
	buttonCount := t.searchForm.GetButtonCount()
	total := itemCount + buttonCount
	if index < 0 || index >= total {
		return false
	}

	if index < itemCount {
		item := t.searchForm.GetFormItem(index)
		if item == nil {
			return false
		}
		if t.searchSummary != nil && item == t.searchSummary {
			return false
		}
		return true
	}

	buttonIndex := index - itemCount
	button := t.searchForm.GetButton(buttonIndex)
	if button == nil {
		return false
	}
	return !button.IsDisabled()
}

func (t *TUI) setSearchFormFocus(index int) {
	if t.searchForm == nil {
		return
	}

	itemCount := t.searchForm.GetFormItemCount()
	buttonCount := t.searchForm.GetButtonCount()
	total := itemCount + buttonCount
	if index < 0 || index >= total {
		return
	}

	t.searchForm.SetFocus(index)

	if t.app == nil {
		return
	}

	if index < itemCount {
		if item := t.searchForm.GetFormItem(index); item != nil {
			if primitive, ok := item.(tview.Primitive); ok {
				t.app.SetFocus(primitive)
			}
		}
		return
	}

	buttonIndex := index - itemCount
	if button := t.searchForm.GetButton(buttonIndex); button != nil {
		t.app.SetFocus(button)
	}
}

func (t *TUI) searchFormItemIndex(target tview.FormItem) int {
	if t.searchForm == nil || target == nil {
		return -1
	}

	itemCount := t.searchForm.GetFormItemCount()
	for index := 0; index < itemCount; index++ {
		if t.searchForm.GetFormItem(index) == target {
			return index
		}
	}

	return -1
}

func (t *TUI) closeSearchForm() {
	returnPage := t.searchReturnPage
	if returnPage == "" {
		returnPage = "collections"
	}
	t.searchReturnPage = ""

	switch returnPage {
	case "items":
		t.pages.SwitchToPage("items")
		t.app.SetFocus(t.itemsList)
	case "collections":
		fallthrough
	default:
		t.pages.SwitchToPage("collections")
		t.app.SetFocus(t.collectionsList)
	}
}

func (t *TUI) ensureSearchSelectionsValid() {
	valid := make(map[string]struct{}, len(t.cols))
	for _, col := range t.cols {
		valid[col.Id] = struct{}{}
	}

	filtered := t.searchSelectedOrder[:0]
	for _, id := range t.searchSelectedOrder {
		if _, ok := valid[id]; ok && t.searchSelectedCollections[id] {
			filtered = append(filtered, id)
		}
	}
	t.searchSelectedOrder = filtered

	for id := range t.searchSelectedCollections {
		if _, ok := valid[id]; !ok {
			delete(t.searchSelectedCollections, id)
		}
	}
}

func (t *TUI) rebuildSearchCollectionsList() {
	if t.searchCollectionsList == nil {
		return
	}

	t.searchCollectionsList.Clear()
	for _, col := range t.cols {
		main, secondary := t.searchCollectionListTexts(col)
		t.searchCollectionsList.AddItem(main, secondary, 0, nil)
	}
}

func (t *TUI) searchCollectionListTexts(col *stac.Collection) (string, string) {
	if col == nil {
		return "", ""
	}
	checked := t.searchSelectedCollections != nil && t.searchSelectedCollections[col.Id]
	indicator := "[ ]"
	if checked {
		indicator = "[green][x][white]"
	}
	label := strings.TrimSpace(col.Title)
	if label == "" {
		label = col.Id
	}
	main := fmt.Sprintf("%s %s", indicator, label)
	return main, ""
}

func parseBBoxInput(text string) ([]float64, string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, "", nil
	}

	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		switch r {
		case ',', ' ', '\n', '\t':
			return true
		default:
			return false
		}
	})

	if len(parts) == 0 {
		return nil, "", fmt.Errorf("bbox must have 4 or 6 numeric values")
	}

	coords := make([]float64, 0, len(parts))
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, "", fmt.Errorf("bbox must contain only numeric values")
		}
		coords = append(coords, value)
		normalized = append(normalized, strconv.FormatFloat(value, 'f', -1, 64))
	}

	if len(coords) != 4 && len(coords) != 6 {
		return nil, "", fmt.Errorf("bbox must have 4 or 6 numeric values")
	}

	return coords, strings.Join(normalized, ","), nil
}

func (t *TUI) toggleSearchCollection(index int) {
	if index < 0 || index >= len(t.cols) {
		return
	}
	if t.searchSelectedCollections == nil {
		t.searchSelectedCollections = make(map[string]bool)
	}

	col := t.cols[index]
	id := col.Id
	if t.searchSelectedCollections[id] {
		delete(t.searchSelectedCollections, id)
		for i, existing := range t.searchSelectedOrder {
			if existing == id {
				t.searchSelectedOrder = append(t.searchSelectedOrder[:i], t.searchSelectedOrder[i+1:]...)
				break
			}
		}
	} else {
		t.searchSelectedCollections[id] = true
		present := false
		for _, existing := range t.searchSelectedOrder {
			if existing == id {
				present = true
				break
			}
		}
		if !present {
			t.searchSelectedOrder = append(t.searchSelectedOrder, id)
		}
	}

	main, secondary := t.searchCollectionListTexts(col)
	t.searchCollectionsList.SetItemText(index, main, secondary)
	t.updateSearchCollectionsSummary()
}

func (t *TUI) updateSearchCollectionsSummary() {
	if t.searchSummary == nil {
		return
	}
	ids := t.selectedSearchCollectionIDs()
	summary := t.searchSummaryText(ids)
	t.searchSummary.SetText(summary)
}

func (t *TUI) populateSearchFormFields() {
	setField := func(field *tview.InputField, key string) {
		if field == nil {
			return
		}
		value := ""
		if t.lastSearchMetadata != nil {
			if v, ok := t.lastSearchMetadata[key]; ok {
				value = v
			}
		}
		field.SetText(value)
	}

	setField(t.searchDatetime, "datetime")
	setField(t.searchBbox, "bbox")
	setField(t.searchLimit, "limit")
}

func (t *TUI) selectedSearchCollectionIDs() []string {
	if len(t.searchSelectedCollections) == 0 {
		return nil
	}
	var ids []string
	seen := make(map[string]struct{})
	for _, id := range t.searchSelectedOrder {
		if t.searchSelectedCollections[id] {
			ids = append(ids, id)
			seen[id] = struct{}{}
		}
	}
	for _, col := range t.cols {
		if t.searchSelectedCollections[col.Id] {
			if _, ok := seen[col.Id]; !ok {
				ids = append(ids, col.Id)
			}
		}
	}
	return ids
}

func (t *TUI) searchSummaryText(ids []string) string {
	if len(ids) == 0 {
		return "All collections"
	}
	summary := strings.Join(ids, ", ")
	runes := []rune(summary)
	if len(runes) > 60 {
		summary = string(runes[:57]) + "…"
	}
	return summary
}

func (t *TUI) searchDefaultCollectionIndex() int {
	if len(t.searchSelectedOrder) > 0 {
		if idx := t.indexOfCollectionID(t.searchSelectedOrder[0]); idx >= 0 {
			return idx
		}
	}

	if idx := t.collectionsList.GetCurrentItem(); idx >= 0 && idx < len(t.cols) {
		return idx
	}

	if id := t.lastSearchMetadata["collection_id"]; id != "" {
		if idx := t.indexOfCollectionID(id); idx >= 0 {
			return idx
		}
	}
	if list := t.lastSearchMetadata["collections"]; list != "" {
		for _, id := range strings.Split(list, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if idx := t.indexOfCollectionID(id); idx >= 0 {
				return idx
			}
		}
	}

	if len(t.cols) > 0 {
		return 0
	}
	return -1
}

func (t *TUI) indexOfCollectionID(id string) int {
	for i, col := range t.cols {
		if col.Id == id {
			return i
		}
	}
	return -1
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
	t.itemAssets.SetWrapAround(false)

	t.itemAssetDetail = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	t.itemAssetDetail.SetBorder(true).SetTitle("Asset Details")
	t.itemAssetDetail.SetText("Select an asset to view details.")

	t.itemAssets.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		t.updateItemAssetDetail(index)
	})

	t.itemDetailPanes = []tview.Primitive{t.itemProperties, t.itemAssets, t.itemAssetDetail}

	itemDetailHelp := formatting.MakeHelpText("[yellow]Tab[white] next pane  [yellow]Shift+Tab[white] previous pane  [yellow]Enter[white] download asset  [yellow]j[white] raw JSON  [yellow]Esc[white] back  [yellow]Ctrl+C[white] quit")
	itemDetailPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.itemDetail, 0, 1, true).
		AddItem(itemDetailHelp, 3, 0, false)

	t.pages.AddPage("itemDetail", itemDetailPage, true, false)
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
				item := it
				t.itemsList.AddItem(item.Id, "", 0, func() {
					t.showItemDetail(item)
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

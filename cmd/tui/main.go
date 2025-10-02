package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/pkg/client"
	"github.com/robert-malhotra/go-stac-client/pkg/downloader"
)

func main() {
	tui := NewTUI()
	tui.Run()
}

type TUI struct {
	app             *tview.Application
	pages           *tview.Pages
	input           *tview.InputField
	collectionsList *tview.List
	colDetail       *tview.TextView
	itemsList       *tview.List
	itemSummary     *tview.TextView
	itemDetail      *tview.Grid

	// Item detail panes
	itemProperties  *tview.TextView
	itemAssets      *tview.List
	itemLinks       *tview.List
	itemDetailPanes []tview.Primitive
	itemDetailFocus int

	client *client.Client
	cols   []*stac.Collection
	items  []*stac.Item

	// Iterator for items (used synchronously, on-demand)
	stacItemsIterator       func() (*stac.Item, error, bool)
	stacItemsIteratorStop   func()
	stacItemsIteratorCancel context.CancelFunc

	// Paging state
	pageSize       int
	isLoadingItems bool
	isExhausted    bool

	itemLoadingMutex sync.Mutex
}

func NewTUI() *TUI {
	app := tview.NewApplication()

	// Styles
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkSlateGray
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.BorderColor = tcell.ColorWhite
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorYellow
	tview.Styles.TertiaryTextColor = tcell.ColorGreen
	tview.Styles.InverseTextColor = tcell.ColorBlue
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorNavy

	tui := &TUI{
		app:      app,
		pages:    tview.NewPages(),
		pageSize: 10,
	}

	// Input for STAC URL
	tui.input = tview.NewInputField().
		SetLabel("STAC API URL: ").
		SetFieldWidth(60).
		SetText("https://earth-search.aws.element84.com/v1")
	tui.input.SetBorder(true).SetTitle("Enter STAC API Root URL")
	tui.input.SetDoneFunc(tui.onInputDone)

	// Collections UI
	tui.collectionsList = tview.NewList()
	tui.collectionsList.SetBorder(true).SetTitle("Collections")
	tui.colDetail = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	tui.colDetail.SetBorder(true).SetTitle("Collection Details")

	collectionsPage := tview.NewFlex().
		AddItem(tui.collectionsList, 0, 1, true).
		AddItem(tui.colDetail, 0, 1, false)

	tui.collectionsList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < len(tui.cols) {
			col := tui.cols[index]
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("[yellow]Title: [white]%s\n", col.Title))
			builder.WriteString(fmt.Sprintf("[yellow]ID: [white]%s\n", col.Id))
			builder.WriteString(fmt.Sprintf("[yellow]Description: [white]%s\n", col.Description))
			tui.colDetail.SetText(builder.String())
		}
	})

	// Items UI
	tui.itemsList = tview.NewList()
	tui.itemsList.SetBorder(true).SetTitle("Items")
	tui.itemsList.ShowSecondaryText(false)
	tui.itemsList.SetWrapAround(false)

	tui.itemSummary = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	tui.itemSummary.SetBorder(true).SetTitle("Item Summary")

	itemsPage := tview.NewFlex().
		AddItem(tui.itemsList, 0, 1, true).
		AddItem(tui.itemSummary, 0, 1, false)

	tui.itemsList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		// Update summary
		if index < len(tui.items) {
			item := tui.items[index]
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("[yellow]ID: [white]%s\n", item.Id))
			if dt, ok := item.Properties["datetime"].(string); ok {
				builder.WriteString(fmt.Sprintf("[yellow]Datetime: [white]%s\n", dt))
			}
			if p, ok := item.Properties["platform"].(string); ok {
				builder.WriteString(fmt.Sprintf("[yellow]Platform: [white]%s\n", p))
			}
			if c, ok := item.Properties["constellation"].(string); ok {
				builder.WriteString(fmt.Sprintf("[yellow]Constellation: [white]%s\n", c))
			}
		if geomText := formatGeometry(item.Geometry); geomText != "" {
			builder.WriteString("[yellow]Geometry:[white]\n")
			builder.WriteString(geomText)
			if !strings.HasSuffix(geomText, "\n") {
				builder.WriteString("\n")
			}
		}
			tui.itemSummary.SetText(builder.String())
			tui.itemSummary.ScrollToBeginning()
		} else {
			tui.itemSummary.Clear()
		}

		// Pagination
		if index >= tui.itemsList.GetItemCount()-2 {
			lastItem, _ := tui.itemsList.GetItemText(tui.itemsList.GetItemCount() - 1)
			if lastItem == "Load more" {
				go tui.loadNextPage()
			}
		}
	})

	// Item detail UI
	tui.itemDetail = tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0)
	tui.itemDetail.SetBorder(true).SetTitle("Item Detail")

	tui.itemProperties = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	tui.itemProperties.SetBorder(true).SetTitle("Properties")

	tui.itemAssets = tview.NewList()
	tui.itemAssets.SetBorder(true).SetTitle("Assets")
	tui.itemAssets.ShowSecondaryText(false)

	tui.itemLinks = tview.NewList()
	tui.itemLinks.SetBorder(true).SetTitle("Links")
	tui.itemLinks.ShowSecondaryText(false)

	tui.itemDetailPanes = []tview.Primitive{tui.itemProperties, tui.itemAssets, tui.itemLinks}

	// Pages
	tui.pages.AddPage("input", tui.input, true, true)
	tui.pages.AddPage("collections", collectionsPage, true, false)
	tui.pages.AddPage("items", itemsPage, true, false)
	tui.pages.AddPage("itemDetail", tui.itemDetail, true, false)

	// Global key handling
	tui.app.SetInputCapture(tui.onInputCapture)

	return tui
}

func (t *TUI) Run() {
	if err := t.app.SetRoot(t.pages, true).Run(); err != nil {
		panic(err)
	}
}

func (t *TUI) onInputDone(key tcell.Key) {
	if key == tcell.KeyEnter {
		url := t.input.GetText()
		go t.fetchCollections(url)
	}
}

func (t *TUI) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlC {
		t.app.Stop()
		return nil
	}

	currentPage, _ := t.pages.GetFrontPage()

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

	if event.Key() == tcell.KeyEscape {
		switch currentPage {
		case "itemDetail":
			t.pages.SwitchToPage("items")
			return nil
		case "items":
			t.pages.SwitchToPage("collections")
			return nil
		case "collections":
			t.pages.SwitchToPage("input")
			return nil
		}
	}

	return event
}

func (t *TUI) fetchCollections(url string) {
	t.app.QueueUpdateDraw(func() {
		t.collectionsList.Clear()
		t.collectionsList.AddItem("Loading collections...", "", 0, nil)
		t.pages.SwitchToPage("collections")
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
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		case collections := <-collectionsChan:
			t.cols = collections
			t.app.QueueUpdateDraw(func() {
				t.collectionsList.Clear()
				for _, col := range t.cols {
					collection := col
					t.collectionsList.AddItem(col.Title, col.Id, 0, func() {
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
	t.app.QueueUpdateDraw(func() {
		t.itemsList.Clear()
		t.itemSummary.Clear()
		t.itemsList.AddItem("Loading items…", "", 0, nil)
		t.pages.SwitchToPage("items")
	})

	t.items = nil

	if t.stacItemsIteratorCancel != nil {
		t.stacItemsIteratorCancel()
		t.stacItemsIteratorCancel = nil
	}
	if t.stacItemsIteratorStop != nil {
		t.stacItemsIteratorStop()
		t.stacItemsIteratorStop = nil
	}

	t.itemLoadingMutex.Lock()
	t.isLoadingItems = false
	t.isExhausted = false
	t.itemLoadingMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	t.stacItemsIteratorCancel = cancel

	seq := t.client.GetItems(ctx, collectionID)
	next, stop := iter.Pull2(seq)
	t.stacItemsIterator = next
	t.stacItemsIteratorStop = stop

	t.loadNextPage()
}

func (t *TUI) loadNextPage() {
	t.itemLoadingMutex.Lock()
	if t.isLoadingItems || t.isExhausted {
		t.itemLoadingMutex.Unlock()
		return
	}
	t.isLoadingItems = true
	t.itemLoadingMutex.Unlock()

	t.app.QueueUpdateDraw(func() {
		t.itemsList.SetTitle("Items (loading...)")
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

		t.app.QueueUpdateDraw(func() {
			t.itemsList.SetTitle("Items")
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
				itCopy := it
				t.itemsList.AddItem(it.Id, "", 0, func() {
					t.showItemDetail(itCopy)
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

func (t *TUI) showItemDetail(item *stac.Item) {
	// Properties
	t.itemProperties.Clear()
	t.itemProperties.SetText(t.populateProperties(item.Properties, 0))
	t.itemProperties.ScrollToBeginning()

	// Assets
	t.buildAssetsView(item)

	// Links
	t.itemLinks.Clear()
	for _, link := range item.Links {
		t.itemLinks.AddItem(fmt.Sprintf("%s (%s)", link.Rel, link.Type), link.Href, 0, nil)
	}
	t.itemLinks.SetWrapAround(false)
	t.itemLinks.SetCurrentItem(0)

	rightPane := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.itemAssets, 0, 1, false).
		AddItem(t.itemLinks, 0, 1, false)

	t.itemDetail.Clear().
		AddItem(t.itemProperties, 0, 0, 1, 1, 0, 0, true).
		AddItem(rightPane, 0, 1, 1, 1, 0, 0, false)

	t.itemDetailFocus = 0
	t.app.SetFocus(t.itemDetailPanes[t.itemDetailFocus])

	t.pages.SwitchToPage("itemDetail")
}

func formatGeometry(geometry interface{}) string {
	if geometry == nil {
		return ""
	}

	raw, err := json.Marshal(geometry)
	if err != nil {
		return fmt.Sprintf("%v", geometry)
	}

	var geo map[string]interface{}
	if err := json.Unmarshal(raw, &geo); err != nil {
		return string(raw)
	}

	var sections []string

	if typ, ok := geo["type"].(string); ok {
		sections = append(sections, typ)

		if typ == "GeometryCollection" {
			if geoms, ok := geo["geometries"].([]interface{}); ok {
				var summaries []string
				for _, g := range geoms {
					if s := formatGeometry(g); s != "" {
						summaries = append(summaries, s)
					}
				}
				if len(summaries) > 0 {
					sections = append(sections, strings.Join(summaries, " | "))
				}
			}
		}
	}

	if coords, ok := geo["coordinates"]; ok {
		if coordStr := formatCoordinateValue(coords, 0); coordStr != "" {
			sections = append(sections, wrapCoordinateString(coordStr, 70))
		}
	}

	if bbox, ok := geo["bbox"]; ok {
		if bboxStr := formatCoordinateValue(bbox, 0); bboxStr != "" {
			sections = append(sections, "bbox "+wrapCoordinateString(bboxStr, 70))
		}
	}

	if len(sections) == 0 {
		return string(raw)
	}

	return strings.Join(sections, "\n")
}

func formatCoordinateValue(value interface{}, depth int) string {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		parts := make([]string, len(v))
		for i, elem := range v {
			parts[i] = formatCoordinateValue(elem, depth+1)
		}
		sep := ", "
		if depth >= 2 {
			sep = " "
		}
		return "[" + strings.Join(parts, sep) + "]"
	case float64:
		return strconv.FormatFloat(v, 'f', 5, 64)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func wrapCoordinateString(s string, width int) string {
	if len(s) <= width || width <= 0 {
		return s
	}

	var out strings.Builder
	lineLen := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if lineLen >= width && (ch == ',' || ch == ']' || ch == ' ') {
			out.WriteByte('\n')
			lineLen = 0
			if ch == ' ' {
				continue
			}
		}

		out.WriteByte(ch)
		lineLen++
	}

	return out.String()
}

func (t *TUI) populateProperties(properties map[string]interface{}, indent int) string {
	var builder strings.Builder
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := properties[key]
		indentedKey := fmt.Sprintf("%s%s:", strings.Repeat("  ", indent), key)
		builder.WriteString(fmt.Sprintf("[yellow]%-30s[white]", indentedKey))

		jsonBytes, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			builder.WriteString(" Error marshalling value\n")
		} else {
			builder.WriteString(fmt.Sprintf(" %s\n", string(jsonBytes)))
		}
	}
	return builder.String()
}

func (t *TUI) buildAssetsView(item *stac.Item) {
	t.itemAssets.Clear()
	for key, asset := range item.Assets {
		assetCopy := asset
		t.itemAssets.AddItem(key, asset.Title, 0, func() {
			t.downloadAsset(assetCopy)
		})
	}
	t.itemAssets.SetWrapAround(false)
	t.itemAssets.SetCurrentItem(0)
}

func (t *TUI) downloadAsset(asset *stac.Asset) {
	progress := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			t.app.Draw()
		})
	progress.SetBorder(true).SetTitle("Download Progress")

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Downloading %s", asset.Href)).
		AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			t.pages.HidePage("download")
		})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(modal, 0, 1, true).
		AddItem(progress, 0, 1, false)

	t.pages.AddPage("download", layout, true, true)

	go func() {
		dest := getOutputFilename(asset.Href)
		err := downloader.Download(context.Background(), asset.Href, dest)
		t.app.QueueUpdateDraw(func() {
			if err != nil {
				t.showError(fmt.Sprintf("Download failed: %v", err))
			} else {
				modal.SetText(fmt.Sprintf("Asset downloaded to %s", dest))
				progress.SetText("Download complete!")
			}
		})
	}()
}

func getOutputFilename(assetUrl string) string {
	parts := strings.Split(assetUrl, "/")
	return parts[len(parts)-1]
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

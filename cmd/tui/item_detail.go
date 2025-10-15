package main

import (
	"fmt"

	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/cmd/tui/formatting"
)

func (t *TUI) showItemDetail(item *stac.Item) {
	t.currentItem = item

	// Properties
	t.itemProperties.Clear()
	t.itemProperties.SetText(formatting.FormatProperties(item.Properties, 0))
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

func (t *TUI) buildAssetsView(item *stac.Item) {
	t.itemAssets.Clear()
	for key, asset := range item.Assets {
		t.itemAssets.AddItem(key, asset.Title, 0, func() {
			t.downloadAsset(asset)
		})
	}
	t.itemAssets.SetWrapAround(false)
	t.itemAssets.SetCurrentItem(0)
}

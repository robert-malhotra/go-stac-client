package main

import (
	"context"
	"sync"

	"github.com/gdamore/tcell/v2"
	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/pkg/client"
)

type TUI struct {
	app             *tview.Application
	pages           *tview.Pages
	input           *tview.InputField
	collectionsList *tview.List
	colDetail       *tview.TextView
	itemsList       *tview.List
	itemSummary     *tview.TextView
	itemsHelp       *tview.TextView
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

	activeResultLabel  string
	lastSearchMetadata map[string]string

	currentItem *stac.Item

	// Iterator for items (used synchronously, on-demand)
	stacItemsIterator       func() (*stac.Item, error, bool)
	stacItemsIteratorStop   func()
	stacItemsIteratorCancel context.CancelFunc

	// Paging state
	pageSize       int
	isLoadingItems bool
	isExhausted    bool

	itemLoadingMutex sync.Mutex

	baseCtx    context.Context
	baseCancel context.CancelFunc
	stopOnce   sync.Once

	downloadMu     sync.Mutex
	activeDownload *downloadSession

	jsonViewer *jsonViewer
}

func NewTUI(ctx context.Context) *TUI {
	if ctx == nil {
		ctx = context.Background()
	}
	baseCtx, baseCancel := context.WithCancel(ctx)
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
		app:        app,
		pages:      tview.NewPages(),
		pageSize:   10,
		baseCtx:    baseCtx,
		baseCancel: baseCancel,
	}

	tui.setupPages()
	tui.jsonViewer = newJSONViewer(tui)

	// Global key handling
	tui.app.SetInputCapture(tui.onInputCapture)
	tui.app.SetFocus(tui.input)

	return tui
}

func (t *TUI) Run() {
	if err := t.app.SetRoot(t.pages, true).Run(); err != nil {
		panic(err)
	}
}

func (t *TUI) Stop() {
	t.stopOnce.Do(func() {
		if t.baseCancel != nil {
			t.baseCancel()
		}
		t.cancelActiveDownload()
		t.cancelItemIteration()
		t.app.Stop()
	})
}

func (t *TUI) cancelItemIteration() {
	if t.stacItemsIteratorCancel != nil {
		t.stacItemsIteratorCancel()
		t.stacItemsIteratorCancel = nil
	}
	if t.stacItemsIteratorStop != nil {
		t.stacItemsIteratorStop()
		t.stacItemsIteratorStop = nil
	}
}

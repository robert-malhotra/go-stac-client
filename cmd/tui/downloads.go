package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	stac "github.com/planetlabs/go-stac"
	"github.com/rivo/tview"
	"github.com/robert-malhotra/go-stac-client/cmd/tui/formatting"
	"github.com/robert-malhotra/go-stac-client/pkg/downloader"
)

type downloadSession struct {
	cancel func()
}

func (t *TUI) setActiveDownload(session *downloadSession) {
	t.downloadMu.Lock()
	t.activeDownload = session
	t.downloadMu.Unlock()
}

func (t *TUI) clearActiveDownload(session *downloadSession) {
	t.downloadMu.Lock()
	if t.activeDownload == session {
		t.activeDownload = nil
	}
	t.downloadMu.Unlock()
}

func (t *TUI) cancelActiveDownload() {
	t.downloadMu.Lock()
	session := t.activeDownload
	t.activeDownload = nil
	t.downloadMu.Unlock()

	if session != nil && session.cancel != nil {
		session.cancel()
	}
}

func (t *TUI) downloadAsset(asset *stac.Asset) {
	if err := t.baseCtx.Err(); err != nil {
		return
	}

	ctx, cancel := context.WithCancel(t.baseCtx)

	progress := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			t.app.Draw()
		})
	progress.SetBorder(true).SetTitle("Download Progress")
	progress.SetText("Preparing download...")

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Downloading %s", asset.Href)).
		AddButtons([]string{"Cancel"})

	var closeOnce sync.Once
	closeDownloadPage := func() {
		closeOnce.Do(func() {
			t.app.QueueUpdateDraw(func() {
				t.pages.HidePage("download")
				t.pages.RemovePage("download")
			})
		})
	}

	var (
		cancelOnce    sync.Once
		userCancelled atomic.Bool
	)

	session := &downloadSession{
		cancel: func() {
			cancelOnce.Do(func() {
				userCancelled.Store(true)
				cancel()
				closeDownloadPage()
			})
		},
	}
	t.setActiveDownload(session)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		closeDownloadPage()
		session.cancel()
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(modal, 0, 1, true).
		AddItem(progress, 0, 1, false)

	t.pages.RemovePage("download")
	t.pages.AddPage("download", layout, true, true)
	t.app.SetFocus(modal)

	dest := formatting.GetOutputFilename(asset.Href)

	updateProgress := func(downloaded, total int64) {
		if userCancelled.Load() {
			return
		}
		progressText := formatting.RenderDownloadProgress(downloaded, total)
		progress.SetText(progressText)
		modal.SetText(fmt.Sprintf("Downloading %s\n%s", asset.Href, progressText))
	}

	go func() {
		defer cancel()
		defer t.clearActiveDownload(session)

		err := downloader.DownloadWithProgress(ctx, asset.Href, dest, func(downloaded, total int64) {
			t.app.QueueUpdateDraw(func() { updateProgress(downloaded, total) })
		})

		if userCancelled.Load() || errors.Is(err, context.Canceled) {
			return
		}

		if err != nil {
			closeDownloadPage()
			t.showError(fmt.Sprintf("Download failed: %v", err))
			return
		}

		t.app.QueueUpdateDraw(func() {
			if userCancelled.Load() {
				return
			}
			progress.SetText("Download complete!")
			modal.SetText(fmt.Sprintf("Asset downloaded to %s", dest))
			modal.ClearButtons()
			modal.AddButtons([]string{"Close"})
			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				closeDownloadPage()
			})
			t.app.SetFocus(modal)
		})
	}()
}

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
	tui       *TUI
	cancel    func()
	prevFocus tview.Primitive
	prevPage  string

	restoreOnce sync.Once
}

func (s *downloadSession) teardown() {
	if s == nil || s.tui == nil {
		return
	}

	prevPage := s.prevPage
	prevFocus := s.prevFocus

	s.restoreOnce.Do(func() {
		go s.tui.app.QueueUpdateDraw(func() {
			if prevPage != "" {
				s.tui.pages.SwitchToPage(prevPage)
			}

			s.tui.pages.HidePage("download")
			s.tui.pages.RemovePage("download")

			if prevFocus != nil {
				s.tui.app.SetFocus(prevFocus)
			}
		})
	})
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

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Preparing download...\n%s", asset.Href)).
		AddButtons([]string{"Cancel"})

	previousFocus := t.app.GetFocus()
	previousPage, _ := t.pages.GetFrontPage()

	var (
		cancelOnce    sync.Once
		userCancelled atomic.Bool
	)

	session := &downloadSession{
		tui:       t,
		prevFocus: previousFocus,
		prevPage:  previousPage,
	}

	session.cancel = func() {
		cancelOnce.Do(func() {
			userCancelled.Store(true)
			cancel()
			session.teardown()
		})
	}
	t.setActiveDownload(session)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		session.cancel()
	})

	t.app.QueueUpdateDraw(func() {
		t.pages.RemovePage("download")
		t.pages.AddPage("download", modal, true, false)
		t.pages.ShowPage("download")
		t.pages.SwitchToPage("download")
		t.app.SetFocus(modal)
	})

	dest := formatting.GetOutputFilename(asset.Href)

	updateProgress := func(downloaded, total int64) {
		if userCancelled.Load() {
			return
		}
		progressText := formatting.RenderDownloadProgress(downloaded, total)
		modal.SetText(fmt.Sprintf("Downloading %s\n%s", asset.Href, progressText))
	}

	go func() {
		defer cancel()
		defer t.clearActiveDownload(session)

		err := downloader.DownloadWithProgress(ctx, asset.Href, dest, func(downloaded, total int64) {
			t.app.QueueUpdateDraw(func() { updateProgress(downloaded, total) })
		})

		if userCancelled.Load() || errors.Is(err, context.Canceled) {
			session.teardown()
			return
		}

		if err != nil {
			session.teardown()
			t.showError(fmt.Sprintf("Download failed: %v", err))
			return
		}

		t.app.QueueUpdateDraw(func() {
			if userCancelled.Load() {
				return
			}
			modal.SetText(fmt.Sprintf("Asset downloaded to %s", dest))
			modal.ClearButtons()
			modal.AddButtons([]string{"Close"})
			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				session.teardown()
			})
			t.app.SetFocus(modal)
		})
	}()
}

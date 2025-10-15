package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tui := NewTUI(ctx)
	go func() {
		<-ctx.Done()
		tui.Stop()
	}()

	tui.Run()
}

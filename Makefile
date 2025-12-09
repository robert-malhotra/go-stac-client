GO ?= go

.PHONY: build test tui clean

build:
	$(GO) build ./...

# Run unit tests for all packages.
test:
	$(GO) test ./...

# Launch the Bubble Tea TUI browser.
tui:
	$(GO) run ./cmd/tui

# Remove build artifacts produced by Go.
clean:
	$(GO) clean

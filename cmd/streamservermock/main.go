package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	streamservermock "twitchy-api/internal/external/streamserver/mock"
)

// TODO: dockerize mock (xd)
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := streamservermock.Run(ctx, nil); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			cancel()
		}
	}()

	<-ctx.Done()
}

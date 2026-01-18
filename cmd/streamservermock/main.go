package main

import (
	"context"
	"fmt"
	streamservermock "main/internal/external/streamserver/mock"
	"os"
	"os/signal"
	"syscall"
)

// TODO: dockerize mock (xd)
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := streamservermock.Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			cancel()
		}
	}()

	<-ctx.Done()
}

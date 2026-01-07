package main

import (
	"context"
	"fmt"
	"main/internal/lib/streamservermock"
	"os"
)

// TODO: dockerize mock (xd)
func main() {
	ctx := context.Background()
	if err := streamservermock.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

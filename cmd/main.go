package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"GoopBot/internal/bot"
)

func main() {
	// Initialize bot
	ctx := context.Background()
	b, err := bot.NewBot(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating bot: %v\n", err)
		os.Exit(1)
	}

	// Start bot
	if err := b.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error starting bot: %v\n", err)
		os.Exit(1)
	}

	// Handle shutdown signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan

	// Graceful shutdown
	if err := b.Stop(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error stopping bot: %v\n", err)
		os.Exit(1)
	}
}

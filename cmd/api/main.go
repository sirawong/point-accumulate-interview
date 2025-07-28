package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirawong/point-accumulate-interview/internal/di"
)

func main() {
	app, cleanup, err := di.NewNewApplication()
	if err != nil {
		log.Fatal("Failed to initialize di:", err)
	}
	defer cleanup()

	if err := app.Start(); err != nil {
		log.Fatal("Failed to start servers:", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

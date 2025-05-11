package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/437d5/subpub/internal/app"
	"github.com/437d5/subpub/internal/config"
	"github.com/437d5/subpub/pkg/subpub"
)

func main() {
	cfg := config.Load()
	log.Println("Config loaded")
	bus := subpub.NewSubPubImpl()
	log.Println("New bus system initialized")
	app := app.New(cfg, bus)
	errChan := make(chan error, 1)

	go func() {
		log.Println("Starting new application")
		if err := app.Run(); err != nil {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	select {
	case err := <-errChan:
		log.Fatalf("Application error: %v", err)
	case sig := <-quit:
		log.Printf("Received signal: %v. Shutting down...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		if err := app.Stop(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			os.Exit(1)
		} else {
			log.Println("Server exited properly")
		}
	}
}
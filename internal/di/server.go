package di

import (
	"context"
	"errors"
	"log"
	"net/http"
)

func (a *Application) Start() error {
	go func() {
		log.Println("Starting HTTP server...")
		log.Printf("HTTP server is listening on %s", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server stopped with error: %v", err)
		}
	}()

	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	err := a.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/server"
)

func main() {
	c := config.DefaultConfig()
	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	handler, err := server.NewHashEncoder(c, graceful)

	if err != nil {
		fmt.Printf("Creating the hash encoder service resulted in an error, see below \n %+v", err)
		return
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: handler,
	}

	go func() {
		fmt.Printf("Starting server on port %d\n", c.Port)
		if err := httpServer.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				fmt.Println("Server connection closed")
			default:
				fmt.Printf("Unexpected server error: %+v\n", err)
				graceful <- syscall.SIGKILL
			}
		}
	}()

	<-graceful
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Unexpected server shutdown error: %+v\n", err)
	} else {
		fmt.Println("Server exited properly")
	}
}

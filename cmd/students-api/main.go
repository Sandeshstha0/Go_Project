package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sandesh/students-api/internal/config"
)

func main() {
	// Load configuration (env, storage path, server address) from
	// the YAML file / env vars via cleanenv, using MustLoad from our config package.
	cfg := config.MustLoad()

	// database setup
	// setup router

	// http.NewServeMux is Go's built-in HTTP router.
	router := http.NewServeMux()

	// Register a handler for GET requests to "/".
	// The pattern "GET /" (method + path) requires Go 1.22+.
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to student api"))
	})

	// Build the HTTP server, using the address from config
	// and our router as the handler for all incoming requests.
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	log.Printf("starting server on %s (env: %s)", cfg.Addr, cfg.Env)

	// Create a channel to receive OS signals (like Ctrl+C or a kill command).
	// Buffered with size 1 so the signal package never blocks trying to send to it.
	done := make(chan os.Signal, 1)

	// Tell the OS to forward these specific signals to our "done" channel
	// instead of letting them kill the program immediately:
	//   - os.Interrupt / SIGINT: sent when you press Ctrl+C
	//   - syscall.SIGTERM: sent by process managers (Docker, systemd, k8s)
	//     asking the process to terminate gracefully
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// Run the server in a separate goroutine so it doesn't block main()
	// from continuing on to set up the shutdown-signal listener below.
	go func() {
		err := server.ListenAndServe()

		// ListenAndServe always returns a non-nil error.
		// When we deliberately shut the server down later (via server.Shutdown),
		// it returns http.ErrServerClosed — that's the *expected* exit path,
		// not a real failure, so we must not fatal-crash on it.
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server: ", err)
		}
	}()

	// Block here until a signal arrives on the "done" channel
	// (i.e. until the user hits Ctrl+C or the OS sends SIGTERM).
	// Everything below this line only runs once that happens.
	<-done

	slog.Info("shutting down the server")

	// Give in-flight requests up to 5 seconds to finish before
	// forcefully cutting them off. context.WithTimeout creates a context
	// that automatically "expires" after the given duration.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // always release the context's resources once we're done with it

	// Gracefully stop the server: stop accepting new connections,
	// and wait (up to the 5s timeout) for active requests to complete.
	err := server.Shutdown(ctx)
	if err != nil {
		slog.Error("failed to shut down server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
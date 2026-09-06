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
	// load config
	cfg := config.MustLoad()

	// database setup
	// setup router

	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to student api"))
	})

	// setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	log.Printf("starting server on %s (env: %s)", cfg.Addr, cfg.Env)

	done := make(chan os.Signal,1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT,syscall.SIGALRM)


	//grace full shut down
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("failed to start server: ", err)
		}

	}()

	<- done

	slog.Info("Shutting down the server")
	
    ctx,cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	err:=server.Shutdown(ctx)
	if err != nil{
		slog.Error("failed to shut down server",slog.String("error",err.Error()))
	}

	slog.Info("server shutdown successfully")
}

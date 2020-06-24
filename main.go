package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	//"github.com/go-chi/chi" // does no have go.mod file, therefore marked as "incompatible" in our go.mod
	"github.com/mkrill/gosea/api"
	"github.com/mkrill/gosea/seabackend"
	"github.com/mkrill/gosea/status"
)

var Version = "latest"
var GoseaLogger *log.Logger

func main() {
	var err error

	ctx, cancel := context.WithCancel(context.Background())

	// create logfile and init logger
	logfile, err := os.Create("messages.log")
	if err != nil {
		log.Fatalf("error opening log file: %s", err.Error())
	}
	defer func() {
		log.Print("closing logfile")
		err := logfile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	GoseaLogger = log.New(os.Stdout, "gosea ", log.LstdFlags)

	// Create channel for os events and receive SIGTERM events
	sigChan := make(chan os.Signal)
	go func() {
		sig := <-sigChan
		log.Printf("received signal %s", sig.String())
		cancel()
	}()
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // channels gets signal, if application is terminated

	// create services
	postsService := seabackend.NewWithSEA(GoseaLogger)
	apiService := api.New(postsService, GoseaLogger)

	//chiRouter := chi.NewRouter()
	//chiRouter.Get("/health", status.Health)
	//chiRouter.Get("/api", apiService.Posts)

	mux := http.NewServeMux()                // similar to constructor, make() with standard data types
	mux.HandleFunc("/health", status.Health) // function should not be called, therefore no parameter
	mux.HandleFunc("/api", apiService.Posts)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		//Handler: chiRouter,
	} // For clarity directly as a pointer

	go func() {
		err := srv.ListenAndServe() // function is blocking until server is terminated
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			GoseaLogger.Fatalf("error starting server: %s", err.Error())
		}
	}()

	GoseaLogger.Printf("starting gosea %s", Version)

	<-ctx.Done()

	err = srv.Close()
	if err != nil {
		GoseaLogger.Fatal(err)
	}

	GoseaLogger.Print("stopping service")
}

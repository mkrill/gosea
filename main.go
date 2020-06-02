package main

import (
	"errors"
	"github.com/mkrill/gosea/api"
	"github.com/mkrill/gosea/posts"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	//"github.com/go-chi/chi" // does no have go.mod file, therefore marked as "incompatible" in our go.mod
	"github.com/mkrill/gosea/status"
)

// Prototype
func main() {
	var err error

	// create logfile and init logger
	logfile, err := os.Create("messages.log")
	if err != nil {
		log.Fatal("error opening log file")
	}
	defer func() {
		log.Print("closing logfile")
		logfile.Close()
	}()

	logger := log.New(os.Stdout, "gosea ", log.LstdFlags)

	// Create channel for os events and receive SIGTERM events
	sigChan := make(chan os.Signal)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // channels gets signal, if application is terminated

	// create services
	postsService := posts.NewWithSEA()
	apiService := api.New(postsService)

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
			logger.Fatalf("error starting server: %s", err.Error())
		}
	}()

	logger.Print("Started service")

	<-sigChan // code line blocks, until signal is received from channel

	srv.Close()

	logger.Print("stopping service")
}

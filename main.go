package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mkrill/gosea/status"
)

func main() {
	var err error

	logfile, err := os.Create("messages.log")
	if err != nil {
		log.Fatal("error opening log file")
	}
	defer func() {
		log.Print("closing logfile")
		logfile.Close()
	}()

	logger := log.New(os.Stdout, "gosea ", log.LstdFlags)

	sigChan := make(chan os.Signal)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // channels gets signal, if application is terminated

	mux := http.NewServeMux()                // similar to constructor, make() with standard data types
	mux.HandleFunc("/health", status.Health) // function should not be called, therefore no parameter

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
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

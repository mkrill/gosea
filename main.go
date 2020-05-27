package main

import (
	"net/http"

	"github.com/mkrill/gosea/status"
)

func main() {

	mux := http.NewServeMux() // similar to constructor, make() with standard data types

	mux.HandleFunc("/health", status.Health) // function should not be called, therefore no parameter

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	} // For clarity directly as a pointer

	srv.ListenAndServe()
}

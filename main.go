package main

import (
	"net/http"
	_ "net/http/pprof" // imported, but not used for any calls

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"flamingo.me/flamingo/v3/core/healthcheck"
	"flamingo.me/flamingo/v3/core/requestlogger"

	"github.com/mkrill/gosea/src/seabackend"
	"github.com/mkrill/gosea/src/seaswagger"
)

func main() {

	// insert additional, profiling endpoint
	// Grafische Ansicht: go tool pprof -http localhost:1111 -seconds 20 ...
	go http.ListenAndServe(":8888", nil)

	flamingo.App([]dingo.Module{
		new(healthcheck.Module),
		new(seabackend.Module),
		new(requestlogger.Module),
		new(seaswagger.Module),
	})
}

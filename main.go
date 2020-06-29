package main

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"flamingo.me/flamingo/v3/core/healthcheck"

	"github.com/mkrill/gosea/src/seabackend"
)

func main() {

	flamingo.App([]dingo.Module{
		new(healthcheck.Module),
		new(seabackend.Module),
	})
}

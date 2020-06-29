package main

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"flamingo.me/flamingo/v3/core/requestlogger"

	"github.com/mkrill/gosea/src/seabackend"
)

func main() {

	flamingo.App([]dingo.Module{
		new(seabackend.Module),
		new(requestlogger.Module),
	})
}

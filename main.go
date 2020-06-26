package main

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3"
	"github.com/mkrill/gosea/src/seaBackend"
)

func main() {

	flamingo.App([]dingo.Module{
		new(seaBackend.Module),
	})
}

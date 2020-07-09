package seaswagger

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/web"
)

// Module struct responsible for PreSales Management module configuration
type Module struct{}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	web.BindRoutes(injector, new(routes))

}

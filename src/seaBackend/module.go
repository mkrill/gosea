package seaBackend

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/web"
	"github.com/mkrill/gosea/src/seaBackend/domain/Service"

	"github.com/mkrill/gosea/src/seaBackend/infrastructure"
)

// Module struct responsible for PreSales Management module configuration
type Module struct{}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	web.BindRoutes(injector, new(routes))
	injector.Bind(new(Service.ISeaBackendService)).To(new(infrastructure.SeaBackendServiceAdapter))
}
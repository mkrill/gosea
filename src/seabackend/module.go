package seabackend

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/web"
	"github.com/mkrill/gosea/src/seabackend/application"
	"github.com/mkrill/gosea/src/seabackend/interfaces/controller"

	"github.com/mkrill/gosea/src/seabackend/domain/service"
	"github.com/mkrill/gosea/src/seabackend/infrastructure"
)

// Module struct responsible for PreSales Management module configuration
type Module struct{}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	web.BindRoutes(injector, new(routes))

	injector.Bind(new(service.SeaBackendLoader)).To(new(infrastructure.SeaBackendServiceAdapter))
	injector.Bind(new(controller.PostsWithUsersLoader)).To(new(application.PostsWithUsers))
}
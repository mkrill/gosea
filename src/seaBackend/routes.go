package seaBackend

import (
	"flamingo.me/flamingo/v3/framework/web"
	"github.com/mkrill/gosea/src/seaBackend/interfaces/controller"
)

// routes struct is defined to specify route handlers
type routes struct {
	apiController *controller.ApiController
}

// Inject method which defines all dependency injections used by routes struct
func (r *routes) Inject(api *controller.ApiController) {
	r.apiController = api
}

// Routes method which defines all routes handlers in module
func (r *routes) Routes(registry *web.RouterRegistry) {
	registry.HandleGet("seaBackend.Posts", r.apiController.ShowPostsWithUsers)
	registry.MustRoute("/api", "seaBackend.Posts")
}

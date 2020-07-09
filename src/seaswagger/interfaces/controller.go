package seaswagger

//go:generate mockery --all

import (
	"context"
	"errors"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"flamingo.me/flamingo/v3/framework/web"
)

type ApiController struct {
	responder *web.Responder
	logger    flamingo.Logger
}

func (a *ApiController) Inject(
	responder *web.Responder,
	logger flamingo.Logger,
) *ApiController {

	a.responder = responder
	a.logger = logger
	return a
}

// showPostsWithUsers returns a json response with filtered remote seabackend
func (a *ApiController) ShowPostsWithUsers(ctx context.Context, req *web.Request) web.Result {

	return a.responder.ServerError(errors.New("not implemented yet"))
}

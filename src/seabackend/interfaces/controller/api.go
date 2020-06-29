package controller

import (
	"context"

	"flamingo.me/flamingo/v3/framework/web"

	"github.com/mkrill/gosea/src/seabackend/domain/entity"
)

type (
	ApiController struct {
		postsAndUserApp PostsWithUsersLoader
		responder       *web.Responder
	}

	PostsWithUsersLoader interface {
		RetrievePostsWithUsersFromBackend(ctx context.Context, filter string) ([]entity.Post, error)
	}
)

func (a *ApiController) Inject(
	pwul PostsWithUsersLoader,
	responder *web.Responder,
) *ApiController {

	a.postsAndUserApp = pwul
	a.responder = responder

	return a
}

// showPostsWithUsers returns a json response with filtered remote seabackend
func (a *ApiController) ShowPostsWithUsers(ctx context.Context, req *web.Request) web.Result {
	var err error

	// retrieve query parameter 'filterValue' from URL
	filter := req.Request().URL.Query().Get("filter")

	responsePosts, err := a.postsAndUserApp.RetrievePostsWithUsersFromBackend(ctx, filter)
	if err != nil {
		return a.responder.ServerError(err)
	}

	return a.responder.Data(responsePosts)
}

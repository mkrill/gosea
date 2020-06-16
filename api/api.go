package api

import (
	"context"
	"encoding/json"
	"github.com/mkrill/gosea/seabackend"
	"log"
	"net/http"
)

type postsService interface {
	LoadPosts(ctx context.Context) ([]seabackend.RemotePost, error)
}

type Api struct {
	posts  postsService
	logger *log.Logger
}

func New(posts postsService, logger *log.Logger) *Api {
	return &Api{
		posts:  posts,
		logger: logger,
	}
}

// Posts returns a json response with filtered remote seabackend
func (a *Api) Posts(w http.ResponseWriter, r *http.Request) {
	var err error

	a.logger.Printf("got request %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctxValue := context.WithValue(r.Context(), "id", 1)

	remotePosts, err := a.posts.LoadPosts(ctxValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retrieve query parameter 'filterValue' from URL
	filterValue := r.URL.Query().Get("filter")

	responsePosts := make([]Post, 0)
	for _, remotePost := range remotePosts {
		// if current remotePost dies not match the filter, skip it
		if !remotePost.Contains(filterValue, seabackend.FieldsAll) {
			continue
		}

		post := Post{
			Title: remotePost.Title,
			Body:  remotePost.Body,
		}
		responsePosts = append(responsePosts, post)
	}

	w.Header().Set("content-type", "application/json")
	// encoder enc to convert our responsePosts slice to json
	enc := json.NewEncoder(w)
	err = enc.Encode(responsePosts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

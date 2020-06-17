package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mkrill/gosea/seabackend"
)

type seaBackendService interface {
	LoadPosts(ctx context.Context) ([]seabackend.RemotePost, error)
	LoadUser(ctx context.Context, id string) (seabackend.RemoteUser, error)
}

type Api struct {
	seaBackend seaBackendService
	logger     *log.Logger
}

func New(seaBackend seaBackendService, logger *log.Logger) *Api {
	return &Api{
		seaBackend: seaBackend,
		logger:     logger,
	}
}

// Posts returns a json response with filtered remote seabackend
func (a *Api) Posts(w http.ResponseWriter, r *http.Request) {
	var err error

	a.logger.Printf("got request %s %s", r.Method, r.URL.Path)

	// measure runtime
	start := time.Now()
	defer func() {
		a.logger.Printf("request took %s", time.Now().Sub(start))
	}()

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	ctxValue := context.WithValue(r.Context(), "id", 1)

	remotePosts, err := a.seaBackend.LoadPosts(ctxValue)
	if err != nil {
		a.logger.Printf("Error loading from seabackend: %w", err)
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

		// Load user information from backend
		user, err := a.seaBackend.LoadUser(ctxValue, remotePost.UserID.String())
		if err != nil {
			a.logger.Printf("could not load user %s", remotePost.UserID)
			continue
		}

		post := Post{
			Title:       remotePost.Title,
			Body:        remotePost.Body,
			Username:    user.Username,
			CompanyName: user.Company.Name,
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
